package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
)

type IngredientHandler struct {
	repo *repository.IngredientRepository
}

func NewIngredientHandler(repo *repository.IngredientRepository) *IngredientHandler {
	return &IngredientHandler{repo: repo}
}

func (h *IngredientHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	ings, err := h.repo.GetByTenant(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ings)
}

func (h *IngredientHandler) GetLowStock(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	ings, err := h.repo.GetLowStock(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ings)
}

func (h *IngredientHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	var input struct {
		Name        string  `json:"name"`
		Unit        string  `json:"unit"`
		Stock       float64 `json:"stock"`
		CostPerUnit float64 `json:"cost_per_unit"`
		LowStockAt  float64 `json:"low_stock_at"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" || input.Unit == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name dan unit wajib diisi"})
	}
	ing := &models.Ingredient{
		TenantID:    tenantID,
		Name:        input.Name,
		Unit:        input.Unit,
		Stock:       input.Stock,
		CostPerUnit: input.CostPerUnit,
		LowStockAt:  input.LowStockAt,
	}
	if err := h.repo.Create(ing); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(ing)
}

func (h *IngredientHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	ing, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	var input struct {
		Name        string  `json:"name"`
		Unit        string  `json:"unit"`
		Stock       float64 `json:"stock"`
		CostPerUnit float64 `json:"cost_per_unit"`
		LowStockAt  float64 `json:"low_stock_at"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	ing.Name = input.Name
	ing.Unit = input.Unit
	ing.Stock = input.Stock
	ing.CostPerUnit = input.CostPerUnit
	ing.LowStockAt = input.LowStockAt
	if err := h.repo.Update(ing); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ing)
}

func (h *IngredientHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Bahan baku dihapus"})
}

func (h *IngredientHandler) GetMenuIngredients(c *fiber.Ctx) error {
	menuID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	items, err := h.repo.GetMenuIngredients(uint(menuID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *IngredientHandler) SetMenuIngredients(c *fiber.Ctx) error {
	menuID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	var input []struct {
		IngredientID uint    `json:"ingredient_id"`
		Amount       float64 `json:"amount"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	items := make([]models.MenuIngredient, len(input))
	for i, v := range input {
		items[i] = models.MenuIngredient{
			MenuID:       uint(menuID),
			IngredientID: v.IngredientID,
			Amount:       v.Amount,
		}
	}
	if err := h.repo.ReplaceMenuIngredients(uint(menuID), items); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Ingredients updated", "count": len(items)})
}

// GET /variations/:id/ingredients
func (h *IngredientHandler) GetVariationIngredients(c *fiber.Ctx) error {
	varID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid variation ID"})
	}
	items, err := h.repo.GetVariationIngredients(uint(varID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// PUT /variations/:id/ingredients
func (h *IngredientHandler) SetVariationIngredients(c *fiber.Ctx) error {
	varID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid variation ID"})
	}
	var input []struct {
		IngredientID uint    `json:"ingredient_id"`
		Amount       float64 `json:"amount"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	items := make([]models.VariationIngredient, len(input))
	for i, v := range input {
		items[i] = models.VariationIngredient{
			VariationID:  uint(varID),
			IngredientID: v.IngredientID,
			Amount:       v.Amount,
		}
	}
	if err := h.repo.ReplaceVariationIngredients(uint(varID), items); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Variation ingredients updated", "count": len(items)})
}
