package repository

import (
	"errors"
	"fmt"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type IngredientRepository struct {
	DB *gorm.DB
}

func NewIngredientRepository(db *gorm.DB) *IngredientRepository {
	return &IngredientRepository{DB: db}
}

func (r *IngredientRepository) Create(ing *models.Ingredient) error {
	return r.DB.Create(ing).Error
}

func (r *IngredientRepository) GetByTenant(tenantID uint) ([]models.Ingredient, error) {
	var ings []models.Ingredient
	err := r.DB.Where("tenant_id = ?", tenantID).Order("name asc").Find(&ings).Error
	return ings, err
}

func (r *IngredientRepository) GetByID(id, tenantID uint) (*models.Ingredient, error) {
	var ing models.Ingredient
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&ing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ingredient not found")
		}
		return nil, err
	}
	return &ing, nil
}

func (r *IngredientRepository) Update(ing *models.Ingredient) error {
	return r.DB.Save(ing).Error
}

func (r *IngredientRepository) Delete(id, tenantID uint) error {
	var count int64
	r.DB.Model(&models.MenuIngredient{}).Where("ingredient_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("bahan baku masih digunakan oleh menu, hapus dari menu terlebih dahulu")
	}
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Ingredient{}).Error
}

// DeductStock — kurangi stok, dipanggil saat order checkout
func (r *IngredientRepository) DeductStock(items map[uint]float64) error {
	// Validasi semua stok SEBELUM deduct — atomik
	tx := r.DB.Begin()

	// Pass 1: cek semua stok cukup dulu
	for ingID, amount := range items {
		var ing models.Ingredient
		if err := tx.First(&ing, ingID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("bahan baku tidak ditemukan (id=%d)", ingID)
		}
		if ing.Stock < amount {
			tx.Rollback()
			return fmt.Errorf("STOCK_INSUFFICIENT:%d:%s:%.2f:%.2f",
				ingID, ing.Name, ing.Stock, amount)
			// Format: STOCK_INSUFFICIENT:<id>:<nama>:<stok_ada>:<stok_butuh>
		}
	}

	// Pass 2: deduct semua
	for ingID, amount := range items {
		result := tx.Model(&models.Ingredient{}).
			Where("id = ?", ingID).
			Update("stock", gorm.Expr("stock - ?", amount))
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	return tx.Commit().Error
}

func (r *IngredientRepository) GetLowStock(tenantID uint) ([]models.Ingredient, error) {
	var ings []models.Ingredient
	err := r.DB.Where("tenant_id = ? AND low_stock_at > 0 AND stock <= low_stock_at", tenantID).
		Order("stock asc").Find(&ings).Error
	return ings, err
}

func (r *IngredientRepository) GetMenuIngredients(menuID uint) ([]models.MenuIngredient, error) {
	var items []models.MenuIngredient
	err := r.DB.Preload("Ingredient").Where("menu_id = ?", menuID).Find(&items).Error
	return items, err
}

func (r *IngredientRepository) ReplaceMenuIngredients(menuID uint, items []models.MenuIngredient) error {
	tx := r.DB.Begin()
	if err := tx.Where("menu_id = ?", menuID).Delete(&models.MenuIngredient{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(items) > 0 {
		for i := range items {
			items[i].MenuID = menuID
		}
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetIngredientsByMenuIDs — untuk kalkulasi COGS massal di margins
func (r *IngredientRepository) GetIngredientsByMenuIDs(menuIDs []uint) ([]models.MenuIngredient, error) {
	var items []models.MenuIngredient
	err := r.DB.Preload("Ingredient").Where("menu_id IN ?", menuIDs).Find(&items).Error
	return items, err
}

// ── VariationIngredient methods ───────────────────────────────────────────────

func (r *IngredientRepository) GetVariationIngredients(variationID uint) ([]models.VariationIngredient, error) {
	var items []models.VariationIngredient
	err := r.DB.Preload("Ingredient").Where("variation_id = ?", variationID).Find(&items).Error
	return items, err
}

func (r *IngredientRepository) ReplaceVariationIngredients(variationID uint, items []models.VariationIngredient) error {
	tx := r.DB.Begin()
	if err := tx.Where("variation_id = ?", variationID).Delete(&models.VariationIngredient{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(items) > 0 {
		for i := range items {
			items[i].VariationID = variationID
		}
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetVariationIngredientsByIDs — untuk deduct stock massal
func (r *IngredientRepository) GetVariationIngredientsByIDs(variationIDs []uint) ([]models.VariationIngredient, error) {
	if len(variationIDs) == 0 {
		return nil, nil
	}
	var items []models.VariationIngredient
	err := r.DB.Preload("Ingredient").Where("variation_id IN ?", variationIDs).Find(&items).Error
	return items, err
}
