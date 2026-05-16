package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/mubarok-ridho/casheer-menu-service/internal/handlers"
	"github.com/mubarok-ridho/casheer-menu-service/internal/middleware"
	"github.com/mubarok-ridho/casheer-menu-service/internal/models"
	"github.com/mubarok-ridho/casheer-menu-service/internal/repository"
	"github.com/mubarok-ridho/casheer-menu-service/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found, using environment variables")
	}
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("📦 Running database migrations...")
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Menu{},
		&models.MenuVariation{},
		&models.MenuImage{},
		&models.Order{},
		&models.OrderItem{},
		&models.Promo{},
		&models.PromoItem{},
		&models.Ingredient{},
		&models.MenuIngredient{},
		&models.VariationIngredient{},
	); err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}
	log.Println("✅ Database migration completed")

	menuRepo := repository.NewMenuRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	variationRepo := repository.NewVariationRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	promoRepo := repository.NewPromoRepository(db)
	ingRepo := repository.NewIngredientRepository(db)

	menuHandler := handlers.NewMenuHandler(menuRepo, categoryRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, ingRepo)
	variationHandler := handlers.NewVariationHandler(variationRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	promoHandler := handlers.NewPromoHandler(promoRepo)
	ingHandler := handlers.NewIngredientHandler(ingRepo)
	marginsHandler := handlers.NewMarginsHandler(db, ingRepo)

	app := fiber.New(fiber.Config{AppName: os.Getenv("APP_NAME")})
	app.Use(cors.New())
	setupRoutes(app, menuHandler, orderHandler, variationHandler, categoryHandler, promoHandler, ingHandler, marginsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	log.Printf("🚀 Menu Service starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(
	app *fiber.App,
	menuHandler *handlers.MenuHandler,
	orderHandler *handlers.OrderHandler,
	variationHandler *handlers.VariationHandler,
	categoryHandler *handlers.CategoryHandler,
	promoHandler *handlers.PromoHandler,
	ingHandler *handlers.IngredientHandler,
	marginsHandler *handlers.MarginsHandler,
) {
	public := app.Group("/public")
	public.Get("/menu/:tenantId", menuHandler.GetPublicMenu)
	public.Get("/promos/:tenantId", promoHandler.GetActive)

	api := app.Group("/api/v1", middleware.AuthMiddleware())

	menuRoutes := api.Group("/menus")
	menuRoutes.Post("/", menuHandler.Create)
	menuRoutes.Get("/", menuHandler.GetAll)
	menuRoutes.Get("/bestseller", menuHandler.GetBestSeller)
	menuRoutes.Get("/:id", menuHandler.GetByID)
	menuRoutes.Put("/:id", menuHandler.Update)
	menuRoutes.Delete("/:id", menuHandler.Delete)
	menuRoutes.Post("/:id/images", menuHandler.UploadImage)
	menuRoutes.Delete("/:id/images/:imageId", menuHandler.DeleteImage)
	menuRoutes.Patch("/:id/availability", menuHandler.ToggleAvailability)
	// Ingredients per menu
	menuRoutes.Get("/:id/ingredients", ingHandler.GetMenuIngredients)
	menuRoutes.Put("/:id/ingredients", ingHandler.SetMenuIngredients)

	variationRoutes := api.Group("/variations")
	variationRoutes.Post("/", variationHandler.Create)
	variationRoutes.Put("/:id", variationHandler.Update)
	variationRoutes.Delete("/:id", variationHandler.Delete)
	variationRoutes.Patch("/:id/stock", variationHandler.UpdateStock)
	variationRoutes.Get("/:id/ingredients", ingHandler.GetVariationIngredients)
	variationRoutes.Put("/:id/ingredients", ingHandler.SetVariationIngredients)
	variationRoutes.Get("/:id/ingredients", ingHandler.GetVariationIngredients)
	variationRoutes.Put("/:id/ingredients", ingHandler.SetVariationIngredients)

	categoryRoutes := api.Group("/categories")
	categoryRoutes.Get("/", categoryHandler.GetAll)
	categoryRoutes.Post("/", categoryHandler.Create)
	categoryRoutes.Put("/:id", categoryHandler.Update)
	categoryRoutes.Delete("/:id", categoryHandler.Delete)

	orderRoutes := api.Group("/orders")
	orderRoutes.Post("/", orderHandler.Create)
	orderRoutes.Get("/", orderHandler.GetAll)
	orderRoutes.Get("/tenant/:tenantId", orderHandler.GetByTenant)
	orderRoutes.Delete("/cleanup", orderHandler.DeleteOldOrders)
	orderRoutes.Get("/:id", orderHandler.GetByID)

	promoRoutes := api.Group("/promos")
	promoRoutes.Get("/", promoHandler.GetAll)
	promoRoutes.Post("/", promoHandler.Create)
	promoRoutes.Put("/:id", promoHandler.Update)
	promoRoutes.Delete("/:id", promoHandler.Delete)
	promoRoutes.Post("/upload-image", promoHandler.UploadImage)

	// Ingredients (bahan baku)
	ingRoutes := api.Group("/ingredients")
	ingRoutes.Get("/", ingHandler.GetAll)
	ingRoutes.Get("/low-stock", ingHandler.GetLowStock)
	ingRoutes.Post("/", ingHandler.Create)
	ingRoutes.Put("/:id", ingHandler.Update)
	ingRoutes.Delete("/:id", ingHandler.Delete)

	// Margins
	marginsRoutes := api.Group("/margins")
	marginsRoutes.Get("/", marginsHandler.GetSummary)
	marginsRoutes.Get("/menu-breakdown", marginsHandler.GetMenuBreakdown)
}
