package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"gorm.io/gorm"
)

type MarginsHandler struct {
	DB      *gorm.DB
	ingRepo *repository.IngredientRepository
}

func NewMarginsHandler(db *gorm.DB, ingRepo *repository.IngredientRepository) *MarginsHandler {
	return &MarginsHandler{DB: db, ingRepo: ingRepo}
}

type MenuMargin struct {
	MenuID   uint    `json:"menu_id"`
	MenuName string  `json:"menu_name"`
	QtySold  int64   `json:"qty_sold"`
	Revenue  float64 `json:"revenue"`
	COGS     float64 `json:"cogs"`
	Profit   float64 `json:"profit"`
	Margin   float64 `json:"margin_pct"`
}

type orderItemRow struct {
	MenuID      uint
	VariationID *uint
	QtySold     int64
	Revenue     float64
}

// calcCOGS — hitung COGS dari order_items rows
// ingByMenu: menu_id -> []MenuIngredient
// ingByVar:  variation_id -> []VariationIngredient
func calcCOGS(
	rows []orderItemRow,
	ingByMenu map[uint][]models.MenuIngredient,
	ingByVar map[uint][]models.VariationIngredient,
) float64 {
	var total float64
	for _, row := range rows {
		// Base ingredients
		for _, mi := range ingByMenu[row.MenuID] {
			total += mi.Amount * float64(row.QtySold) * mi.Ingredient.CostPerUnit
		}
		// Variation delta ingredients
		if row.VariationID != nil {
			for _, vi := range ingByVar[*row.VariationID] {
				total += vi.Amount * float64(row.QtySold) * vi.Ingredient.CostPerUnit
			}
		}
	}
	return total
}

// fetchIngMaps — ambil semua ingredient maps yang dibutuhkan
func (h *MarginsHandler) fetchIngMaps(menuIDs []uint, varIDs []uint) (
	map[uint][]models.MenuIngredient,
	map[uint][]models.VariationIngredient,
	error,
) {
	ingByMenu := make(map[uint][]models.MenuIngredient)
	ingByVar := make(map[uint][]models.VariationIngredient)

	if len(menuIDs) > 0 {
		menuIngs, err := h.ingRepo.GetIngredientsByMenuIDs(menuIDs)
		if err != nil {
			return nil, nil, err
		}
		for _, mi := range menuIngs {
			ingByMenu[mi.MenuID] = append(ingByMenu[mi.MenuID], mi)
		}
	}

	if len(varIDs) > 0 {
		varIngs, err := h.ingRepo.GetVariationIngredientsByIDs(varIDs)
		if err != nil {
			return nil, nil, err
		}
		for _, vi := range varIngs {
			ingByVar[vi.VariationID] = append(ingByVar[vi.VariationID], vi)
		}
	}

	return ingByMenu, ingByVar, nil
}

// GET /margins?days=30
func (h *MarginsHandler) GetSummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	days := c.QueryInt("days", 30)
	startDate := time.Now().AddDate(0, 0, -days)

	// Total revenue
	var totalRevenue float64
	h.DB.Model(&models.Order{}).
		Where("tenant_id = ? AND created_at >= ? AND payment_status = 'paid'", tenantID, startDate).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

	// Order items dengan variation info
	type Row struct {
		MenuID      uint
		VariationID *uint
		QtySold     int64
		Revenue     float64
	}
	var rows []Row
	h.DB.Table("order_items").
		Select("order_items.menu_id, order_items.variation_id, SUM(order_items.quantity) as qty_sold, SUM(order_items.subtotal) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.tenant_id = ? AND orders.created_at >= ? AND orders.payment_status = 'paid'", tenantID, startDate).
		Group("order_items.menu_id, order_items.variation_id").
		Scan(&rows)

	if len(rows) == 0 {
		return c.JSON(fiber.Map{
			"total_revenue": totalRevenue, "total_cogs": 0,
			"net_profit": totalRevenue, "margin_pct": 100, "days": days,
		})
	}

	// Kumpulkan unique IDs
	menuIDSet := make(map[uint]bool)
	varIDSet := make(map[uint]bool)
	for _, r := range rows {
		menuIDSet[r.MenuID] = true
		if r.VariationID != nil {
			varIDSet[*r.VariationID] = true
		}
	}
	menuIDs := mapKeys(menuIDSet)
	varIDs := mapKeys(varIDSet)

	ingByMenu, ingByVar, err := h.fetchIngMaps(menuIDs, varIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Konversi ke orderItemRow
	itemRows := make([]orderItemRow, len(rows))
	for i, r := range rows {
		itemRows[i] = orderItemRow{MenuID: r.MenuID, VariationID: r.VariationID, QtySold: r.QtySold, Revenue: r.Revenue}
	}

	totalCOGS := calcCOGS(itemRows, ingByMenu, ingByVar)

	marginPct := 0.0
	if totalRevenue > 0 {
		marginPct = ((totalRevenue - totalCOGS) / totalRevenue) * 100
	}

	return c.JSON(fiber.Map{
		"total_revenue": totalRevenue,
		"total_cogs":    totalCOGS,
		"net_profit":    totalRevenue - totalCOGS,
		"margin_pct":    marginPct,
		"days":          days,
	})
}

// GET /margins/menu-breakdown?days=30
func (h *MarginsHandler) GetMenuBreakdown(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	days := c.QueryInt("days", 30)
	startDate := time.Now().AddDate(0, 0, -days)

	// Order items dengan variation info, group by menu+variation
	type Row struct {
		MenuID      uint
		MenuName    string
		VariationID *uint
		QtySold     int64
		Revenue     float64
	}
	var rows []Row
	h.DB.Table("order_items").
		Select("order_items.menu_id, menus.name as menu_name, order_items.variation_id, SUM(order_items.quantity) as qty_sold, SUM(order_items.subtotal) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN menus ON menus.id = order_items.menu_id").
		Where("orders.tenant_id = ? AND orders.created_at >= ? AND orders.payment_status = 'paid'", tenantID, startDate).
		Group("order_items.menu_id, menus.name, order_items.variation_id").
		Order("revenue desc").
		Scan(&rows)

	if len(rows) == 0 {
		return c.JSON([]MenuMargin{})
	}

	// Kumpulkan unique IDs
	menuIDSet := make(map[uint]bool)
	varIDSet := make(map[uint]bool)
	for _, r := range rows {
		menuIDSet[r.MenuID] = true
		if r.VariationID != nil {
			varIDSet[*r.VariationID] = true
		}
	}

	ingByMenu, ingByVar, err := h.fetchIngMaps(mapKeys(menuIDSet), mapKeys(varIDSet))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Aggregate per menu (gabungkan semua variasi dalam 1 menu)
	type menuAgg struct {
		name    string
		qtySold int64
		revenue float64
		cogs    float64
	}
	aggMap := make(map[uint]*menuAgg)

	for _, r := range rows {
		if _, ok := aggMap[r.MenuID]; !ok {
			aggMap[r.MenuID] = &menuAgg{name: r.MenuName}
		}
		agg := aggMap[r.MenuID]
		agg.qtySold += r.QtySold
		agg.revenue += r.Revenue

		// COGS untuk row ini
		rowCOGS := calcCOGS([]orderItemRow{
			{MenuID: r.MenuID, VariationID: r.VariationID, QtySold: r.QtySold, Revenue: r.Revenue},
		}, ingByMenu, ingByVar)
		agg.cogs += rowCOGS
	}

	result := make([]MenuMargin, 0, len(aggMap))
	for menuID, agg := range aggMap {
		profit := agg.revenue - agg.cogs
		marginPct := 0.0
		if agg.revenue > 0 {
			marginPct = (profit / agg.revenue) * 100
		}
		result = append(result, MenuMargin{
			MenuID:   menuID,
			MenuName: agg.name,
			QtySold:  agg.qtySold,
			Revenue:  agg.revenue,
			COGS:     agg.cogs,
			Profit:   profit,
			Margin:   marginPct,
		})
	}

	// Sort by revenue desc
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Revenue > result[i].Revenue {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return c.JSON(result)
}

func mapKeys(m map[uint]bool) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
