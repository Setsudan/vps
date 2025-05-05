package repositories

import (
	"context"
	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type PermissionOverwriteRepository struct{ db *gorm.DB }

func NewPermissionOverwriteRepository(db *gorm.DB) *PermissionOverwriteRepository {
	return &PermissionOverwriteRepository{db}
}

func (r *PermissionOverwriteRepository) Create(ctx context.Context, ow *guilds.PermissionOverwrite) error {
	return r.db.WithContext(ctx).Create(ow).Error
}

func (r *PermissionOverwriteRepository) Update(ctx context.Context, ow *guilds.PermissionOverwrite) error {
	return r.db.WithContext(ctx).Save(ow).Error
}

func (r *PermissionOverwriteRepository) Delete(ctx context.Context, owID string) error {
	return r.db.WithContext(ctx).
		Delete(&guilds.PermissionOverwrite{}, "id = ?", owID).
		Error
}

// List for a category or channel (one of categoryID/channelID must be non-nil)
func (r *PermissionOverwriteRepository) List(ctx context.Context, guildID string, categoryID, channelID *string) ([]guilds.PermissionOverwrite, error) {
	var list []guilds.PermissionOverwrite
	q := r.db.WithContext(ctx).Where("guild_id = ?", guildID)
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if channelID != nil {
		q = q.Where("channel_id = ?", *channelID)
	}
	err := q.Order("created_at ASC").Find(&list).Error
	return list, err
}
