package repositories

import (
	"context"
	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

// GuildRoleRepository manages roles within a guild.
type GuildRoleRepository struct{ db *gorm.DB }

func NewGuildRoleRepository(db *gorm.DB) *GuildRoleRepository {
	return &GuildRoleRepository{db}
}

func (r *GuildRoleRepository) Create(ctx context.Context, role *guilds.GuildRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *GuildRoleRepository) Update(ctx context.Context, role *guilds.GuildRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *GuildRoleRepository) Delete(ctx context.Context, roleID string) error {
	return r.db.WithContext(ctx).
		Delete(&guilds.GuildRole{}, "id = ?", roleID).Error
}

func (r *GuildRoleRepository) Get(ctx context.Context, roleID string) (*guilds.GuildRole, error) {
	var role guilds.GuildRole
	if err := r.db.WithContext(ctx).First(&role, "id = ?", roleID).Error; err != nil {
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
