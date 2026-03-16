package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"github.com/mubarok-ridho/casheer-menu-service/pkg/messaging"
)

type OrderHandler struct {
	repo *repository.OrderRepository
	rmq  *messaging.RabbitMQ
}

func NewOrderHandler(repo *repository.OrderRepository) *OrderHandler {
	rmq, _ := messaging.NewRabbitMQ()
	return &OrderHandler{repo: repo, rmq: rmq}
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	type OrderItemInput struct {
		MenuID      uint    `json:"menu_id"`
		VariationID *uint   `json:"variation_id"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
		Notes       string  `json:"notes"`
	}
	var input struct {
		CustomerName  string           `json:"customer_name"`
		CustomerPhone string           `json:"customer_phone"`
		PaymentMethod string           `json:"payment_method"`
		Notes         string           `json:"notes"`
		Items         []OrderItemInput `json:"items"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if len(input.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Order must have at least one item"})
	}

	orderNumber := "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
	order := &models.Order{
		TenantID:      tenantID,
		OrderNumber:   orderNumber,
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		PaymentMethod: input.PaymentMethod,
		PaymentStatus: "paid",
		OrderStatus:   "completed",
		Notes:         input.Notes,
		TotalAmount:   0,
	}
	for _, item := range input.Items {
		subtotal := float64(item.Quantity) * item.Price
		order.TotalAmount += subtotal
		order.Items = append(order.Items, models.OrderItem{
			MenuID:      item.MenuID,
			VariationID: item.VariationID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    subtotal,
			Notes:       item.Notes,
		})
	}
	if err := h.repo.Create(order); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if h.rmq != nil {
		event := messaging.OrderCompletedEvent{
			OrderID:     order.ID,
			TenantID:    tenantID,
			TotalAmount: order.TotalAmount,
			Date:        time.Now().Format("2006-01-02"),
		}
		h.rmq.PublishOrderCompleted(event)
	}
	return c.Status(201).JSON(order)
}

func (h *OrderHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	page      := c.QueryInt("page", 1)
	limit     := c.QueryInt("limit", 20)
	startDate := c.Query("start_date")
	endDate   := c.Query("end_date")

	orders, total, err := h.repo.GetAll(tenantID, page, limit, startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": orders, "total": total, "page": page, "limit": limit})
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err  := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order ID"})
	}
	order, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Order not found"})
	}
	return c.JSON(order)
}

func (h *OrderHandler) GetByTenant(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	days     := c.QueryInt("days", 7)
	orders, err := h.repo.GetByTenant(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

// DeleteOldOrders — cleanup orders older than X days
func (h *OrderHandler) DeleteOldOrders(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	days     := c.QueryInt("days", 30)
	if days < 7 {
		return c.Status(400).JSON(fiber.Map{"error": "Minimum 7 hari"})
	}
	deleted, err := h.repo.DeleteOlderThan(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Cleanup berhasil", "deleted": deleted, "days": days})
}
