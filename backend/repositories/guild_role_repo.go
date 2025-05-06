package repositories

import (
	"context"

	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type GuildRoleRepository struct {
	db *gorm.DB
}

func NewGuildRoleRepository(db *gorm.DB) *GuildRoleRepository {
	return &GuildRoleRepository{db}
}

func (r *GuildRoleRepository) Create(ctx context.Context, role *guilds.GuildRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *GuildRoleRepository) Get(ctx context.Context, id string) (*guilds.GuildRole, error) {
	var role guilds.GuildRole
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *GuildRoleRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.GuildRole, error) {
	var roles []guilds.GuildRole
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("position ASC").
		Find(&roles).Error
	return roles, err
}

func (r *GuildRoleRepository) Update(ctx context.Context, role *guilds.GuildRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *GuildRoleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&guilds.GuildRole{}, "id = ?", id).Error
}
