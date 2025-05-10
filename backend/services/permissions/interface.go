package permissions

import (
	"context"

	m "launay-dot-one/models/guilds"
)

type Service interface {
	Create(ctx context.Context, o *m.PermissionOverwrite) error
	Update(ctx context.Context, o *m.PermissionOverwrite) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, guildID string, categoryID, channelID *string) ([]m.PermissionOverwrite, error)
}
