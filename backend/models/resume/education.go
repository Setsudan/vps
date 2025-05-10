package resume

import "time"

type Education struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ResumeID    string     `json:"resume_id" gorm:"not null;index"`
	Institution string     `json:"institution"`
	Degree      string     `json:"degree"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
