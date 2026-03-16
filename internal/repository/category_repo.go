package repository

import (
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.DB.Create(category).Error
}

func (r *CategoryRepository) GetByTenant(tenantID uint) ([]models.Category, error) {
	var categories []models.Category
	err := r.DB.Where("tenant_id = ? AND is_active = true", tenantID).
		Order("sort_order asc, name asc").
		Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetByID(id uint, tenantID uint) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.DB.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint, tenantID uint) error {
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&models.Category{}).Error
}
