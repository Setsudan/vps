package guilds

import "time"

type ChannelType string

const (
	ChannelText  ChannelType = "text"
	ChannelVoice ChannelType = "voice"
)

// Channel lives under an optional Category and inherits its permissions.
type Channel struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	GuildID    string    `json:"guild_id" gorm:"not null;index"`
	CategoryID *string   `json:"category_id" gorm:"index"` // nil = no category
	Name       string    `json:"name"`
	Type       string    `json:"type" gorm:"type:text;default:'text'"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
