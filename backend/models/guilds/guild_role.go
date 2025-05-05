package guilds

import "time"

// GuildRole defines a named set of permissions within a Guild.
type GuildRole struct {
	ID      string `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	GuildID string `json:"guild_id" gorm:"not null;index"`
	Name    string `json:"name"`
	// Permissions bitfield (e.g. view_channel, send_messages, manage_roles…)
	Permissions uint64    `json:"permissions" gorm:"type:bigint"`
	Color       int       `json:"color"`    // optional display color
	Hoist       bool      `json:"hoist"`    // display separately
	Position    int       `json:"position"` // order in role list
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
