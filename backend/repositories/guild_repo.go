package repositories

import (
	"context"
	"gorm.io/gorm"
	"launay-dot-one/models/guilds"
)

// GuildRepository manages Guild records.
type GuildRepository struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) *GuildRepository {
	return &GuildRepository{db}
}

func (r *GuildRepository) Create(ctx context.Context, g *guilds.Guild) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GuildRepository) Update(ctx context.Context, g *guilds.Guild) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *GuildRepository) Delete(ctx context.Context, guildID string) error {
	return r.db.WithContext(ctx).
		Delete(&guilds.Guild{}, "id = ?", guildID).Error
}

func (r *GuildRepository) GetByID(ctx context.Context, guildID string) (*guilds.Guild, error) {
	var g guilds.Guild
	if err := r.db.WithContext(ctx).
		First(&g, "id = ?", guildID).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GuildRepository) List(ctx context.Context) ([]guilds.Guild, error) {
	var list []guilds.Guild
	return list, r.db.WithContext(ctx).Find(&list).Error
}

func (r *GuildRepository) ListByUser(ctx context.Context, userID string) ([]guilds.Guild, error) {
	var out []guilds.Guild
	err := r.db.WithContext(ctx).
		Table("guilds").
		Select("guilds.*").
		Joins("JOIN guild_members ON guilds.id = guild_members.guild_id").
		Where("guild_members.user_id = ?", userID).
		Find(&out).Error
	return out, err
}
