package guildroles

import (
	"context"

	"launay-dot-one/models/guilds"
)

type Service interface {
	Create(ctx context.Context, role *guilds.GuildRole) error
	Get(ctx context.Context, id string) (*guilds.GuildRole, error)
	List(ctx context.Context, guildID string) ([]guilds.GuildRole, error)
	Update(ctx context.Context, role *guilds.GuildRole) error
	Delete(ctx context.Context, id string) error
}
