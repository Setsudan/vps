// models/resume.go
package models

import (
	models "launay-dot-one/models/resume"
	"time"
)

type Resume struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Educations     []models.Education     `gorm:"foreignKey:ResumeID"`
	Experiences    []models.Experience    `gorm:"foreignKey:ResumeID"`
	Projects       []models.Project       `gorm:"foreignKey:ResumeID"`
	Certifications []models.Certification `gorm:"foreignKey:ResumeID"`
	Skills         []models.Skill         `gorm:"foreignKey:ResumeID"`
	Interests      []models.Interest      `gorm:"foreignKey:ResumeID"`
}
