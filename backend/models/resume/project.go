package resume

import "time"

type Project struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ResumeID    string    `json:"resume_id" gorm:"not null;index"`
	Name        string    `json:"name"`
	Description string    `json:"description" gorm:"type:text"`
	URL         string    `json:"url"`
	RepoURL     string    `json:"repo_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
