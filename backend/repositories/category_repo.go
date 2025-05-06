package repositories

import (
	"context"

	"launay-dot-one/models/guilds"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *guilds.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*guilds.Category, error) {
	var c guilds.Category
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) ListByGuild(ctx context.Context, guildID string) ([]guilds.Category, error) {
	var out []guilds.Category
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("position ASC").
		Find(&out).Error
	return out, err
}

func (r *CategoryRepository) Update(ctx context.Context, c *guilds.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&guilds.Category{}, "id = ?", id).Error
}
