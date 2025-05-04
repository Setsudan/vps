package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"launay-dot-one/controllers"
	"launay-dot-one/listeners"
	"launay-dot-one/models"
	"launay-dot-one/realtime"
	"launay-dot-one/repositories"
	"launay-dot-one/services"
	"launay-dot-one/storage"
	"launay-dot-one/utils"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitServer initializes dependencies and returns an HTTP server.
func InitServer() (*http.Server, error) {
	if err := loadEnv(); err != nil {
		log.Println(err)
	}

	var jwtSecret = []byte(utils.MustEnv("JWT_SECRET"))
	logger := utils.GetLogger()

	rdb, err := initRedis(logger)
	if err != nil {
		return nil, err
	}

	s3Client, err := initS3Client()
	if err != nil {
		return nil, err
	}

	storageService, err := initStorageService(s3Client)
	if err != nil {
		return nil, err
	}

	db, err := initDatabaseWithDefaults()
	if err != nil {
		logger.Fatal("Database initialization with defaults failed: ", err)
		return nil, err
	}

	// Initialize repositories.
	groupRepo := repositories.NewGroupRepository(db)
	messagingRepo := repositories.NewMessagingRepository(db)

	// Initialize services.
	presenceService := realtime.NewPresenceService(rdb)
	authService := services.NewAuthService(db, string(jwtSecret))
	userService := services.NewUserService(storageService, db)
	groupService := services.NewGroupService(groupRepo)
	messagingService := services.NewMessagingService(rdb, messagingRepo)
	locationService := services.NewLocationService(rdb)

	// Initialize controllers.
	presenceController := controllers.NewPresenceController(presenceService, rdb, logger)
	authController := controllers.NewAuthController(authService, logger)
	userController := controllers.NewUserController(logger, userService, string(jwtSecret))
	messagingController := controllers.NewMessagingController(messagingService, groupService, logger)
	groupController := controllers.NewGroupController(groupService, logger)
	locationController := controllers.NewLocationController(locationService, logger)

	if err := listeners.RedisExpiredListener(context.Background(), rdb, messagingService); err != nil {
		log.Printf("Error starting Redis expired listener: %v", err)
	}

	// Setup Router.
	router := SetupRouter(
		authController,
		presenceController,
		userController,
		messagingController,
		groupController,
		locationController,
	)

	// Configure CORS.
	corsRouter := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:1420"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
	)(router)

	server := &http.Server{
		Handler:      corsRouter,
		Addr:         ":" + getEnv("APP_PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Server initialization completed")
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			err := messagingService.TransferExpiredMessages(context.Background())
			if err != nil {
				logger.Error("Error transferring expired messages: ", err)
			}
		}
	}()
	logger.Info("Initialized go routine for transferring expired messages")
	return server, nil
}

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("no .env file found, using environment variables")
	}
	return nil
}

func initRedis(logger *logrus.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDRESS", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Could not connect to Redis: ", err)
		return nil, err
	}

	logger.Info("Connected to Redis")
	return rdb, nil
}

func initS3Client() (*minio.Client, error) {
	rawURL := getEnv("SEAWEEDFS_URL", "http://localhost:8333")

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid SEAWEEDFS_URL %q: %w", rawURL, err)
	}

	endpoint := u.Host            // host[:port]
	secure := u.Scheme == "https" // true ↔ HTTPS

	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	const maxRetries = 5
	var lastErr error

	for i := 1; i <= maxRetries; i++ {
		clt, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: secure,
		})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err = clt.ListBuckets(ctx); err == nil {
				return clt, nil // success
			}
		}

		lastErr = err
		log.Printf("Warning: S3 init failed (%d/%d): %v", i, maxRetries, err)
		time.Sleep(time.Duration(1<<i) * time.Second) // 2 s, 4 s, 8 s…
	}
	return nil, fmt.Errorf("failed to initialise S3 client: %w", lastErr)
}

func initStorageService(client *minio.Client) (*storage.StorageService, error) {
	bucket := "avatars"
	var storageService *storage.StorageService

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			exists, err := client.BucketExists(ctx, bucket)
			if err != nil {
				log.Printf("Warning: error checking bucket existence: %v. Retrying...", err)
			} else if !exists {
				if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
					log.Printf("Error creating bucket: %v. Retrying...", err)
				} else {
					log.Printf("Bucket '%s' created successfully", bucket)
					break
				}
			} else {
				log.Printf("Bucket '%s' already exists", bucket)
				break
			}

			time.Sleep(10 * time.Second) // Retry after 10 seconds
		}
	}()

	storageService = storage.NewStorageService(client, bucket)
	return storageService, nil
}

func initDatabaseWithDefaults() (*gorm.DB, error) {
	host := utils.MustEnv("DB_HOST")
	user := utils.MustEnv("DB_USER")
	pass := utils.MustEnv("DB_PASSWORD")
	dbName := utils.MustEnv("DB_NAME")
	sslmode := utils.GetEnv("DB_SSLMODE", "disable") // override later if you add TLS
	tz := utils.GetEnv("DB_TIMEZONE", "UTC")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, user, pass, dbName, sslmode, tz,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMembership{},
		&models.Message{},
	); err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
