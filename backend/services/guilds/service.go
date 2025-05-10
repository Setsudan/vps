package guilds

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"launay-dot-one/models/guilds"
	"launay-dot-one/repositories"
)

var ErrUnauthorized = errors.New("unauthorized")

type service struct {
	guildRepo  *repositories.GuildRepository
	memberRepo *repositories.GuildMemberRepository
}

// NewService constructs a guild service.
func NewService(
	guildRepo *repositories.GuildRepository,
	memberRepo *repositories.GuildMemberRepository,
) Service {
	return &service{guildRepo: guildRepo, memberRepo: memberRepo}
}

func (s *service) CreateGuild(ctx context.Context, guild *guilds.Guild, ownerID string) error {
	guild.ID = uuid.NewString()
	now := time.Now()
	guild.OwnerID = ownerID
	guild.CreatedAt = now
	guild.UpdatedAt = now

	if err := s.guildRepo.Create(ctx, guild); err != nil {
		return err
	}

	// add owner as first member with no roles (roles can be assigned later)
	mem := &guilds.GuildMember{
		GuildID:   guild.ID,
		UserID:    ownerID,
		RoleIDs:   datatypes.JSON([]byte("[]")),
		JoinedAt:  now,
		UpdatedAt: now,
	}
	return s.memberRepo.Add(ctx, mem)
}

func (s *service) ListGuilds(ctx context.Context) ([]guilds.Guild, error) {
	return s.guildRepo.List(ctx)
}

func (s *service) GetGuild(ctx context.Context, guildID string) (*guilds.Guild, error) {
	return s.guildRepo.GetByID(ctx, guildID)
}

func (s *service) UpdateGuild(ctx context.Context, guildID string, update *guilds.Guild, requesterID string) error {
	// TODO: check requester permissions
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return err
	}
	if update.Name != "" {
		guild.Name = update.Name
	}
	if update.Description != "" {
		guild.Description = update.Description
	}
	guild.UpdatedAt = time.Now()
	return s.guildRepo.Update(ctx, guild)
}

func (s *service) DeleteGuild(ctx context.Context, guildID, requesterID string) error {
	// TODO: check requester permissions
	return s.guildRepo.Delete(ctx, guildID)
}

func (s *service) AddMember(ctx context.Context, guildID, userID string, roleIDs []string, requesterID string) error {
	// TODO: check requester permissions
	now := time.Now()
	b, _ := json.Marshal(roleIDs)
	mem := &guilds.GuildMember{
		GuildID:   guildID,
		UserID:    userID,
		RoleIDs:   datatypes.JSON(b),
		JoinedAt:  now,
		UpdatedAt: now,
	}
	return s.memberRepo.Add(ctx, mem)
}

func (s *service) UpdateMemberRoles(
	ctx context.Context, guildID, userID string, roleIDs []string, requesterID string,
) error {
	// TODO: check requester permissions
	mem, err := s.memberRepo.Get(ctx, guildID, userID)
	if err != nil {
		return err
	}
	b, _ := json.Marshal(roleIDs)
	mem.RoleIDs = datatypes.JSON(b)
	mem.UpdatedAt = time.Now()
	return s.memberRepo.Update(ctx, mem)
}

func (s *service) RemoveMember(ctx context.Context, guildID, userID, requesterID string) error {
	// TODO: check requester permissions
	return s.memberRepo.Remove(ctx, guildID, userID)
}

func (s *service) ListMembers(ctx context.Context, guildID string) ([]guilds.GuildMember, error) {
	return s.memberRepo.ListByGuild(ctx, guildID)
}
