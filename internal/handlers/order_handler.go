package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"github.com/mubarok-ridho/casheer-menu-service/pkg/messaging"
)

type OrderHandler struct {
	repo    *repository.OrderRepository
	ingRepo *repository.IngredientRepository
	rmq     *messaging.RabbitMQ
}

func NewOrderHandler(repo *repository.OrderRepository, ingRepo *repository.IngredientRepository) *OrderHandler {
	rmq, _ := messaging.NewRabbitMQ()
	return &OrderHandler{repo: repo, ingRepo: ingRepo, rmq: rmq}
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	type OrderItemInput struct {
		MenuID      uint    `json:"menu_id"`
		VariationID *uint   `json:"variation_id"`
		PromoID     *uint   `json:"promo_id"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
		Notes       string  `json:"notes"`
	}
	var input struct {
		CustomerName   string           `json:"customer_name"`
		CustomerPhone  string           `json:"customer_phone"`
		PaymentMethod  string           `json:"payment_method"`
		Notes          string           `json:"notes"`
		DiscountType   string           `json:"discount_type"`
		DiscountAmount float64          `json:"discount_amount"`
		Items          []OrderItemInput `json:"items"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if len(input.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Order must have at least one item"})
	}

	orderNumber := "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
	order := &models.Order{
		TenantID:       tenantID,
		OrderNumber:    orderNumber,
		CustomerName:   input.CustomerName,
		CustomerPhone:  input.CustomerPhone,
		PaymentMethod:  input.PaymentMethod,
		PaymentStatus:  "paid",
		OrderStatus:    "completed",
		Notes:          input.Notes,
		DiscountType:   input.DiscountType,
		DiscountAmount: input.DiscountAmount,
		TotalAmount:    0,
	}

	menuQtyMap := make(map[uint]int)
	varQtyMap := make(map[uint]int)

	for _, item := range input.Items {
		if item.MenuID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Setiap item harus memiliki menu_id yang valid"})
		}
		if item.Quantity <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Quantity harus lebih dari 0"})
		}

		subtotal := float64(item.Quantity) * item.Price
		order.TotalAmount += subtotal
		order.Items = append(order.Items, models.OrderItem{
			MenuID:      item.MenuID,
			VariationID: item.VariationID,
			PromoID:     item.PromoID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    subtotal,
			Notes:       item.Notes,
		})

		menuQtyMap[item.MenuID] += item.Quantity
		if item.VariationID != nil {
			varQtyMap[*item.VariationID] += item.Quantity
		}
	}

	if input.DiscountAmount > 0 {
		order.TotalAmount -= input.DiscountAmount
		if order.TotalAmount < 0 {
			order.TotalAmount = 0
		}
	}

	if h.ingRepo != nil {
		if err := h.deductIngredients(menuQtyMap, varQtyMap); err != nil {
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "STOCK_INSUFFICIENT:") {
				parts := strings.Split(errMsg, ":")
				if len(parts) >= 5 {
					nama := parts[2]
					ada, _ := strconv.ParseFloat(parts[3], 64)
					butuh, _ := strconv.ParseFloat(parts[4], 64)
					return c.Status(400).JSON(fiber.Map{
						"error":      fmt.Sprintf("Stok %s tidak mencukupi (tersedia: %.0f, dibutuhkan: %.0f)", nama, ada, butuh),
						"error_code": "STOCK_INSUFFICIENT",
						"ingredient": nama,
						"available":  ada,
						"required":   butuh,
					})
				}
			}
			return c.Status(400).JSON(fiber.Map{
				"error":      errMsg,
				"error_code": "STOCK_INSUFFICIENT",
			})
		}
	}

	// Baru simpan order setelah stock berhasil dideduct
	if err := h.repo.Create(order); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if h.rmq != nil {
		h.rmq.PublishOrderCompleted(messaging.OrderCompletedEvent{
			OrderID:     order.ID,
			TenantID:    tenantID,
			TotalAmount: order.TotalAmount,
			Date:        time.Now().Format("2006-01-02"),
		})
	}
	return c.Status(201).JSON(order)
}

func (h *OrderHandler) deductIngredients(menuQtyMap map[uint]int, varQtyMap map[uint]int) error {
	deductMap := make(map[uint]float64)

	menuIDs := make([]uint, 0, len(menuQtyMap))
	for id := range menuQtyMap {
		menuIDs = append(menuIDs, id)
	}
	menuIngs, err := h.ingRepo.GetIngredientsByMenuIDs(menuIDs)
	if err != nil {
		return fmt.Errorf("gagal ambil bahan baku menu: %w", err)
	}
	for _, mi := range menuIngs {
		deductMap[mi.IngredientID] += mi.Amount * float64(menuQtyMap[mi.MenuID])
	}

	if len(varQtyMap) > 0 {
		varIDs := make([]uint, 0, len(varQtyMap))
		for id := range varQtyMap {
			varIDs = append(varIDs, id)
		}
		varIngs, err := h.ingRepo.GetVariationIngredientsByIDs(varIDs)
		if err != nil {
			return fmt.Errorf("gagal ambil bahan baku variasi: %w", err)
		}
		for _, vi := range varIngs {
			deductMap[vi.IngredientID] += vi.Amount * float64(varQtyMap[vi.VariationID])
		}
	}

	if len(deductMap) == 0 {
		return nil
	}
	return h.ingRepo.DeductStock(deductMap)
}

func (h *OrderHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	orders, total, err := h.repo.GetAll(tenantID, page, limit, startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": orders, "total": total, "page": page, "limit": limit})
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
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
	days := c.QueryInt("days", 7)
	orders, err := h.repo.GetByTenant(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

func (h *OrderHandler) DeleteOldOrders(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	days := c.QueryInt("days", 30)
	if days < 7 {
		return c.Status(400).JSON(fiber.Map{"error": "Minimum 7 hari"})
	}
	deleted, err := h.repo.DeleteOlderThan(tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Cleanup berhasil", "deleted": deleted, "days": days})
}
