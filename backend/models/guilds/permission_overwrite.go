package guilds

import "time"

// OverwriteType indicates whether it's for a role or a member.
type OverwriteType string

const (
	OverwriteRole   OverwriteType = "role"
	OverwriteMember OverwriteType = "member"
)

// PermissionOverwrite stores allow/deny bitfields on a Category or Channel.
type PermissionOverwrite struct {
	ID            string        `json:"id" gorm:"primaryKey;type:uuid"`
	GuildID       string        `json:"guild_id"`
	CategoryID    string        `json:"category_id,omitempty"`
	ChannelID     string        `json:"channel_id,omitempty"`
	OverwriteType OverwriteType `json:"overwrite_type" gorm:"type:text;default:'role'"`
	TargetID      string        `json:"target_id"`
	Allow         int64         `json:"allow"`
	Deny          int64         `json:"deny"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
