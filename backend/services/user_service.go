package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"launay-dot-one/models"
	"launay-dot-one/storage"

	"gorm.io/gorm"
)

type UserService struct {
	storageService *storage.StorageService
	db             *gorm.DB
}

func NewUserService(storageService *storage.StorageService, db *gorm.DB) *UserService {
	return &UserService{
		storageService: storageService,
		db:             db,
	}
}

func (us *UserService) GetDB() *gorm.DB {
	return us.db
}

func (us *UserService) ChangeAvatar(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	avatarURL, err := us.storageService.UploadFile(ctx, file, header, userID)
	if err != nil {
		return "", err
	}

	if err := us.updateUserAvatar(ctx, userID, avatarURL); err != nil {
		return "", err
	}

	return avatarURL, nil
}

func (us *UserService) updateUserAvatar(ctx context.Context, userID, avatarURL string) error {
	var user models.User
	if err := us.db.WithContext(ctx).
		First(&user, "id = ?", userID).
		Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// rewrite the URL
	user.Avatar = strings.Replace(
		avatarURL,
		"http://seaweedfs:8333/",
		"https://api.launay.one/storage/",
		1,
	)

	if err := us.db.WithContext(ctx).
		Save(&user).
		Error; err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}
	return nil
}

func (us *UserService) GetUserByID(ctx context.Context, userID string) (*models.PublicUser, error) {
	var user models.User
	if err := us.db.WithContext(ctx).
		First(&user, "id = ?", userID).
		Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	publicUser := user.ToPublic()
	return &publicUser, nil
}

func (us *UserService) GetCurrentAuthenticatedUser(ctx context.Context, userID string) (*models.PublicUser, error) {
	return us.GetUserByID(ctx, userID)
}

func (us *UserService) GetAllUsers(ctx context.Context) ([]models.PublicUser, error) {
	var users []models.User
	if err := us.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	publicUsers := make([]models.PublicUser, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublic()
	}

	return publicUsers, nil
}
