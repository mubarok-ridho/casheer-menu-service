package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"github.com/mubarok-ridho/casheer-menu-service/internal/utils"
)

type PromoHandler struct {
	repo *repository.PromoRepository
}

func NewPromoHandler(repo *repository.PromoRepository) *PromoHandler {
	return &PromoHandler{repo: repo}
}

type itemInput struct {
	MenuID      uint  `json:"menu_id"`
	VariationID *uint `json:"variation_id"`
	// AddonMode: "" = tidak ada variasi, "fixed" = variasi terkunci, "dynamic" = customer pilih saat checkout
	AddonMode string `json:"addon_mode"`
	Quantity  int    `json:"quantity"`
}

type promoInput struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PromoPrice  float64     `json:"promo_price"`
	ImageURL    string      `json:"image_url"`
	IsActive    bool        `json:"is_active"`
	StartAt     string      `json:"start_at"`
	EndAt       *string     `json:"end_at"`
	StartTime   string      `json:"start_time"` // "08:00"
	EndTime     string      `json:"end_time"`   // "10:00"
	Items       []itemInput `json:"items"`
}

func (h *PromoHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	promos, err := h.repo.GetByTenant(tenantID, false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(promos)
}

func (h *PromoHandler) GetActive(c *fiber.Ctx) error {
	tenantID, err := strconv.ParseUint(c.Params("tenantId"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid tenant ID"})
	}
	promos, err := h.repo.GetByTenant(uint(tenantID), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Filter berdasarkan jam aktif (WIB)
	now := time.Now().In(time.FixedZone("WIB", 7*3600))
	currentTime := now.Format("15:04")

	type PromoWithStatus struct {
		models.Promo
		UnavailableReason string `json:"unavailable_reason"` // kosong = tersedia
	}

	var result []PromoWithStatus
	for _, p := range promos {
		pw := PromoWithStatus{Promo: p}

		// Cek jam aktif
		if p.StartTime != "" && p.EndTime != "" {
			if currentTime < p.StartTime || currentTime > p.EndTime {
				pw.UnavailableReason = "Di luar jam promo (" + p.StartTime + "–" + p.EndTime + ")"
				result = append(result, pw)
				continue
			}
		}

		// Cek menu habis dalam bundle
		for _, item := range p.Items {
			if item.Menu.ID != 0 && !item.Menu.IsAvailable {
				pw.UnavailableReason = "Menu \"" + item.Menu.Name + "\" sedang habis"
				break
			}
		}

		result = append(result, pw)
	}

	if result == nil {
		result = []PromoWithStatus{}
	}
	return c.JSON(result)
}

func (h *PromoHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	var input promoInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" || input.PromoPrice <= 0 || len(input.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Name, promo_price, dan items wajib diisi"})
	}

	startAt := time.Now()
	if input.StartAt != "" {
		if t, err := time.Parse("2006-01-02", input.StartAt); err == nil {
			startAt = t
		}
	}

	promo := &models.Promo{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		PromoPrice:  input.PromoPrice,
		ImageURL:    input.ImageURL,
		IsActive:    input.IsActive,
		StartAt:     startAt,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
	}
	if input.EndAt != nil && *input.EndAt != "" {
		if t, err := time.Parse("2006-01-02", *input.EndAt); err == nil {
			promo.EndAt = &t
		}
	}

	for _, item := range input.Items {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		// Validasi: jika fixed, VariationID harus ada
		// Jika dynamic, VariationID boleh nil (dipilih saat checkout)
		addonMode := item.AddonMode
		if addonMode != "" && addonMode != models.AddonModeFixed && addonMode != models.AddonModeDynamic {
			return c.Status(400).JSON(fiber.Map{"error": "addon_mode harus 'fixed', 'dynamic', atau kosong"})
		}
		if addonMode == models.AddonModeFixed && item.VariationID == nil {
			return c.Status(400).JSON(fiber.Map{"error": "addon_mode 'fixed' membutuhkan variation_id"})
		}
		promo.Items = append(promo.Items, models.PromoItem{
			MenuID:      item.MenuID,
			VariationID: item.VariationID,
			AddonMode:   addonMode,
			Quantity:    qty,
		})
	}
	if err := h.repo.Create(promo); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(promo)
}

func (h *PromoHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid promo ID"})
	}

	var input promoInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	promo, err := h.repo.GetByID(uint(id), tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promo not found"})
	}

	promo.Name = input.Name
	promo.Description = input.Description
	promo.PromoPrice = input.PromoPrice
	promo.ImageURL = input.ImageURL
	promo.IsActive = input.IsActive
	promo.StartTime = input.StartTime
	promo.EndTime = input.EndTime

	if input.StartAt != "" {
		if t, err := time.Parse("2006-01-02", input.StartAt); err == nil {
			promo.StartAt = t
		}
	}
	if input.EndAt != nil && *input.EndAt != "" {
		if t, err := time.Parse("2006-01-02", *input.EndAt); err == nil {
			promo.EndAt = &t
		}
	} else {
		promo.EndAt = nil
	}

	if err := h.repo.Update(promo); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var newItems []models.PromoItem
	for _, item := range input.Items {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		newItems = append(newItems, models.PromoItem{
			MenuID:      item.MenuID,
			VariationID: item.VariationID,
			Quantity:    qty,
		})
	}
	h.repo.ReplaceItems(promo.ID, newItems)

	updated, _ := h.repo.GetByID(promo.ID, tenantID)
	return c.JSON(updated)
}

func (h *PromoHandler) UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No image uploaded"})
	}
	if file.Size > 2*1024*1024 { // max 2MB
		return c.Status(400).JSON(fiber.Map{"error": "Ukuran file max 2MB"})
	}
	imageURL, _, err := utils.UploadToCloudinary(file, "promos")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"image_url": imageURL})
}

func (h *PromoHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid promo ID"})
	}
	if err := h.repo.Delete(uint(id), tenantID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Promo dihapus"})
}
