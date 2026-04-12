package handler

// BKL-998: Health score configurável por projeto.
// Endpoints:
//   PUT /api/v1/projects/:id/health-config  — salva pesos e thresholds
//   GET /api/v1/projects/:id/health         — retorna score calculado com a config salva

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectHealthHandler expõe os endpoints de health score configurável por projeto.
type ProjectHealthHandler struct {
	projectRepo *repository.ProjectRepository
	db          *pgxpool.Pool
	serviceToken string
}

// NewProjectHealthHandler cria o handler.
func NewProjectHealthHandler(projectRepo *repository.ProjectRepository, db *pgxpool.Pool, serviceToken string) *ProjectHealthHandler {
	return &ProjectHealthHandler{
		projectRepo:  projectRepo,
		db:           db,
		serviceToken: serviceToken,
	}
}

// RegisterRoutes registra as rotas de health score configurável.
func (h *ProjectHealthHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/projects", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Put("/:id/health-config", h.UpdateHealthConfig)
	api.Get("/:id/health", h.GetHealth)
}

// UpdateHealthConfig godoc
// PUT /api/v1/projects/:id/health-config
// Body: {"weights":{"flags_with_audit":0.3,...},"thresholds":{"green":80,"yellow":60}}
func (h *ProjectHealthHandler) UpdateHealthConfig(c *fiber.Ctx) error {
	projectID := c.Params("id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project id obrigatório"})
	}

	var req domain.UpdateHealthConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}

	// Valida que os pesos somam ~1.0 (tolerância 0.01)
	total := req.Weights.FlagsWithAudit + req.Weights.ExperimentsRunning +
		req.Weights.EventsTracked24h + req.Weights.ErrorRate
	if total > 0 && (total < 0.99 || total > 1.01) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: fmt.Sprintf("soma dos pesos deve ser 1.0, recebido %.4f", total),
		})
	}

	cfg := domain.ProjectHealthConfig{
		Weights:    req.Weights,
		Thresholds: req.Thresholds,
	}
	if err := h.projectRepo.UpdateHealthConfig(c.Context(), projectID, cfg); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{"message": "health config atualizada", "project_id": projectID, "config": cfg})
}

// GetHealth godoc
// GET /api/v1/projects/:id/health
// Calcula o health score usando a config salva do projeto.
// Dimensões calculadas em tempo real (últimas 24h/7d de dados).
func (h *ProjectHealthHandler) GetHealth(c *fiber.Ctx) error {
	projectID := c.Params("id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project id obrigatório"})
	}

	cfg, err := h.projectRepo.GetHealthConfig(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao carregar config"})
	}

	dims, err := h.calcDimensions(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao calcular score"})
	}

	score := dims["flags_with_audit"]*cfg.Weights.FlagsWithAudit*100 +
		dims["experiments_running"]*cfg.Weights.ExperimentsRunning*100 +
		dims["events_tracked_24h"]*cfg.Weights.EventsTracked24h*100 +
		dims["error_rate"]*cfg.Weights.ErrorRate*100

	status := "critical"
	switch {
	case score >= cfg.Thresholds.Green:
		status = "healthy"
	case score >= cfg.Thresholds.Yellow:
		status = "at_risk"
	}

	return c.JSON(domain.ProjectHealthScoreResponse{
		ProjectID:    projectID,
		Score:        round2(score),
		Status:       status,
		Dimensions:   dims,
		Config:       cfg,
		CalculatedAt: time.Now().UTC(),
	})
}

// calcDimensions calcula as 4 dimensões do health score para um projeto.
// Cada dimensão retorna um valor 0.0–1.0 (normalizado).
func (h *ProjectHealthHandler) calcDimensions(ctx context.Context, projectID string) (map[string]float64, error) {
	dims := map[string]float64{
		"flags_with_audit":     0,
		"experiments_running":  0,
		"events_tracked_24h":   0,
		"error_rate":           0,
	}

	// Dimensão 1: proporção de flags com pelo menos 1 avaliação registrada (audit coverage)
	var totalFlags, auditedFlags int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM feature_flags WHERE project_id = $1
	`, projectID).Scan(&totalFlags)
	if totalFlags > 0 {
		_ = h.db.QueryRow(ctx, `
			SELECT COUNT(DISTINCT flag_key) FROM flag_evaluations WHERE project_id = $1
		`, projectID).Scan(&auditedFlags)
		dims["flags_with_audit"] = float64(auditedFlags) / float64(totalFlags)
	}

	// Dimensão 2: tem experimentos ativos (flag_evaluations nas últimas 24h)
	var recentEvals int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM flag_evaluations
		WHERE project_id = $1 AND evaluated_at >= NOW() - INTERVAL '24 hours'
	`, projectID).Scan(&recentEvals)
	// Normaliza: 100+ evals nas últimas 24h = score 1.0
	if recentEvals >= 100 {
		dims["experiments_running"] = 1.0
	} else {
		dims["experiments_running"] = float64(recentEvals) / 100.0
	}

	// Dimensão 3: eventos rastreados nas últimas 24h
	var eventsCount int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM events
		WHERE properties->>'project_id' = $1
		  AND created_at >= NOW() - INTERVAL '24 hours'
	`, projectID).Scan(&eventsCount)
	// Normaliza: 500+ eventos nas últimas 24h = score 1.0
	if eventsCount >= 500 {
		dims["events_tracked_24h"] = 1.0
	} else {
		dims["events_tracked_24h"] = float64(eventsCount) / 500.0
	}

	// Dimensão 4: error_rate invertida — proporção de eventos de erro nas últimas 24h
	// Score 1.0 = sem erros; 0.0 = 100% erros
	var errorEvents int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM events
		WHERE properties->>'project_id' = $1
		  AND created_at >= NOW() - INTERVAL '24 hours'
		  AND (event_type LIKE 'error%' OR event_type LIKE '%_error' OR event_type = 'exception')
	`, projectID).Scan(&errorEvents)
	if eventsCount > 0 {
		errorRate := float64(errorEvents) / float64(eventsCount)
		dims["error_rate"] = 1.0 - errorRate // invertida: menos erros = score maior
	} else {
		dims["error_rate"] = 1.0 // sem dados = assume sem erros
	}

	return dims, nil
}
