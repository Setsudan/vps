package repositories

import (
	"context"
	"gorm.io/gorm"
	"launay-dot-one/models/friendships"
)

type FriendRequestRepository struct{ db *gorm.DB }

func NewFriendRequestRepository(db *gorm.DB) *FriendRequestRepository {
	return &FriendRequestRepository{db}
}

func (r *FriendRequestRepository) Create(ctx context.Context, fr *friendships.FriendRequest) error {
	return r.db.WithContext(ctx).Create(fr).Error
}

func (r *FriendRequestRepository) UpdateStatus(ctx context.Context, id string, status friendships.FriendStatus) error {
	return r.db.WithContext(ctx).
		Model(&friendships.FriendRequest{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *FriendRequestRepository) Get(ctx context.Context, id string) (*friendships.FriendRequest, error) {
	var fr friendships.FriendRequest
	if err := r.db.WithContext(ctx).First(&fr, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &fr, nil
}

func (r *FriendRequestRepository) ListForUser(ctx context.Context, userID string) ([]friendships.FriendRequest, error) {
	var list []friendships.FriendRequest
	err := r.db.WithContext(ctx).
		Where("requester_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// ListFriends returns accepted friendships for a user.
func (r *FriendRequestRepository) ListFriends(ctx context.Context, userID string) ([]friendships.FriendRequest, error) {
	var friends []friendships.FriendRequest
	err := r.db.WithContext(ctx).
		Where("status = ? AND (requester_id = ? OR receiver_id = ?)",
			friendships.FriendAccepted, userID, userID).
		Find(&friends).Error
	return friends, err
}
