package models

import (
	"time"

	"gorm.io/datatypes"
)

type Message struct {
	ID         string `json:"id" gorm:"primaryKey"`
	SenderID   string `json:"sender_id"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"` // "user", "group", or "channel"
	Content    string `json:"content"`
	// Store attachments and reactions as JSON.
	Attachments datatypes.JSON `json:"attachments,omitempty" gorm:"type:jsonb"`
	Reactions   datatypes.JSON `json:"reactions,omitempty" gorm:"type:jsonb"`
	Timestamp   time.Time      `json:"timestamp"`
}
