// services/groups/service.go
package groups

import (
	"context"
	"errors"
	"time"

	m "launay-dot-one/models/groups"
	"launay-dot-one/repositories"

	"github.com/google/uuid"
)

var ErrUnauthorized = errors.New("unauthorized")

type service struct {
	repo *repositories.GroupRepository
}

// NewService wires up the group service.
func NewService(repo *repositories.GroupRepository) Service {
	return &service{repo: repo}
}

func (s *service) CreateGroup(ctx context.Context, grp *m.Group, creatorID string) error {
	grp.ID = uuid.NewString()
	now := time.Now()
	grp.CreatedAt = now
	grp.UpdatedAt = now

	if err := s.repo.CreateGroup(ctx, grp); err != nil {
		return err
	}
	// creator is admin
	mem := &m.GroupMembership{
		GroupID:   grp.ID,
		UserID:    creatorID,
		Role:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.CreateMembership(ctx, mem)
}

func (s *service) GetGroup(ctx context.Context, groupID string) (*m.Group, error) {
	return s.repo.GetGroup(ctx, groupID)
}

func (s *service) ListGroups(ctx context.Context) ([]m.Group, error) {
	return s.repo.ListGroups(ctx)
}

func (s *service) UpdateGroup(ctx context.Context, groupID string, upd *m.Group, requesterID string) error {
	mem, err := s.repo.GetMembership(ctx, groupID, requesterID)
	if err != nil || mem.Role != "admin" {
		return ErrUnauthorized
	}
	grp, err := s.repo.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if upd.Name != "" {
		grp.Name = upd.Name
	}
	if upd.Description != "" {
		grp.Description = upd.Description
	}
	if upd.Tags != nil {
		grp.Tags = upd.Tags
	}
	grp.UpdatedAt = time.Now()
	return s.repo.UpdateGroup(ctx, grp)
}

func (s *service) DeleteGroup(ctx context.Context, groupID, requesterID string) error {
	mem, err := s.repo.GetMembership(ctx, groupID, requesterID)
	if err != nil || mem.Role != "admin" {
		return ErrUnauthorized
	}
	return s.repo.DeleteGroup(ctx, groupID)
}

func (s *service) AddMember(ctx context.Context, groupID, userID, role, requesterID string) error {
	mem, err := s.repo.GetMembership(ctx, groupID, requesterID)
	if err != nil || mem.Role != "admin" {
		return ErrUnauthorized
	}
	newMem := &m.GroupMembership{
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.CreateMembership(ctx, newMem)
}

func (s *service) UpdateMemberRole(ctx context.Context, groupID, userID, role, requesterID string) error {
	mem, err := s.repo.GetMembership(ctx, groupID, requesterID)
	if err != nil || mem.Role != "admin" {
		return ErrUnauthorized
	}
	target, err := s.repo.GetMembership(ctx, groupID, userID)
	if err != nil {
		return err
	}
	target.Role = role
	target.UpdatedAt = time.Now()
	return s.repo.UpdateMembership(ctx, target)
}

func (s *service) RemoveMember(ctx context.Context, groupID, userID, requesterID string) error {
	mem, err := s.repo.GetMembership(ctx, groupID, requesterID)
	if err != nil || mem.Role != "admin" {
		return ErrUnauthorized
	}
	return s.repo.DeleteMembership(ctx, groupID, userID)
}

func (s *service) ListMembers(ctx context.Context, groupID string) ([]m.GroupMembership, error) {
	return s.repo.ListMemberships(ctx, groupID)
}
