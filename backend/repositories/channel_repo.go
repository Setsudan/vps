package repositories

import (
	"context"
	"gorm.io/gorm"
	"launay-dot-one/models/guilds"
)

type ChannelRepository struct{ db *gorm.DB }

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db}
}

func (r *ChannelRepository) Create(ctx context.Context, ch *guilds.Channel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *ChannelRepository) Update(ctx context.Context, ch *guilds.Channel) error {
	return r.db.WithContext(ctx).Save(ch).Error
}

func (r *ChannelRepository) Delete(ctx context.Context, channelID string) error {
	return r.db.WithContext(ctx).
		Delete(&guilds.Channel{}, "id = ?", channelID).
		Error
}

func (r *ChannelRepository) ListByCategory(ctx context.Context, categoryID string) ([]guilds.Channel, error) {
	var chs []guilds.Channel
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("position ASC").
		Find(&chs).Error
	return chs, err
}

func (r *ChannelRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.Channel, error) {
	var chs []guilds.Channel
	err := r.db.WithContext(ctx).
		Where("guild_id = ? AND category_id IS NULL", guildID).
		Find(&chs).Error
	return chs, err
}
