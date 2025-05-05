package resumes

import (
	"context"

	m "launay-dot-one/models"
)

// Service defines resume‐related business logic.
type Service interface {
	// GetByUser returns the resume (with all sub‐records) for the given user.
	GetByUser(ctx context.Context, userID string) (*m.Resume, error)

	// Create inserts a new resume record.
	Create(ctx context.Context, res *m.Resume) error

	// Update modifies an existing resume.
	Update(ctx context.Context, res *m.Resume) error

	// DeleteByUser deletes the resume belonging to the given user.
	DeleteByUser(ctx context.Context, userID string) error
}
