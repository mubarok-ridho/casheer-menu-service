package repository

import (
	"errors"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type VariationRepository struct {
	DB *gorm.DB
}

func NewVariationRepository(db *gorm.DB) *VariationRepository {
	return &VariationRepository{DB: db}
}

// Create variation
func (r *VariationRepository) Create(variation *models.MenuVariation, tenantID uint) error {
	// Verify menu belongs to tenant
	var menu models.Menu
	if err := r.DB.Where("id = ? AND tenant_id = ?", variation.MenuID, tenantID).First(&menu).Error; err != nil {
		return errors.New("menu not found or unauthorized")
	}

	return r.DB.Create(variation).Error
}

// Get variation by ID
func (r *VariationRepository) GetByID(id uint, tenantID uint) (*models.MenuVariation, error) {
	var variation models.MenuVariation
	err := r.DB.Joins("JOIN menus ON menus.id = menu_variations.menu_id").
		Where("menu_variations.id = ? AND menus.tenant_id = ?", id, tenantID).
		First(&variation).Error
	return &variation, err
}

// Get variations by menu
func (r *VariationRepository) GetByMenu(menuID uint, tenantID uint) ([]models.MenuVariation, error) {
	var variations []models.MenuVariation
	err := r.DB.Joins("JOIN menus ON menus.id = menu_variations.menu_id").
		Where("menu_variations.menu_id = ? AND menus.tenant_id = ?", menuID, tenantID).
		Find(&variations).Error
	return variations, err
}

// Update variation
func (r *VariationRepository) Update(variation *models.MenuVariation) error {
	return r.DB.Save(variation).Error
}

// Delete variation
func (r *VariationRepository) Delete(id uint, tenantID uint) error {
	return r.DB.Joins("JOIN menus ON menus.id = menu_variations.menu_id").
		Where("menu_variations.id = ? AND menus.tenant_id = ?", id, tenantID).
		Delete(&models.MenuVariation{}).Error
}

// Update stock
func (r *VariationRepository) UpdateStock(id uint, tenantID uint, stock int) error {
	return r.DB.Model(&models.MenuVariation{}).
		Joins("JOIN menus ON menus.id = menu_variations.menu_id").
		Where("menu_variations.id = ? AND menus.tenant_id = ?", id, tenantID).
		Update("stock", stock).Error
}
