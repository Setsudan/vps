package repositories

import (
	"context"
	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type GuildMemberRepository struct{ db *gorm.DB }

func NewGuildMemberRepository(db *gorm.DB) *GuildMemberRepository {
	return &GuildMemberRepository{db}
}

func (r *GuildMemberRepository) Add(ctx context.Context, m *guilds.GuildMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GuildMemberRepository) Update(ctx context.Context, m *guilds.GuildMember) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *GuildMemberRepository) Remove(ctx context.Context, guildID, userID string) error {
	return r.db.WithContext(ctx).
		Delete(guilds.GuildMember{}, "guild_id = ? AND user_id = ?", guildID, userID).
		Error
}

func (r *GuildMemberRepository) Get(ctx context.Context, guildID, userID string) (*guilds.GuildMember, error) {
	var m guilds.GuildMember
	if err := r.db.WithContext(ctx).
		First(&m, "guild_id = ? AND user_id = ?", guildID, userID).
		Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *GuildMemberRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.GuildMember, error) {
	var ms []guilds.GuildMember
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Find(&ms).Error
	return ms, err
}
