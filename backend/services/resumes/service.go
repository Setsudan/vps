package resumes

import (
	"context"
	"fmt"

	"launay-dot-one/models"
	"launay-dot-one/repositories"
)

type service struct {
	repo *repositories.ResumeRepository
}

// NewService constructs the resume service.
func NewService(repo *repositories.ResumeRepository) Service {
	return &service{repo: repo}
}

func (s *service) GetByUser(ctx context.Context, userID string) (*models.Resume, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *service) Create(ctx context.Context, res *models.Resume) error {
	return s.repo.Create(ctx, res)
}

func (s *service) Update(ctx context.Context, res *models.Resume) error {
	// Ensure the resume exists and belongs to this user
	existing, err := s.repo.GetByUser(ctx, res.UserID)
	if err != nil {
		return fmt.Errorf("resume not found: %w", err)
	}
	res.ID = existing.ID
	return s.repo.Update(ctx, res)
}

func (s *service) DeleteByUser(ctx context.Context, userID string) error {
	// Fetch to get ID, then delete
	existing, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("resume not found: %w", err)
	}
	return s.repo.Delete(ctx, existing.ID)
}
