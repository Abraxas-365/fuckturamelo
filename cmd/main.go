package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Abraxas-365/craftable/errx/errxfiber"
	"github.com/Abraxas-365/fuckturamelo/providers/providersapi"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := initDatabase(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Fiber app with errx error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: errxfiber.FiberErrorHandler(),
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup API routes
	setupRoutes(app, db)

	// Global health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Shutting down server...")
		_ = app.ShutdownWithTimeout(30 * time.Second)
	}()

	// Start server
	log.Printf("🚀 Server starting on port %s", config.Server.Port)
	if err := app.Listen(":" + config.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// setupRoutes configures all API routes
func setupRoutes(app *fiber.App, db *sqlx.DB) {
	// API v1 group
	api := app.Group("/api/v1")

	// Initialize Providers API and setup routes
	providersAPI, err := providersapi.New(providersapi.Config{DB: db})
	if err != nil {
		log.Fatalf("Failed to initialize providers API: %v", err)
	}

	// Setup providers routes under /api/v1/providers
	providersGroup := api.Group("/providers")
	providersAPI.SetupRoutes(providersGroup)

	// You can add more domains here in the same way:
	// invoicesAPI, err := invoicesapi.New(invoicesapi.Config{DB: db})
	// invoicesGroup := api.Group("/invoices")
	// invoicesAPI.SetupRoutes(invoicesGroup)
}

// loadConfig and initDatabase functions (same as before)
func loadConfig() (*AppConfig, error) {
	// Implementation here...
	return nil, nil
}

func initDatabase(config *AppConfig) (*sqlx.DB, error) {
	// Implementation here...
	return nil, nil
}

type AppConfig struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	Database struct {
		URL string `json:"url"`
	} `json:"database"`
}
