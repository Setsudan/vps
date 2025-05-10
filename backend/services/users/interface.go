package users

import (
	"context"
	"mime/multipart"

	m "launay-dot-one/models"
)

// Service handles user‐profile operations.
type Service interface {
	// ChangeAvatar uploads a new avatar and returns its public URL.
	ChangeAvatar(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error)

	// GetByID returns a public view of the specified user.
	GetByID(ctx context.Context, userID string) (*m.PublicUser, error)

	// GetCurrent returns a public view of the authenticated user.
	GetCurrent(ctx context.Context, userID string) (*m.PublicUser, error)

	// List returns all users in public form.
	List(ctx context.Context) ([]m.PublicUser, error)

	// UpdateProfile updates the user profile with the provided data.
	UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error
}
