package guilds

import (
	"context"

	mg "launay-dot-one/models/guilds"
)

// Service defines guild‐related business logic.
type Service interface {
	// Guild CRUD
	CreateGuild(ctx context.Context, guild *mg.Guild, ownerID string) error
	ListGuilds(ctx context.Context) ([]mg.Guild, error)
	GetGuild(ctx context.Context, guildID string) (*mg.Guild, error)
	UpdateGuild(ctx context.Context, guildID string, update *mg.Guild, requesterID string) error
	DeleteGuild(ctx context.Context, guildID string, requesterID string) error

	// Membership management
	AddMember(ctx context.Context, guildID, userID string, roleIDs []string, requesterID string) error
	UpdateMemberRoles(ctx context.Context, guildID, userID string, roleIDs []string, requesterID string) error
	RemoveMember(ctx context.Context, guildID, userID, requesterID string) error
	ListMembers(ctx context.Context, guildID string) ([]mg.GuildMember, error)
	ListGuildsForUser(ctx context.Context, userID string) ([]mg.Guild, error)
}
