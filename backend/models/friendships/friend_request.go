package friendships

import "time"

type FriendStatus string

const (
	FriendPending  FriendStatus = "pending"
	FriendAccepted FriendStatus = "accepted"
	FriendRejected FriendStatus = "rejected"
	FriendBlocked  FriendStatus = "blocked"
)

type FriendRequest struct {
	ID          string       `json:"id" gorm:"primaryKey;type:uuid"`
	RequesterID string       `json:"requester_id"`
	ReceiverID  string       `json:"receiver_id"`
	Status      FriendStatus `json:"status" gorm:"type:text;default:'pending'"`
	BlockerID   string       `json:"blocker_id,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
