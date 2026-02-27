package utils

import (
	"time"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type BestSellerCalculator struct {
	DB *gorm.DB
}

func NewBestSellerCalculator(db *gorm.DB) *BestSellerCalculator {
	return &BestSellerCalculator{DB: db}
}

// Calculate best sellers based on order count
func (c *BestSellerCalculator) Calculate(tenantID uint, limit int, days int) ([]models.Menu, error) {
	var menus []models.Menu

	startDate := time.Now().AddDate(0, 0, -days)

	err := c.DB.Table("menus").
		Select("menus.*, COUNT(order_items.id) as order_count").
		Joins("LEFT JOIN order_items ON order_items.menu_id = menus.id").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Where("menus.tenant_id = ? AND orders.created_at >= ?", tenantID, startDate).
		Group("menus.id").
		Order("order_count DESC").
		Limit(limit).
		Preload("Images").
		Find(&menus).Error

	return menus, err
}

// Get popular variations
func (c *BestSellerCalculator) GetPopularVariations(menuID uint, tenantID uint, limit int) ([]models.MenuVariation, error) {
	var variations []models.MenuVariation

	err := c.DB.Table("menu_variations").
		Select("menu_variations.*, COUNT(order_items.id) as usage_count").
		Joins("LEFT JOIN order_items ON order_items.variation_id = menu_variations.id").
		Joins("JOIN menus ON menus.id = menu_variations.menu_id").
		Where("menu_variations.menu_id = ? AND menus.tenant_id = ?", menuID, tenantID).
		Group("menu_variations.id").
		Order("usage_count DESC").
		Limit(limit).
		Find(&variations).Error

	return variations, err
}
