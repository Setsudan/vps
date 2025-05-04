package services

import (
	"context"
	"errors"
	"time"

	"launay-dot-one/models"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

// GroupDB defines the database operations required for group management.
type GroupDB interface {
	// Group operations.
	CreateGroup(ctx context.Context, group *models.Group) error
	UpdateGroup(ctx context.Context, group *models.Group) error
	DeleteGroup(ctx context.Context, groupID string) error
	GetGroup(ctx context.Context, groupID string) (*models.Group, error)
	ListGroups(ctx context.Context) ([]models.Group, error)

	// Membership operations.
	CreateMembership(ctx context.Context, membership *models.GroupMembership) error
	UpdateMembership(ctx context.Context, membership *models.GroupMembership) error
	DeleteMembership(ctx context.Context, groupID, userID string) error
	ListMemberships(ctx context.Context, groupID string) ([]models.GroupMembership, error)
	GetMembership(ctx context.Context, groupID, userID string) (*models.GroupMembership, error)
}

// GroupService defines the methods for managing groups and memberships.
type GroupService interface {
	// Group CRUD.
	CreateGroup(ctx context.Context, group *models.Group, creatorUserID string) error
	GetGroup(ctx context.Context, groupID string) (*models.Group, error)
	ListGroups(ctx context.Context) ([]models.Group, error)
	UpdateGroup(ctx context.Context, groupID string, update *models.Group, requesterID string) error
	DeleteGroup(ctx context.Context, groupID string, requesterID string) error

	// Membership management.
	AddMember(ctx context.Context, groupID, userID, role, requesterID string) error
	UpdateMemberRole(ctx context.Context, groupID, userID, role, requesterID string) error
	RemoveMember(ctx context.Context, groupID, userID, requesterID string) error
	ListMembers(ctx context.Context, groupID string) ([]models.GroupMembership, error)
}

type groupService struct {
	db GroupDB
}

// NewGroupService creates a new instance of GroupService.
func NewGroupService(db GroupDB) GroupService {
	return &groupService{db: db}
}

// CreateGroup creates a new group and automatically adds the creator as an admin.
func (gs *groupService) CreateGroup(ctx context.Context, group *models.Group, creatorUserID string) error {
	group.ID = uuid.New().String()
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	if err := gs.db.CreateGroup(ctx, group); err != nil {
		return err
	}

	// Create the membership for the creator as admin.
	membership := &models.GroupMembership{
		GroupID:   group.ID,
		UserID:    creatorUserID,
		Role:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return gs.db.CreateMembership(ctx, membership)
}

func (gs *groupService) GetGroup(ctx context.Context, groupID string) (*models.Group, error) {
	return gs.db.GetGroup(ctx, groupID)
}

func (gs *groupService) ListGroups(ctx context.Context) ([]models.Group, error) {
	return gs.db.ListGroups(ctx)
}

// UpdateGroup allows an admin to update the group's name, description, and tags.
func (gs *groupService) UpdateGroup(ctx context.Context, groupID string, update *models.Group, requesterID string) error {
	// Check if the requester is an admin.
	membership, err := gs.db.GetMembership(ctx, groupID, requesterID)
	if err != nil || membership.Role != "admin" {
		return ErrUnauthorized
	}

	group, err := gs.db.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}

	if update.Name != "" {
		group.Name = update.Name
	}
	if update.Description != "" {
		group.Description = update.Description
	}
	if update.Tags != nil {
		group.Tags = update.Tags
	}
	group.UpdatedAt = time.Now()
	return gs.db.UpdateGroup(ctx, group)
}

// DeleteGroup allows an admin to delete the group.
func (gs *groupService) DeleteGroup(ctx context.Context, groupID string, requesterID string) error {
	membership, err := gs.db.GetMembership(ctx, groupID, requesterID)
	if err != nil || membership.Role != "admin" {
		return ErrUnauthorized
	}
	return gs.db.DeleteGroup(ctx, groupID)
}

// AddMember allows an admin to add a new member to the group.
func (gs *groupService) AddMember(ctx context.Context, groupID, userID, role, requesterID string) error {
	membership, err := gs.db.GetMembership(ctx, groupID, requesterID)
	if err != nil || membership.Role != "admin" {
		return ErrUnauthorized
	}
	newMembership := &models.GroupMembership{
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return gs.db.CreateMembership(ctx, newMembership)
}

// UpdateMemberRole allows an admin to update a member's role in the group.
func (gs *groupService) UpdateMemberRole(ctx context.Context, groupID, userID, role, requesterID string) error {
	// Only admin can update membership roles.
	membership, err := gs.db.GetMembership(ctx, groupID, requesterID)
	if err != nil || membership.Role != "admin" {
		return ErrUnauthorized
	}
	targetMembership, err := gs.db.GetMembership(ctx, groupID, userID)
	if err != nil {
		return err
	}
	targetMembership.Role = role
	targetMembership.UpdatedAt = time.Now()
	return gs.db.UpdateMembership(ctx, targetMembership)
}

// RemoveMember allows an admin to remove a member from the group.
func (gs *groupService) RemoveMember(ctx context.Context, groupID, userID, requesterID string) error {
	membership, err := gs.db.GetMembership(ctx, groupID, requesterID)
	if err != nil || membership.Role != "admin" {
		return ErrUnauthorized
	}
	return gs.db.DeleteMembership(ctx, groupID, userID)
}

func (gs *groupService) ListMembers(ctx context.Context, groupID string) ([]models.GroupMembership, error) {
	return gs.db.ListMemberships(ctx, groupID)
}
