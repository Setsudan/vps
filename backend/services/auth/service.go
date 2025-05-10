package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"launay-dot-one/models"
	"launay-dot-one/repositories"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type service struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
	ttl       time.Duration
}

// NewService constructs the auth service.
func NewService(
	userRepo *repositories.UserRepository,
	jwtSecret string,
	ttl time.Duration,
) Service {
	return &service{userRepo, jwtSecret, ttl}
}

func (s *service) RegisterUser(ctx context.Context, user *models.User) error {
	if _, err := s.userRepo.GetByEmail(ctx, user.Email); err == nil {
		return errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.ID = uuid.NewString()
	user.Password = string(hash)
	user.Role = "user"

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *service) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
