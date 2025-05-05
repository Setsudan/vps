package models

import (
	"time"
)

// Status represents a user's online state.
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusIdle    Status = "idle"
	StatusDND     Status = "dnd"
)

// User is your full user record.
type User struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username   string     `json:"username" gorm:"uniqueIndex;not null"`
	Email      string     `json:"email" gorm:"uniqueIndex;not null"`
	Password   string     `json:"password" gorm:"not null"`
	Role       string     `json:"role" gorm:"type:text;default:'user'"`
	Avatar     string     `json:"avatar"`
	Bio        string     `json:"bio"`
	Status     Status     `json:"status" gorm:"type:text;default:'offline'"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// PublicUser is the safe projection for API responses.
type PublicUser struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Bio        string     `json:"bio"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Avatar     string     `json:"avatar"`
	Status     Status     `json:"status"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ToPublic converts the full User into its PublicUser view.
func (u *User) ToPublic() PublicUser {
	return PublicUser{
		ID:         u.ID,
		Username:   u.Username,
		Bio:        u.Bio,
		Email:      u.Email,
		Role:       u.Role,
		Avatar:     u.Avatar,
		Status:     u.Status,
		LastSeenAt: u.LastSeenAt,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}
