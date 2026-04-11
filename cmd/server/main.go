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
	"github.com/institutoitinerante/pulse-service/internal/pkg/flagscheduler"
	"github.com/institutoitinerante/pulse-service/internal/pkg/retentionpurge"
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
	eventHandler := handler.NewEventHandler(eventSvc, cfg.ServiceToken).WithHub(wsHub).WithRedis(redisClient) // BKL-1077: quota enforcement

	// Wire up layers — retention metrics
	retentionRepo := repository.NewRetentionRepository(pool)
	retentionSvc := service.NewRetentionService(retentionRepo)
	retentionHandler := handler.NewRetentionHandler(retentionSvc, cfg.ServiceToken)

	// Wire up layers — projects (BKL-099)
	projectRepo := repository.NewProjectRepository(pool)
	projectHandler := handler.NewProjectHandler(projectRepo, cfg.ServiceToken)

	// Wire up layers — environments (BKL-1075)
	environmentRepo := repository.NewEnvironmentRepository(pool)
	environmentHandler := handler.NewEnvironmentHandler(environmentRepo, cfg.ServiceToken)

	// Wire up layers — flag audit log (BKL-1074)
	flagEvalRepo := repository.NewFlagEvaluationRepository(pool)
	flagAuditHandler := handler.NewFlagAuditHandler(flagEvalRepo, cfg.ServiceToken)

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

	// BKL-1053: Prometheus metrics + instrumentação (antes de tudo para capturar todas as rotas)
	metricsHandler := handler.NewMetricsHandler()
	metricsHandler.RegisterRoutes(app)

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

	// Register environment routes (BKL-1075)
	environmentHandler.RegisterRoutes(app)

	// Register flag audit log routes (BKL-1074)
	flagAuditHandler.RegisterRoutes(app)

	// Register user data export/delete (BKL-1036 — LGPD portabilidade)
	userExportHandler := handler.NewUserExportHandler(pool, cfg.ServiceToken)
	userExportHandler.RegisterRoutes(app)

	// Register cost tracking routes
	costHandler.RegisterRoutes(app)

	// Register CSAT / NPS routes
	csatHandler.RegisterRoutes(app)

	// Register health score routes (BKL-801)
	healthScoreHandler.RegisterRoutes(app)

	// BKL-1044: Métricas de negócio vinculadas a flags (lift = treatment vs. control)
	flagMetricRepo := repository.NewFlagMetricRepository(pool)
	flagMetricHandler := handler.NewFlagMetricHandler(flagMetricRepo, cfg.ServiceToken)
	flagMetricHandler.RegisterRoutes(app)

	// BKL-1079: Approval workflow para feature flags
	flagApprovalRepo := repository.NewFlagApprovalRepository(pool)
	flagApprovalHandler := handler.NewFlagApprovalHandler(flagApprovalRepo, cfg.ServiceToken)
	flagApprovalHandler.RegisterRoutes(app)

	// BKL-1080: Schema registry de eventos
	eventSchemaRepo := repository.NewEventSchemaRepository(pool)
	eventSchemaHandler := handler.NewEventSchemaHandler(eventSchemaRepo, cfg.ServiceToken)
	eventSchemaHandler.RegisterRoutes(app)

	// BKL-1043 + BKL-1073: Análise estatística de experimentos A/B (chi-squared, p-value, CI)
	experimentHandler := handler.NewExperimentHandler()
	experimentHandler.RegisterRoutes(app)

	// BKL-1086: Upgrade recommender inteligente — recomenda plano baseado em comportamento
	upgradeHandler := handler.NewUpgradeHandler(pool, cfg.ServiceToken)
	upgradeHandler.RegisterRoutes(app)

	// BKL-1078: Webhooks para eventos de produto (HMAC-SHA256, n8n/Make/Zapier)
	webhookHandler := handler.NewWebhookHandler(pool, cfg.ServiceToken)
	webhookHandler.RegisterRoutes(app)

	// BKL-1058: Funis de conversão server-side (POST /api/v1/funnels/analyze)
	funnelHandler := handler.NewFunnelHandler(pool, cfg.ServiceToken)
	funnelHandler.RegisterRoutes(app)

	// BKL-1064: AI Configs — feature flags para prompts e parâmetros de LLM
	aiConfigHandler := handler.NewAIConfigHandler(pool, cfg.ServiceToken)
	aiConfigHandler.RegisterRoutes(app)

	// BKL-1063: LLM Observability — métricas por modelo, latência P50/P99, error rate
	llmObsHandler := handler.NewLLMObservabilityHandler(pool, cfg.ServiceToken)
	llmObsHandler.RegisterRoutes(app)

	// BKL-150: Register WebSocket routes
	wsHandler.RegisterRoutes(app)

	// BKL-344: Servir arquivos estáticos do widget CSAT embeddable.
	// Qualquer site pode embedar com: <script src=".../static/csat/widget.js" data-survey-id="..."></script>
	app.Static("/static", "./frontend/public")

	// BKL-1072: SDK IIFE via CDN — sem bundler necessário.
	// Uso: <script src="https://pulse-api.institutoitinerante.com.br/sdk/pulse.js"></script>
	// ou:  <script src=".../sdk/pulse.min.js"></script>
	app.Static("/sdk", "./frontend/public/sdk")

	// Start uptime monitor background checker
	monitorSvc.StartBackgroundChecker()

	// BKL-801: Start daily health score calculation job
	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()
	service.StartHealthScoreJob(jobCtx, healthScoreSvc)

	// BKL-814: Churn alert — email diário para admin com tenants score < 40
	churnAlertScheduler := service.NewChurnAlertScheduler(healthScoreRepo, 24*time.Hour)
	churnAlertScheduler.Start(jobCtx)

	// BKL-1084: Churn trend alert — 7 dias consecutivos score < 30 → alerta crítico
	churnTrendScheduler := service.NewChurnTrendScheduler(pool)
	churnTrendScheduler.Start(jobCtx)

	// BKL-1040: Flag scheduler — ativa/desativa flags agendadas (starts_at/ends_at) a cada 60s
	go flagscheduler.Start(jobCtx, flagRepo, 60*time.Second)

	// BKL-1037: Retention purge — deleta eventos mais antigos que data_retention_days de cada projeto (1x/dia)
	go retentionpurge.Start(jobCtx, pool, 24*time.Hour)

	// BKL-763: mapeamento cross-produto de identidade de usuário
	userIdentityRepo := repository.NewUserIdentityRepository(pool)
	userIdentityHandler := handler.NewUserIdentityHandler(userIdentityRepo, cfg.ServiceToken)
	userIdentityHandler.RegisterRoutes(app)

	// BKL-864: Weekly insight narrativo — gera toda segunda-feira via Haiku
	weeklyInsightSvc := service.NewWeeklyInsightService(pool)
	insightHandler := handler.NewInsightHandler(weeklyInsightSvc, cfg.ServiceToken)
	insightHandler.RegisterRoutes(app)
	go func() {
		// Verifica toda segunda-feira às 08h se o insight da semana já foi gerado
		for {
			now := time.Now().UTC()
			// Próxima segunda-feira às 08h
			daysUntilMonday := (8 - int(now.Weekday())) % 7
			if daysUntilMonday == 0 && now.Hour() >= 8 {
				daysUntilMonday = 7
			}
			nextRun := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 8, 0, 0, 0, time.UTC)
			select {
			case <-jobCtx.Done():
				return
			case <-time.After(time.Until(nextRun)):
				if err := weeklyInsightSvc.Generate(jobCtx); err != nil {
					log.Printf("weekly_insight: %v", err)
				}
			}
		}
	}()
	// Gera imediatamente ao startup caso ainda não exista insight desta semana
	go func() {
		if err := weeklyInsightSvc.Generate(jobCtx); err != nil {
			log.Printf("weekly_insight startup: %v", err)
		}
	}()

	// BKL-1082: Email sequence de onboarding — processa 1x/dia
	emailSeqSvc := service.NewEmailSequenceService(pool)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		// Processa imediatamente ao startup para não perder o ciclo do dia
		if err := emailSeqSvc.ProcessDue(jobCtx); err != nil {
			log.Printf("email_sequence: ProcessDue: %v", err)
		}
		for {
			select {
			case <-jobCtx.Done():
				return
			case <-ticker.C:
				if err := emailSeqSvc.ProcessDue(jobCtx); err != nil {
					log.Printf("email_sequence: ProcessDue: %v", err)
				}
			}
		}
	}()

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
