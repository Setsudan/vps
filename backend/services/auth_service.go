package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"launay-dot-one/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	RegisterUser(ctx context.Context, user *models.User) error
	LoginUser(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthService(db *gorm.DB, jwtSecret string) AuthService {
	return &authService{db: db, jwtSecret: jwtSecret}
}

func (s *authService) RegisterUser(ctx context.Context, user *models.User) error {
	var existing models.User
	if err := s.db.WithContext(ctx).Where("email = ?", user.Email).First(&existing).Error; err == nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.ID = uuid.NewString()
	if user.Username == "" {
		atIndex := strings.Index(user.Email, "@")
		if atIndex != -1 {
			user.Username = user.Email[:atIndex]
		}
	}

	return s.db.WithContext(ctx).Create(user).Error
}

func (s *authService) LoginUser(ctx context.Context, email, password string) (string, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}
