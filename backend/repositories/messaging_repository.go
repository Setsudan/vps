package repositories

import (
	"context"

	"launay-dot-one/models"

	"gorm.io/gorm"
)

type MessagingRepository struct {
	db *gorm.DB
}

func NewMessagingRepository(db *gorm.DB) *MessagingRepository {
	return &MessagingRepository{db: db}
}

func (r *MessagingRepository) SaveMessage(ctx context.Context, msg models.Message) error {
	return r.db.WithContext(ctx).Create(&msg).Error
}

// GetMessages retrieves messages for a given target (user, group, or channel)
// ordered by timestamp.
func (r *MessagingRepository) GetMessages(ctx context.Context, targetID, targetType string) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND target_type = ?", targetID, targetType).
		Order("timestamp ASC").
		Find(&messages).Error
	return messages, err
}

func (r *MessagingRepository) GetMessagesBetweenUsers(ctx context.Context, userA, userB string) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where(
			r.db.
				Where("sender_id = ? AND target_id = ?", userA, userB).
				Or("sender_id = ? AND target_id = ?", userB, userA),
		).
		Order("timestamp ASC").
		Find(&messages).Error
	return messages, err
}

func (r *MessagingRepository) GetAllMessagesForUser(ctx context.Context, userID string) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where("sender_id = ? OR target_id = ?", userID, userID).
		Order("timestamp ASC").
		Find(&messages).Error

	return messages, err
}
