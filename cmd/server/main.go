package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/institutoitinerante/pulse-service/internal/handler"
	"github.com/institutoitinerante/pulse-service/internal/pkg/config"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/institutoitinerante/pulse-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.ServiceToken == "" {
		log.Fatal("SERVICE_TOKEN environment variable is required")
	}

	// Connect to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping PostgreSQL: %v", err)
	}
	log.Println("connected to PostgreSQL")

	// Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		log.Fatalf("failed to ping Redis: %v", err)
	}
	log.Println("connected to Redis")

	// Wire up layers — feature flags
	flagRepo := repository.NewFlagRepository(pool)
	flagSvc := service.NewFlagService(flagRepo, redisClient)
	flagHandler := handler.NewFlagHandler(flagSvc, cfg.ServiceToken)

	// Wire up layers — uptime monitors
	monitorRepo := repository.NewMonitorRepository(pool)
	monitorSvc := service.NewMonitorService(monitorRepo)
	monitorHandler := handler.NewMonitorHandler(monitorSvc, cfg.ServiceToken)

	// Wire up layers — events ingestion
	eventRepo := repository.NewEventRepository(pool)
	eventSvc := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventSvc, cfg.ServiceToken)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "pulse-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency}\n",
	}))

	// CORS — allow all origins for SDK access on /evaluate
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, X-Service-Token",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "pulse-service",
		})
	})

	// Register feature flag routes
	flagHandler.RegisterRoutes(app)

	// Register uptime monitor routes
	monitorHandler.RegisterRoutes(app)

	// Register events routes
	eventHandler.RegisterRoutes(app)

	// Start uptime monitor background checker
	monitorSvc.StartBackgroundChecker()

	// Start server in background
	addr := fmt.Sprintf(":%s", cfg.Port)
	go func() {
		log.Printf("pulse-service listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down pulse-service...")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	_ = shutCtx

	if err := app.Shutdown(); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("pulse-service stopped")
}
