package categories

import (
	"context"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

type service struct {
	repo *repositories.CategoryRepository
}

func NewService(repo *repositories.CategoryRepository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, c *guilds.Category) error {
	return s.repo.Create(ctx, c)
}

func (s *service) Get(ctx context.Context, id string) (*guilds.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, guildID string) ([]guilds.Category, error) {
	return s.repo.ListByGuild(ctx, guildID)
}

func (s *service) Update(ctx context.Context, c *guilds.Category) error {
	return s.repo.Update(ctx, c)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
