package auth

import (
	"context"

	m "launay-dot-one/models"
)

// Service handles registration and login.
type Service interface {
	// RegisterUser hashes password, creates a User record.
	RegisterUser(ctx context.Context, user *m.User) error

	// LoginUser validates credentials and returns a JWT.
	LoginUser(ctx context.Context, email, password string) (string, error)
}
