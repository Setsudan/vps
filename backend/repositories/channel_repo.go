package repositories

import (
	"context"

	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db}
}

func (r *ChannelRepository) Create(ctx context.Context, ch *guilds.Channel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*guilds.Channel, error) {
	var ch guilds.Channel
	if err := r.db.WithContext(ctx).First(&ch, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.Channel, error) {
	var out []guilds.Channel
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("position ASC").
		Find(&out).Error
	return out, err
}

func (r *ChannelRepository) ListByCategory(ctx context.Context, categoryID string) ([]guilds.Channel, error) {
	var out []guilds.Channel
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("position ASC").
		Find(&out).Error
	return out, err
}

func (r *ChannelRepository) Update(ctx context.Context, ch *guilds.Channel) error {
	return r.db.WithContext(ctx).Save(ch).Error
}

func (r *ChannelRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&guilds.Channel{}, "id = ?", id).Error
}
