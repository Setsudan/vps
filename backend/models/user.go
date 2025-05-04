package models

import (
	"time"
)

// User represents a full user record with sensitive fields.
type User struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	Username  string     `json:"username" gorm:"uniqueIndex;not null"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null"`
	Password  string     `json:"password" gorm:"not null"`
	Role      string     `json:"role" gorm:"default:user"`
	Avatar    string     `json:"avatar"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	Bio       string     `json:"bio"`
	Latitude  float64    `json:"latitude" gorm:"type:decimal(9,6)"`
	Longitude float64    `json:"longitude" gorm:"type:decimal(9,6)"`
}

// PublicUser represents a safe version of the User record.
type PublicUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	Latitude  float64   `json:"latitude" gorm:"type:decimal(9,6)"`
	Longitude float64   `json:"longitude" gorm:"type:decimal(9,6)"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToPublic converts a full User to a PublicUser.
func (u *User) ToPublic() PublicUser {

	return PublicUser{
		ID:        u.ID,
		Username:  u.Username,
		Latitude:  u.Latitude,
		Longitude: u.Longitude,
		Bio:       u.Bio,
		Email:     u.Email,
		Role:      u.Role,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
