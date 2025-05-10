package groups

import (
	"context"

	m "launay-dot-one/models/groups"
)

// Service defines the group-and-membership business logic.
type Service interface {
	// Group CRUD
	CreateGroup(ctx context.Context, grp *m.Group, creatorID string) error
	GetGroup(ctx context.Context, groupID string) (*m.Group, error)
	ListGroups(ctx context.Context) ([]m.Group, error)
	UpdateGroup(ctx context.Context, groupID string, upd *m.Group, requesterID string) error
	DeleteGroup(ctx context.Context, groupID, requesterID string) error

	// Membership management
	AddMember(ctx context.Context, groupID, userID, role, requesterID string) error
	UpdateMemberRole(ctx context.Context, groupID, userID, role, requesterID string) error
	RemoveMember(ctx context.Context, groupID, userID, requesterID string) error
	ListMembers(ctx context.Context, groupID string) ([]m.GroupMembership, error)
}
