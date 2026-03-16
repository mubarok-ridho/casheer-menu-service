package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"github.com/mubarok-ridho/casheer-menu-service/internal/utils"
)

type MenuHandler struct {
	repo         *repository.MenuRepository
	categoryRepo *repository.CategoryRepository
}

func NewMenuHandler(repo *repository.MenuRepository, categoryRepo *repository.CategoryRepository) *MenuHandler {
	return &MenuHandler{repo: repo, categoryRepo: categoryRepo}
}

func (h *MenuHandler) ToggleAvailability(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	var input struct {
		IsAvailable bool `json:"is_available"`
	}
	c.BodyParser(&input)
	menu, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Menu not found"})
	}
	menu.IsAvailable = input.IsAvailable
	if err := h.repo.Update(menu); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(menu)
}

func (h *MenuHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	type VariationInput struct {
		Name   string  `json:"name"`
		Option string  `json:"option"`
		Price  float64 `json:"price"`
		Stock  int     `json:"stock"`
	}
	var input struct {
		CategoryID  uint             `json:"category_id"`
		Name        string           `json:"name"`
		Description string           `json:"description"`
		BasePrice   float64          `json:"base_price"`
		PrepTime    int              `json:"prep_time"`
		Variations  []VariationInput `json:"variations"`
	}
	if c.Get("Content-Type") == "application/json" {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}
	} else {
		input.Name = c.FormValue("name")
		input.Description = c.FormValue("description")
		if v, err := strconv.ParseUint(c.FormValue("category_id"), 10, 32); err == nil {
			input.CategoryID = uint(v)
		}
		if v, err := strconv.ParseFloat(c.FormValue("base_price"), 64); err == nil {
			input.BasePrice = v
		}
		if v, err := strconv.Atoi(c.FormValue("prep_time")); err == nil {
			input.PrepTime = v
		}
		if variations := c.FormValue("variations"); variations != "" {
			json.Unmarshal([]byte(variations), &input.Variations)
		}
	}
	if input.Name == "" || input.BasePrice <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Name and base price are required"})
	}
	menu := &models.Menu{
		TenantID:    tenantID,
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		Description: input.Description,
		BasePrice:   input.BasePrice,
		PrepTime:    input.PrepTime,
		IsAvailable: true,
	}
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for i, file := range files {
			imageURL, publicID, err := utils.UploadToCloudinary(file, "menus")
			if err != nil {
				continue
			}
			menu.Images = append(menu.Images, models.MenuImage{
				ImageURL:  imageURL,
				PublicID:  publicID,
				IsPrimary: i == 0,
				SortOrder: i,
			})
		}
	}
	for _, v := range input.Variations {
		menu.Variations = append(menu.Variations, models.MenuVariation{
			Name:   v.Name,
			Option: v.Option,
			Price:  v.Price,
			Stock:  v.Stock,
		})
	}
	if err := h.repo.Create(menu); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(menu)
}

func (h *MenuHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	categoryID := c.QueryInt("category_id", 0)
	search := c.Query("search")
	menus, err := h.repo.GetByTenant(tenantID, categoryID, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(menus)
}

func (h *MenuHandler) GetByID(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	menu, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Menu not found"})
	}
	return c.JSON(menu)
}

func (h *MenuHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	var input struct {
		CategoryID  uint    `json:"category_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		BasePrice   float64 `json:"base_price"`
		PrepTime    int     `json:"prep_time"`
		IsAvailable bool    `json:"is_available"`
	}
	if c.Get("Content-Type") == "application/json" {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}
	} else {
		input.Name = c.FormValue("name")
		input.Description = c.FormValue("description")
		if v, err := strconv.ParseUint(c.FormValue("category_id"), 10, 32); err == nil {
			input.CategoryID = uint(v)
		}
		if v, err := strconv.ParseFloat(c.FormValue("base_price"), 64); err == nil {
			input.BasePrice = v
		}
		if v, err := strconv.Atoi(c.FormValue("prep_time")); err == nil {
			input.PrepTime = v
		}
		input.IsAvailable = c.FormValue("is_available") == "true"
	}
	menu, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Menu not found"})
	}
	menu.CategoryID = input.CategoryID
	menu.Name = input.Name
	menu.Description = input.Description
	menu.BasePrice = input.BasePrice
	menu.PrepTime = input.PrepTime
	menu.IsAvailable = input.IsAvailable
	if err := h.repo.Update(menu); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	type VariationInput struct {
		Name   string  `json:"name"`
		Option string  `json:"option"`
		Price  float64 `json:"price"`
	}
	var variations []VariationInput
	if v := c.FormValue("variations"); v != "" {
		json.Unmarshal([]byte(v), &variations)
	}
	h.repo.DB.Where("menu_id = ?", menu.ID).Delete(&models.MenuVariation{})
	for _, v := range variations {
		h.repo.DB.Create(&models.MenuVariation{
			MenuID: menu.ID,
			Name:   v.Name,
			Option: v.Option,
			Price:  v.Price,
		})
	}
	return c.JSON(menu)
}

func (h *MenuHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Menu deleted successfully"})
}

func (h *MenuHandler) UploadImage(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid menu ID"})
	}
	menu, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Menu not found"})
	}
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No image file uploaded"})
	}
	if file.Size > 2*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "File too large. Max size 2MB"})
	}
	imageURL, publicID, err := utils.UploadToCloudinary(file, "menus")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to upload image: " + err.Error()})
	}
	isPrimary := len(menu.Images) == 0
	menuImage := models.MenuImage{
		MenuID:    menu.ID,
		ImageURL:  imageURL,
		PublicID:  publicID,
		IsPrimary: isPrimary,
		SortOrder: len(menu.Images),
	}
	if err := h.repo.AddImage(&menuImage); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(menuImage)
}

func (h *MenuHandler) DeleteImage(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	menuID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	imageID, _ := strconv.ParseUint(c.Params("imageId"), 10, 32)
	if err := h.repo.DeleteImage(uint(imageID), uint(menuID), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Image deleted successfully"})
}

func (h *MenuHandler) GetBestSeller(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	limit := c.QueryInt("limit", 10)
	days := c.QueryInt("days", 30)
	menus, err := h.repo.GetBestSeller(tenantID, limit, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(menus)
}

func (h *MenuHandler) GetPublicMenu(c *fiber.Ctx) error {
	tenantID, err := strconv.ParseUint(c.Params("tenantId"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid tenant ID"})
	}
	categoryID := c.QueryInt("category_id", 0)
	menus, err := h.repo.GetPublicMenus(uint(tenantID), categoryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	categories, err := h.categoryRepo.GetByTenant(uint(tenantID))
	if err != nil {
		categories = []models.Category{}
	}
	return c.JSON(fiber.Map{
		"menus":      menus,
		"categories": categories,
		"tenant_id":  tenantID,
	})
}
