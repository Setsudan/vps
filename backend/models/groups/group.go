package groups

import (
	"time"

	"gorm.io/datatypes"
)

type Group struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tags        datatypes.JSON `json:"tags,omitempty" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
