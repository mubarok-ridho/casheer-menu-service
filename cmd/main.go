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
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found, using environment variables")
	}

	// Initialize Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Auto Migrate
	log.Println("📦 Running database migrations...")
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Menu{},
		&models.MenuVariation{},
		&models.MenuImage{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}
	log.Println("✅ Database migration completed")

	// Initialize repositories
	menuRepo := repository.NewMenuRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	variationRepo := repository.NewVariationRepository(db)

	// Initialize handlers
	menuHandler := handlers.NewMenuHandler(menuRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	variationHandler := handlers.NewVariationHandler(variationRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: os.Getenv("APP_NAME"),
	})

	app.Use(cors.New())

	// Setup routes
	setupRoutes(app, menuHandler, orderHandler, variationHandler)

	// Start server
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
) {
	// Public routes (if any)

	// Protected routes (require JWT)
	api := app.Group("/api/v1", middleware.AuthMiddleware())

	// Menu routes
	menuRoutes := api.Group("/menus")
	menuRoutes.Post("/", menuHandler.Create)
	menuRoutes.Get("/", menuHandler.GetAll)
	menuRoutes.Get("/bestseller", menuHandler.GetBestSeller)
	menuRoutes.Get("/:id", menuHandler.GetByID)
	menuRoutes.Put("/:id", menuHandler.Update)
	menuRoutes.Delete("/:id", menuHandler.Delete)
	menuRoutes.Post("/:id/images", menuHandler.UploadImage)
	menuRoutes.Delete("/:id/images/:imageId", menuHandler.DeleteImage)

	// Variation routes
	variationRoutes := api.Group("/variations")
	variationRoutes.Post("/", variationHandler.Create)
	variationRoutes.Put("/:id", variationHandler.Update)
	variationRoutes.Delete("/:id", variationHandler.Delete)
	variationRoutes.Patch("/:id/stock", variationHandler.UpdateStock)

	// Order routes
	orderRoutes := api.Group("/orders")
	orderRoutes.Post("/", orderHandler.Create)
	orderRoutes.Get("/", orderHandler.GetAll)
	orderRoutes.Get("/:id", orderHandler.GetByID)
	orderRoutes.Get("/tenant/:tenantId", orderHandler.GetByTenant)
}
