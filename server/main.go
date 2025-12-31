package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	database "github.com/immatheus/gitback/databases"
	"github.com/immatheus/gitback/handlers"
	"github.com/immatheus/gitback/middleware"
	"github.com/immatheus/gitback/storage"
)

func main() {
	godotenv.Load() // only for dev, gcp injects this for us

	if err := database.Init(os.Getenv("DATABASE_URL")); err != nil {
		log.Printf("ERROR: Database initialization failed: %v", err)
		log.Printf("Continuing without database - data will not be persisted")
	} else {
		log.Printf("Database initialized successfully")
	}
	defer database.Close()

	if err := storage.Init(); err != nil {
		log.Printf("WARNING: Storage cache initialization failed: %v", err)
		log.Printf("Continuing without cache - requests will be slower")
	} else {
		log.Printf("GCP Storage cache initialized successfully")
	}
	defer storage.Close()

	app := fiber.New(fiber.Config{
		AppName:      "GitBack v2.0.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Unhandled error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Security and recovery middleware
	app.Use(middleware.RecoveryMiddleware())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.InputValidation())

	// Rate limiting: 100 requests per minute per IP for analyze endpoint
	analyzeRateLimit := middleware.CreateRateLimiter(middleware.RateLimitConfig{
		Max:        100,
		Expiration: time.Minute,
	})

	// General rate limiting: 1000 requests per minute per IP
	generalRateLimit := middleware.CreateRateLimiter(middleware.RateLimitConfig{
		Max:        1000,
		Expiration: time.Minute,
	})

	// Compression and logging
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path} - ${ip}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"version": "2.0.0",
			"time":    time.Now().Unix(),
		})
	})

	// API routes with rate limiting
	api := app.Group("/api", generalRateLimit)
	api.Post("/analyze", analyzeRateLimit, handlers.AnalyzeRepo)
	api.Get("/top-repos", getTopRepos)

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "GitBack API",
			"version": "1.0.0",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting GitBack server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// getTopRepos handles the top repositories endpoint
func getTopRepos(c *fiber.Ctx) error {
	repos, err := database.GetTopRepos()
	if err != nil {
		log.Printf("Failed to get top repos: %v", err)
		return middleware.InternalError(c, "Failed to fetch top repositories")
	}

	return c.JSON(fiber.Map{
		"repos": repos,
	})
}
