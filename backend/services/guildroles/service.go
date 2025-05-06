package guildroles

import (
	"context"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

type service struct {
	repo *repositories.GuildRoleRepository
}

func NewService(repo *repositories.GuildRoleRepository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, role *guilds.GuildRole) error {
	return s.repo.Create(ctx, role)
}

func (s *service) Get(ctx context.Context, id string) (*guilds.GuildRole, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) List(ctx context.Context, guildID string) ([]guilds.GuildRole, error) {
	return s.repo.ListByGuild(ctx, guildID)
}

func (s *service) Update(ctx context.Context, role *guilds.GuildRole) error {
	return s.repo.Update(ctx, role)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
