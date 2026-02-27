package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
)

type VariationHandler struct {
	repo *repository.VariationRepository
}

func NewVariationHandler(repo *repository.VariationRepository) *VariationHandler {
	return &VariationHandler{repo: repo}
}

// Create variation
func (h *VariationHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var input struct {
		MenuID uint    `json:"menu_id"`
		Name   string  `json:"name"`
		Option string  `json:"option"`
		Price  float64 `json:"price"`
		Stock  int     `json:"stock"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validasi
	if input.Name == "" || input.Option == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and option are required",
		})
	}

	variation := &models.MenuVariation{
		MenuID: input.MenuID,
		Name:   input.Name,
		Option: input.Option,
		Price:  input.Price,
		Stock:  input.Stock,
	}

	if err := h.repo.Create(variation, tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(variation)
}

// Update variation
func (h *VariationHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid variation ID"})
	}

	var input struct {
		Name   string  `json:"name"`
		Option string  `json:"option"`
		Price  float64 `json:"price"`
		Stock  int     `json:"stock"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	variation, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Variation not found"})
	}

	variation.Name = input.Name
	variation.Option = input.Option
	variation.Price = input.Price
	variation.Stock = input.Stock

	if err := h.repo.Update(variation); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(variation)
}

// Delete variation
func (h *VariationHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid variation ID"})
	}

	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Variation deleted successfully"})
}

// Update stock
func (h *VariationHandler) UpdateStock(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid variation ID"})
	}

	var input struct {
		Stock int `json:"stock"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.repo.UpdateStock(uint(id), tenantID, input.Stock); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Stock updated successfully"})
}
