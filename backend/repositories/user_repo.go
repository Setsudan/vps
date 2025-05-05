package repositories

import (
	"context"

	"launay-dot-one/models"

	"gorm.io/gorm"
)

// UserRepository handles CRUD operations on User models.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new User into the database.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a User by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a User by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateAvatar updates only the Avatar field of the specified user.
func (r *UserRepository) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("avatar", avatarURL).
		Error
}

// ListAll returns all User records.
func (r *UserRepository) ListAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Update updates a User's fields based on the provided userID and updates map.
func (r *UserRepository) UpdateFields(ctx context.Context, userID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).
		Error
}
