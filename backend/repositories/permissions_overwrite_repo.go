package repositories

import (
	"context"

	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type PermissionOverwriteRepository struct {
	db *gorm.DB
}

func NewPermissionOverwriteRepository(db *gorm.DB) *PermissionOverwriteRepository {
	return &PermissionOverwriteRepository{db}
}

func (r *PermissionOverwriteRepository) Create(ctx context.Context, o *guilds.PermissionOverwrite) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *PermissionOverwriteRepository) Update(ctx context.Context, o *guilds.PermissionOverwrite) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *PermissionOverwriteRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&guilds.PermissionOverwrite{}, "id = ?", id).Error
}

// List all or filter by guild, category, channel
func (r *PermissionOverwriteRepository) List(
	ctx context.Context,
	guildID string,
	categoryID, channelID *string,
) ([]guilds.PermissionOverwrite, error) {
	q := r.db.WithContext(ctx).Where("guild_id = ?", guildID)
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if channelID != nil {
		q = q.Where("channel_id = ?", *channelID)
	}
	var out []guilds.PermissionOverwrite
	return out, q.Find(&out).Error
}
