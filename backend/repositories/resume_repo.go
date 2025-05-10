package repositories

import (
	"context"
	"gorm.io/gorm"
	"launay-dot-one/models"
)

type ResumeRepository struct{ db *gorm.DB }

func NewResumeRepository(db *gorm.DB) *ResumeRepository {
	return &ResumeRepository{db}
}

func (r *ResumeRepository) Create(ctx context.Context, res *models.Resume) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *ResumeRepository) Update(ctx context.Context, res *models.Resume) error {
	return r.db.WithContext(ctx).Save(res).Error
}

func (r *ResumeRepository) Delete(ctx context.Context, resumeID string) error {
	return r.db.WithContext(ctx).
		Delete(&models.Resume{}, "id = ?", resumeID).
		Error
}

func (r *ResumeRepository) GetByUser(ctx context.Context, userID string) (*models.Resume, error) {
	var res models.Resume
	err := r.db.WithContext(ctx).
		Preload("Educations").
		Preload("Experiences").
		Preload("Projects").
		Preload("Certifications").
		Preload("Skills").
		Preload("Interests").
		First(&res, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ResumeRepository) ListAll(ctx context.Context) ([]models.Resume, error) {
	var list []models.Resume
	return list, r.db.WithContext(ctx).Find(&list).Error
}
