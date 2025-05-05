package guilds

import "time"

type ChannelMember struct {
	ChannelID string    `json:"channel_id" gorm:"primaryKey;index"`
	UserID    string    `json:"user_id" gorm:"primaryKey;index"`
	Role      string    `json:"role" gorm:"type:enum('moderator','member');default:'member'"`
	JoinedAt  time.Time `json:"joined_at"`
}
