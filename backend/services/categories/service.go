package categories

import (
	"context"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

type service struct {
	repo        *repositories.CategoryRepository
	channelRepo *repositories.ChannelRepository
}

func NewService(
	repo *repositories.CategoryRepository,
	channelRepo *repositories.ChannelRepository,
) Service {
	return &service{repo: repo, channelRepo: channelRepo}
}

func (s *service) Create(
	ctx context.Context,
	c *guilds.Category,
	channels []*guilds.Channel,
) error {
	if err := s.repo.Create(ctx, c); err != nil {
		return err
	}
	for _, ch := range channels {
		ch.GuildID = c.GuildID
		ch.CategoryID = &c.ID
		if err := s.channelRepo.Create(ctx, ch); err != nil {
			return err
		}
	}
	return nil
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
