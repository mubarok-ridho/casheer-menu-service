package repository

import (
	"time"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

// Create order
func (r *OrderRepository) Create(order *models.Order) error {
	return r.DB.Create(order).Error
}

// Get order by ID
func (r *OrderRepository) GetByID(id uint, tenantID uint) (*models.Order, error) {
	var order models.Order
	err := r.DB.Preload("Items.Menu").
		Preload("Items.Variation").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&order).Error
	return &order, err
}

// Get all orders with pagination
func (r *OrderRepository) GetAll(tenantID uint, page, limit int, startDate, endDate string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.DB.Model(&models.Order{}).Where("tenant_id = ?", tenantID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Items.Menu").
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

// Get orders by tenant (recent days)
func (r *OrderRepository) GetByTenant(tenantID uint, days int) ([]models.Order, error) {
	var orders []models.Order
	startDate := time.Now().AddDate(0, 0, -days)

	err := r.DB.Preload("Items").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Order("created_at desc").
		Find(&orders).Error

	return orders, err
}

// Get daily revenue
func (r *OrderRepository) GetDailyRevenue(tenantID uint, date string) (float64, error) {
	var total float64
	err := r.DB.Model(&models.Order{}).
		Where("tenant_id = ? AND DATE(created_at) = ? AND payment_status = 'paid'", tenantID, date).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}
