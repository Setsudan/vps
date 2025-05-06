package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"launay-dot-one/controllers"
	"launay-dot-one/listeners"
	"launay-dot-one/models"
	"launay-dot-one/models/friendships"
	"launay-dot-one/models/groups"
	"launay-dot-one/models/guilds"
	"launay-dot-one/models/resume"
	"launay-dot-one/realtime"
	"launay-dot-one/repositories"

	authsvc "launay-dot-one/services/auth"
	"launay-dot-one/services/categories"
	"launay-dot-one/services/channels"
	frdsvc "launay-dot-one/services/friendships"
	groupsvc "launay-dot-one/services/groups" // legacy groups
	"launay-dot-one/services/guildroles"
	guildsvc "launay-dot-one/services/guilds" // new guilds
	msgsrv "launay-dot-one/services/messaging"
	"launay-dot-one/services/permissions"
	resumeSvc "launay-dot-one/services/resumes"
	usersvc "launay-dot-one/services/users"

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

func InitServer() (*http.Server, error) {
	_ = godotenv.Load() // ignore missing .env
	jwtSecret := utils.MustEnv("JWT_SECRET")
	logger := utils.GetLogger()

	// ─── Redis
	rdb, err := initRedis(logger)
	if err != nil {
		return nil, err
	}

	// ─── Storage
	minioClient, err := initMinioClient()
	if err != nil {
		return nil, err
	}
	storageService, err := initStorageService(minioClient)
	if err != nil {
		return nil, err
	}

	// ─── Database & Migrations
	db, err := initDatabaseWithDefaults()
	if err != nil {
		logger.Fatal("Database init failed: ", err)
		return nil, err
	}

	// ─── Repositories
	userRepo := repositories.NewUserRepository(db)
	groupRepo := repositories.NewGroupRepository(db) // legacy groups
	messagingRepo := repositories.NewMessagingRepository(db)
	friendRepo := repositories.NewFriendRequestRepository(db)
	resumeRepo := repositories.NewResumeRepository(db)
	guildRepo := repositories.NewGuildRepository(db)
	guildMemberRepo := repositories.NewGuildMemberRepository(db)
	permRepo := repositories.NewPermissionOverwriteRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	guildRoleRepo := repositories.NewGuildRoleRepository(db)

	// ─── Services
	authService := authsvc.NewService(userRepo, jwtSecret, 72*time.Hour)
	userService := usersvc.NewService(storageService, userRepo)
	groupService := groupsvc.NewService(groupRepo)
	friendService := frdsvc.NewService(friendRepo, db)
	messagingService := msgsrv.NewService(rdb, messagingRepo)
	resumeService := resumeSvc.NewService(resumeRepo)
	guildService := guildsvc.NewService(guildRepo, guildMemberRepo)
	presenceService := realtime.NewPresenceService(rdb)
	permService := permissions.NewService(permRepo)
	categoryService := categories.NewService(categoryRepo)
	channelService := channels.NewService(channelRepo)
	guildRoleService := guildroles.NewService(guildRoleRepo)

	// ─── Controllers
	authController := controllers.NewAuthController(authService, logger)
	userController := controllers.NewUserController(logger, userService)
	groupController := controllers.NewGroupController(groupService, logger)
	messagingController := controllers.NewMessagingController(messagingService, groupService, logger)
	presenceController := controllers.NewPresenceController(presenceService, rdb, logger)
	resumeController := controllers.NewResumeController(resumeService, logger)
	friendshipController := controllers.NewFriendshipController(friendService, logger)
	guildController := controllers.NewGuildController(guildService, logger)
	permissionsController := controllers.NewPermissionsController(permService, logger)
	categoryController := controllers.NewCategoriesController(categoryService, logger)
	channelController := controllers.NewChannelsController(channelService, logger)
	guildRolesController := controllers.NewGuildRolesController(guildRoleService, logger)

	// ─── Redis‐TTL cleanup & expired listener
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := messagingService.TransferExpiredMessages(context.Background()); err != nil {
				logger.Error("TransferExpiredMessages error:", err)
			}
		}
	}()
	if err := listeners.RedisExpiredListener(context.Background(), rdb, messagingService); err != nil {
		logger.Errorf("RedisExpiredListener error: %v", err)
	}

	// ─── Router & CORS
	router := SetupRouter(
		authController,
		presenceController,
		userController,
		messagingController,
		groupController,
		resumeController,
		friendshipController,
		guildController,
		permissionsController,
		categoryController,
		channelController,
		guildRolesController,
	)
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{utils.GetEnv("CORS_ORIGIN", "http://localhost:1420")}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
	)(router)

	srv := &http.Server{
		Addr:         ":" + utils.GetEnv("APP_PORT", "8080"),
		Handler:      cors,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	logger.Info("Server initialized on port ", utils.GetEnv("APP_PORT", "8080"))
	return srv, nil
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

func initMinioClient() (*minio.Client, error) {
	rawURL := utils.MustEnv("STORAGE_ENDPOINT") // e.g. http://minio:9000
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid STORAGE_ENDPOINT %q: %w", rawURL, err)
	}

	var client *minio.Client
	retries := 3
	for i := 0; i < retries; i++ {
		client, err = minio.New(u.Host, &minio.Options{
			Creds: credentials.NewStaticV4(
				os.Getenv("MINIO_ROOT_USER"),
				os.Getenv("MINIO_ROOT_PASSWORD"),
				""),
			Secure: u.Scheme == "https",
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second) // wait before retrying
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client after %d retries: %w", retries, err)
	}
	return client, nil
}

func initStorageService(client *minio.Client) (*storage.StorageService, error) {
	bucket := utils.GetEnv("STORAGE_BUCKET", "avatars")
	svc := storage.NewStorageService(client, bucket)
	return svc, nil
}

func initDatabaseWithDefaults() (*gorm.DB, error) {
	host := utils.MustEnv("DB_HOST")
	user := utils.MustEnv("DB_USER")
	pass := utils.MustEnv("DB_PASSWORD")
	dbName := utils.MustEnv("DB_NAME")
	sslmode := utils.GetEnv("DB_SSLMODE", "disable")
	tz := utils.GetEnv("DB_TIMEZONE", "UTC")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, user, pass, dbName, sslmode, tz,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	if err := db.AutoMigrate(
		// core
		&models.User{},
		&models.Message{},

		// legacy group feature
		&groups.Group{},
		&groups.GroupMembership{},

		// friendship system
		&friendships.FriendRequest{},

		// Discord‐style guilds
		&guilds.Guild{},
		&guilds.GuildRole{},
		&guilds.GuildMember{},
		&guilds.Category{},
		&guilds.Channel{},
		&guilds.PermissionOverwrite{},

		// resumes
		&models.Resume{},
		&resume.Education{},
		&resume.Experience{},
		&resume.Project{},
		&resume.Certification{},
		&resume.Skill{},
		&resume.Interest{},
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
