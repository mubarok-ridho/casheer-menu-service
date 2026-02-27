package handlers

import (
	"casheer-menu-service/internal/models"
	"casheer-menu-service/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type MenuHandler struct {
	DB *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{DB: db}
}

// Create menu dengan variasi
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
		Variations  []VariationInput `json:"variations"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Mulai transaction
	tx := h.DB.Begin()

	menu := models.Menu{
		TenantID:    tenantID,
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		Description: input.Description,
		BasePrice:   input.BasePrice,
		IsAvailable: true,
	}

	if err := tx.Create(&menu).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create variations
	for _, v := range input.Variations {
		variation := models.MenuVariation{
			MenuID: menu.ID,
			Name:   v.Name,
			Option: v.Option,
			Price:  v.Price,
			Stock:  v.Stock,
		}
		if err := tx.Create(&variation).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Handle image upload
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for i, file := range files {
			// Upload ke Cloudinary
			imageURL, publicID, err := utils.UploadToCloudinary(file, "menu")
			if err != nil {
				continue
			}

			menuImage := models.MenuImage{
				MenuID:    menu.ID,
				ImageURL:  imageURL,
				PublicID:  publicID,
				IsPrimary: i == 0, // Jadikan gambar pertama sebagai primary
			}
			tx.Create(&menuImage)
		}
	}

	tx.Commit()

	return c.Status(201).JSON(menu)
}

// Get best seller
func (h *MenuHandler) GetBestSeller(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)

	var menus []models.Menu

	// Query menu dengan jumlah order terbanyak (misal 30 hari terakhir)
	h.DB.Table("menus").
		Select("menus.*, COUNT(orders.id) as order_count").
		Joins("left join orders on orders.menu_id = menus.id").
		Where("menus.tenant_id = ? AND orders.created_at > NOW() - INTERVAL '30 days'", tenantID).
		Group("menus.id").
		Order("order_count DESC").
		Limit(10).
		Preload("Images").
		Find(&menus)

	return c.JSON(menus)
}
