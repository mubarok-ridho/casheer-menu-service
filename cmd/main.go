package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"kasir-menu-service/internal/handlers"
	"kasir-menu-service/internal/models"
	"kasir-menu-service/internal/repository"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto migrate
	db.AutoMigrate(
		&models.Category{},
		&models.Menu{},
		&models.MenuVariation{},
		&models.MenuImage{},
		&models.Order{},
	)

	// Initialize repositories
	menuRepo := repository.NewMenuRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Initialize handlers
	menuHandler := handlers.NewMenuHandler(menuRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	// Fiber app
	app := fiber.New()
	app.Use(cors.New())

	// Routes
	api := app.Group("/api/v1")

	// Menu routes
	menuRoutes := api.Group("/menus")
	menuRoutes.Post("/", menuHandler.Create)
	menuRoutes.Get("/", menuHandler.GetAll)
	menuRoutes.Get("/:id", menuHandler.GetByID)
	menuRoutes.Put("/:id", menuHandler.Update)
	menuRoutes.Delete("/:id", menuHandler.Delete)
	menuRoutes.Post("/:id/images", menuHandler.UploadImage)

	// Order routes
	orderRoutes := api.Group("/orders")
	orderRoutes.Post("/", orderHandler.Create)
	orderRoutes.Get("/bestseller", orderHandler.GetBestSeller)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	log.Fatal(app.Listen(":" + port))
}
