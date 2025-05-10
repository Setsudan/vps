package users

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	m "launay-dot-one/models"
	"launay-dot-one/repositories"
	"launay-dot-one/storage"
)

type service struct {
	storageSvc *storage.StorageService
	userRepo   *repositories.UserRepository
}

// NewService constructs the user service.
func NewService(
	storageSvc *storage.StorageService,
	userRepo *repositories.UserRepository,
) Service {
	return &service{storageSvc, userRepo}
}

func (s *service) ChangeAvatar(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	url, err := s.storageSvc.UploadFile(ctx, file, header, userID)
	if err != nil {
		return "", err
	}
	// rewrite filesystem URL → public URL
	publicURL := strings.Replace(url, "http://seaweedfs:8333/", "https://api.launay.one/storage/", 1)
	if err := s.userRepo.UpdateAvatar(ctx, userID, publicURL); err != nil {
		return "", fmt.Errorf("update avatar: %w", err)
	}
	return publicURL, nil
}

func (s *service) GetByID(ctx context.Context, userID string) (*m.PublicUser, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	pu := u.ToPublic()
	return &pu, nil
}

func (s *service) GetCurrent(ctx context.Context, userID string) (*m.PublicUser, error) {
	return s.GetByID(ctx, userID)
}

func (s *service) List(ctx context.Context) ([]m.PublicUser, error) {
	users, err := s.userRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]m.PublicUser, len(users))
	for i, u := range users {
		out[i] = u.ToPublic()
	}
	return out, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	if err := s.userRepo.UpdateFields(ctx, userID, updates); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}
