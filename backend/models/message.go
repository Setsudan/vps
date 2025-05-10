package models

import (
	"time"

	"gorm.io/datatypes"
)

type Message struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ChannelID   string         `json:"channel_id" gorm:"not null;index"`
	AuthorID    string         `json:"author_id" gorm:"not null;index"`
	Content     string         `json:"content" gorm:"type:text"`
	Attachments datatypes.JSON `json:"attachments,omitempty" gorm:"type:jsonb"`
	Reactions   datatypes.JSON `json:"reactions,omitempty" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
