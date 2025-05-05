package repositories

import (
	"context"

	"launay-dot-one/models/groups"

	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, group *groups.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *GroupRepository) UpdateGroup(ctx context.Context, group *groups.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *GroupRepository) DeleteGroup(ctx context.Context, groupID string) error {
	return r.db.WithContext(ctx).Delete(&groups.Group{}, "id = ?", groupID).Error
}

func (r *GroupRepository) GetGroup(ctx context.Context, groupID string) (*groups.Group, error) {
	var group groups.Group
	if err := r.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) ListGroups(ctx context.Context) ([]groups.Group, error) {
	var groups []groups.Group
	err := r.db.WithContext(ctx).Find(&groups).Error
	return groups, err
}

func (r *GroupRepository) CreateMembership(ctx context.Context, membership *groups.GroupMembership) error {
	return r.db.WithContext(ctx).Create(membership).Error
}

func (r *GroupRepository) UpdateMembership(ctx context.Context, membership *groups.GroupMembership) error {
	return r.db.WithContext(ctx).Save(membership).Error
}

func (r *GroupRepository) DeleteMembership(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).Delete(&groups.GroupMembership{}, "group_id = ? AND user_id = ?", groupID, userID).Error
}

func (r *GroupRepository) ListMemberships(ctx context.Context, groupID string) ([]groups.GroupMembership, error) {
	var memberships []groups.GroupMembership
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&memberships).Error
	return memberships, err
}

func (r *GroupRepository) GetMembership(ctx context.Context, groupID, userID string) (*groups.GroupMembership, error) {
	var membership groups.GroupMembership
	if err := r.db.WithContext(ctx).First(&membership, "group_id = ? AND user_id = ?", groupID, userID).Error; err != nil {
		return nil, err
	}
	return &membership, nil
}
