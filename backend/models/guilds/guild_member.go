package guilds

import (
	"time"

	"gorm.io/datatypes"
)

// GuildMember links a User to a Guild with one or more Roles.
type GuildMember struct {
	GuildID string `json:"guild_id" gorm:"primaryKey;index"`
	UserID  string `json:"user_id" gorm:"primaryKey;index"`
	// You can allow multiple roles per member; this could instead be a join table.
	RoleIDs   datatypes.JSON `json:"role_ids" gorm:"type:jsonb"`
	JoinedAt  time.Time      `json:"joined_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
