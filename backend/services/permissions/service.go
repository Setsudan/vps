package permissions

import (
	"context"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

type service struct {
	repo *repositories.PermissionOverwriteRepository
}

func NewService(repo *repositories.PermissionOverwriteRepository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, o *guilds.PermissionOverwrite) error {
	return s.repo.Create(ctx, o)
}

func (s *service) Update(ctx context.Context, o *guilds.PermissionOverwrite) error {
	return s.repo.Update(ctx, o)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) List(
	ctx context.Context,
	guildID string,
	categoryID, channelID *string,
) ([]guilds.PermissionOverwrite, error) {
	return s.repo.List(ctx, guildID, categoryID, channelID)
}
