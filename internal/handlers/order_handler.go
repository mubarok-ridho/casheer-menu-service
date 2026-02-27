package handlers

import (
	"casheer-menu-service/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type OrderHandler struct {
	DB *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

// Create order
func (h *OrderHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	type OrderItem struct {
		MenuID      uint    `json:"menu_id"`
		VariationID *uint   `json:"variation_id"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
		Note        string  `json:"note"`
	}

	var input struct {
		Items   []OrderItem `json:"items"`
		Total   float64     `json:"total"`
		Payment string      `json:"payment_method"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	tx := h.DB.Begin()

	// Create order
	order := models.Order{
		TenantID:      tenantID,
		OrderNumber:   generateOrderNumber(),
		TotalAmount:   input.Total,
		PaymentMethod: input.Payment,
		Status:        "completed",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create order items dan update stock
	for _, item := range input.Items {
		orderItem := models.OrderItem{
			OrderID:     order.ID,
			MenuID:      item.MenuID,
			VariationID: item.VariationID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Note:        item.Note,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Update stock jika ada variasi
		if item.VariationID != nil {
			tx.Model(&models.MenuVariation{}).
				Where("id = ?", item.VariationID).
				Update("stock", gorm.Expr("stock - ?", item.Quantity))
		}
	}

	tx.Commit()

	// Kirim event ke report service via RabbitMQ untuk update revenue
	utils.PublishOrderCompleted(order)

	return c.Status(201).JSON(order)
}
