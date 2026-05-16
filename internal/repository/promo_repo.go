package repository

import (
	"time"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type PromoRepository struct {
	DB *gorm.DB
}

func NewPromoRepository(db *gorm.DB) *PromoRepository {
	return &PromoRepository{DB: db}
}

func (r *PromoRepository) Create(promo *models.Promo) error {
	return r.DB.Create(promo).Error
}

func (r *PromoRepository) GetByTenant(tenantID uint, activeOnly bool) ([]models.Promo, error) {
	var promos []models.Promo
	query := r.DB.
		Preload("Items.Menu.Images").
		Preload("Items.Menu.Variations").
		Preload("Items.Variation"). // load variation yang dipilih di promo item
		Where("tenant_id = ?", tenantID)
	if activeOnly {
		now := time.Now()
		query = query.Where(
			"is_active = true AND start_at <= ? AND (end_at IS NULL OR end_at >= ?)",
			now, now,
		)
	}
	err := query.Order("created_at desc").Find(&promos).Error
	return promos, err
}

func (r *PromoRepository) GetByID(id uint, tenantID uint) (*models.Promo, error) {
	var promo models.Promo
	err := r.DB.
		Preload("Items.Menu.Images").
		Preload("Items.Menu.Variations").
		Preload("Items.Variation").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&promo).Error
	return &promo, err
}

func (r *PromoRepository) Update(promo *models.Promo) error {
	return r.DB.Session(&gorm.Session{FullSaveAssociations: false}).Save(promo).Error
}

func (r *PromoRepository) Delete(id uint, tenantID uint) error {
	r.DB.Where("promo_id = ?", id).Delete(&models.PromoItem{})
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Promo{}).Error
}

func (r *PromoRepository) ReplaceItems(promoID uint, items []models.PromoItem) error {
	r.DB.Where("promo_id = ?", promoID).Delete(&models.PromoItem{})
	for i := range items {
		items[i].PromoID = promoID
	}
	if len(items) == 0 {
		return nil
	}
	return r.DB.Create(&items).Error
}
