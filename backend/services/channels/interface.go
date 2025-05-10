package channels

import (
	"context"

	"launay-dot-one/models/guilds"
)

type Service interface {
	Create(ctx context.Context, ch *guilds.Channel, categoryID *string) error
	Get(ctx context.Context, id string) (*guilds.Channel, error)
	ListByGuild(ctx context.Context, guildID string) ([]guilds.Channel, error)
	ListByCategory(ctx context.Context, categoryID string) ([]guilds.Channel, error)
	Update(ctx context.Context, ch *guilds.Channel) error
	Delete(ctx context.Context, id string) error
}
