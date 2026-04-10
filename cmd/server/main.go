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
	custommw "github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/pkg/config"
	"github.com/institutoitinerante/pulse-service/internal/pkg/wshub"
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

	// BKL-150: inicializar hub WebSocket e iniciar loop em background
	wsHub := wshub.New()
	go wsHub.Run()

	// Wire up layers — events ingestion
	eventRepo := repository.NewEventRepository(pool)
	eventSvc := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventSvc, cfg.ServiceToken).WithHub(wsHub)

	// Wire up layers — retention metrics
	retentionRepo := repository.NewRetentionRepository(pool)
	retentionSvc := service.NewRetentionService(retentionRepo)
	retentionHandler := handler.NewRetentionHandler(retentionSvc, cfg.ServiceToken)

	// Wire up layers — projects (BKL-099)
	projectRepo := repository.NewProjectRepository(pool)
	projectHandler := handler.NewProjectHandler(projectRepo, cfg.ServiceToken)

	// Wire up layers — cost tracking
	costRepo := repository.NewCostRepository(pool)
	costHandler := handler.NewCostHandler(costRepo, cfg.ServiceToken)

	// Wire up layers — CSAT / NPS
	csatRepo := repository.NewCSATRepository(pool)
	csatHandler := handler.NewCSATHandler(csatRepo, cfg.ServiceToken)

	// Wire up layers — health scores (BKL-801)
	healthScoreRepo := repository.NewHealthScoreRepository(pool)
	healthScoreSvc := service.NewHealthScoreService(healthScoreRepo)
	healthScoreHandler := handler.NewHealthScoreHandler(healthScoreSvc, cfg.ServiceToken)

	// BKL-150: WebSocket handler
	wsHandler := handler.NewWSHandler(wsHub, cfg.ServiceToken)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "pulse-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(custommw.SecurityHeaders())
	app.Use(custommw.RateLimit(redisClient)) // BKL-463: Redis sliding window (multi-réplica safe)
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency}\n",
	}))

	// CORS — allow all origins for SDK access on /flags
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

	// Register retention metrics routes
	retentionHandler.RegisterRoutes(app)

	// Register project routes (BKL-099)
	projectHandler.RegisterRoutes(app)

	// Register cost tracking routes
	costHandler.RegisterRoutes(app)

	// Register CSAT / NPS routes
	csatHandler.RegisterRoutes(app)

	// Register health score routes (BKL-801)
	healthScoreHandler.RegisterRoutes(app)

	// BKL-150: Register WebSocket routes
	wsHandler.RegisterRoutes(app)

	// BKL-344: Servir arquivos estáticos do widget CSAT embeddable.
	// Qualquer site pode embedar com: <script src=".../static/csat/widget.js" data-survey-id="..."></script>
	app.Static("/static", "./frontend/public")

	// Start uptime monitor background checker
	monitorSvc.StartBackgroundChecker()

	// BKL-801: Start daily health score calculation job
	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()
	service.StartHealthScoreJob(jobCtx, healthScoreSvc)

	// BKL-814: Churn alert — email diário para admin com tenants score < 40
	churnAlertScheduler := service.NewChurnAlertScheduler(healthScoreRepo, 24*time.Hour)
	churnAlertScheduler.Start(jobCtx)

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
