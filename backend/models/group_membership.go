package models

import "time"

// GroupMembership represents a user’s membership in a group with an associated role.
type GroupMembership struct {
	GroupID   string    `json:"group_id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"primaryKey"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
