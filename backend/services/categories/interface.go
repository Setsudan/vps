package categories

import (
	"context"

	"launay-dot-one/models/guilds"
)

type Service interface {
	Create(ctx context.Context, c *guilds.Category, channels []*guilds.Channel) error
	Get(ctx context.Context, id string) (*guilds.Category, error)
	List(ctx context.Context, guildID string) ([]guilds.Category, error)
	Update(ctx context.Context, c *guilds.Category) error
	Delete(ctx context.Context, id string) error
}
