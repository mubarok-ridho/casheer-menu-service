package handlers

import (
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
)

type CategoryHandler struct {
	Repo *repository.CategoryRepository
}

func NewCategoryHandler(repo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{Repo: repo}
}

func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	categories, err := h.Repo.GetByTenant(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": categories})
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	var category models.Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	category.TenantID = tenantID
	if err := h.Repo.Create(&category); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": category})
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	category, err := h.Repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}
	if err := c.BodyParser(category); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	category.TenantID = tenantID
	if err := h.Repo.Update(category); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": category})
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.Repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Category deleted"})
}
