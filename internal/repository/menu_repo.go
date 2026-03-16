package repository

import (
	"errors"
	"fmt"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type MenuRepository struct {
	DB *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{DB: db}
}

func (r *MenuRepository) Create(menu *models.Menu) error {
	return r.DB.Create(menu).Error
}

func (r *MenuRepository) GetByID(id uint, tenantID uint) (*models.Menu, error) {
	var menu models.Menu
	err := r.DB.Preload("Category").
		Preload("Variations").
		Preload("Images").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("menu not found")
		}
		return nil, err
	}
	return &menu, nil
}

func (r *MenuRepository) GetByTenant(tenantID uint, categoryID int, search string) ([]models.Menu, error) {
	var menus []models.Menu
	query := r.DB.Preload("Category").
		Preload("Variations").
		Preload("Images").
		Where("tenant_id = ?", tenantID)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Order("created_at desc").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetPublicMenus(tenantID uint, categoryID int) ([]models.Menu, error) {
	var menus []models.Menu
	query := r.DB.Preload("Category").
		Preload("Variations").
		Preload("Images").
		Where("tenant_id = ? AND is_available = true", tenantID)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	err := query.Order("category_id asc, name asc").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) Update(menu *models.Menu) error {
	return r.DB.Session(&gorm.Session{FullSaveAssociations: false}).
		Select("category_id", "name", "description", "base_price", "prep_time", "is_available").
		Save(menu).Error
}

func (r *MenuRepository) Delete(id uint, tenantID uint) error {
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&models.Menu{}).Error
}

func (r *MenuRepository) AddImage(image *models.MenuImage) error {
	return r.DB.Create(image).Error
}

func (r *MenuRepository) DeleteImage(imageID uint, menuID uint, tenantID uint) error {
	var menu models.Menu
	if err := r.DB.Where("id = ? AND tenant_id = ?", menuID, tenantID).First(&menu).Error; err != nil {
		return errors.New("menu not found or unauthorized")
	}
	return r.DB.Where("id = ? AND menu_id = ?", imageID, menuID).
		Delete(&models.MenuImage{}).Error
}

func (r *MenuRepository) GetBestSeller(tenantID uint, limit int, days int) ([]models.Menu, error) {
	var menus []models.Menu
	subQuery := r.DB.Table("order_items").
		Select("menu_id, COUNT(*) as order_count").
		Joins("join orders on orders.id = order_items.order_id").
		Where(fmt.Sprintf("orders.tenant_id = ? AND orders.created_at > NOW() - INTERVAL '%d days'", days), tenantID).
		Group("menu_id").
		Order("order_count desc").
		Limit(limit)
	err := r.DB.Preload("Images").
		Joins("join (?) as best on best.menu_id = menus.id", subQuery).
		Order("best.order_count desc").
		Find(&menus).Error
	return menus, err
}
