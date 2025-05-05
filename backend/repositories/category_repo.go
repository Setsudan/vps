package repositories

import (
	"context"
	"gorm.io/gorm"
	"launay-dot-one/models/guilds"
)

type CategoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *guilds.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepository) Update(ctx context.Context, c *guilds.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryID string) error {
	return r.db.WithContext(ctx).
		Delete(&guilds.Category{}, "id = ?", categoryID).
		Error
}

func (r *CategoryRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.Category, error) {
	var cats []guilds.Category
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("position ASC").
		Find(&cats).Error
	return cats, err
}
