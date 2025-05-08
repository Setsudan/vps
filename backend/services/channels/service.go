package channels

import (
	"context"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

type service struct {
	repo *repositories.ChannelRepository
}

func NewService(repo *repositories.ChannelRepository) Service {
	return &service{repo}
}

func (s *service) Create(
	ctx context.Context,
	ch *guilds.Channel,
	categoryID *string,
) error {
	ch.CategoryID = categoryID
	return s.repo.Create(ctx, ch)
}

func (s *service) Get(ctx context.Context, id string) (*guilds.Channel, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListByGuild(ctx context.Context, guildID string) ([]guilds.Channel, error) {
	return s.repo.ListByGuild(ctx, guildID)
}

func (s *service) ListByCategory(ctx context.Context, categoryID string) ([]guilds.Channel, error) {
	return s.repo.ListByCategory(ctx, categoryID)
}

func (s *service) Update(ctx context.Context, ch *guilds.Channel) error {
	return s.repo.Update(ctx, ch)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
