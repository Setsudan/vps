package repositories

import (
	"context"

	"gorm.io/gorm"

	"launay-dot-one/models"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Save persists a single channel message.
func (r *MessageRepository) Save(ctx context.Context, m *models.Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// ListByChannel returns all messages for a channel in ascending time order.
func (r *MessageRepository) ListByChannel(
	ctx context.Context, channelID string,
) ([]models.Message, error) {
	var out []models.Message
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at ASC").
		Find(&out).Error
	return out, err
}
