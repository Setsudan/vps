package friendships

import (
	"context"

	mfriend "launay-dot-one/models/friendships"
)

// Service defines the friendship‐related business logic.
type Service interface {
	// SendRequest creates a new friend request from fromID to toID.
	SendRequest(ctx context.Context, fromID, toID string) error

	// RespondRequest accepts or rejects an existing request.
	RespondRequest(ctx context.Context, requestID string, accept bool) error

	// ListRequests fetches all incoming or outgoing requests for a user.
	ListRequests(ctx context.Context, userID string) ([]RequestDTO, error)

	// ListFriends returns the IDs of all accepted friends for a user.
	ListFriends(ctx context.Context, userID string) ([]FriendDTO, error)
}

// RequestDTO is a safe projection of a FriendRequest model.
type RequestDTO struct {
	ID          string
	RequesterID string
	ReceiverID  string
	Status      mfriend.FriendStatus
}

// FriendDTO represents a simple friend entry.
type FriendDTO struct {
	UserID string
}
