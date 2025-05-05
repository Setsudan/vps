package resume

import "time"

type Experience struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ResumeID    string     `json:"resume_id" gorm:"not null;index"`
	Company     string     `json:"company"`
	JobTitle    string     `json:"job_title"`
	Location    string     `json:"location"`
	Description string     `json:"description" gorm:"type:text"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
