package handler

// BKL-1044: Endpoints para vinculação de flags a métricas de negócio e cálculo de lift.
// Permite associar um event_type (ex: "checkout_completed") a uma flag,
// e consultar o lift real calculado via JOIN entre flag_evaluations e events.

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type FlagMetricHandler struct {
	repo         *repository.FlagMetricRepository
	serviceToken string
}

func NewFlagMetricHandler(repo *repository.FlagMetricRepository, serviceToken string) *FlagMetricHandler {
	return &FlagMetricHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas de métricas de flags.
func (h *FlagMetricHandler) RegisterRoutes(app *fiber.App) {
	// Todos os endpoints requerem service token (operações administrativas)
	v1 := app.Group("/api/v1", middleware.ServiceTokenAuth(h.serviceToken))

	v1.Get("/flags/:key/metrics", h.List)
	v1.Post("/flags/:key/metrics", h.Create)
	v1.Delete("/flags/:key/metrics/:metric_name", h.Delete)
	v1.Get("/flags/:key/metrics/:metric_name/lift", h.GetLift)
}

// List godoc
// GET /api/v1/flags/:key/metrics?project_id=<uuid>
// Response: []FlagMetric
func (h *FlagMetricHandler) List(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}

	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	metrics, err := h.repo.ListByFlag(c.Context(), key, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if metrics == nil {
		metrics = []domain.FlagMetric{}
	}
	return c.JSON(metrics)
}

// Create godoc
// POST /api/v1/flags/:key/metrics?project_id=<uuid>
// Body: CreateFlagMetricRequest
// Response: FlagMetric (201)
func (h *FlagMetricHandler) Create(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}

	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	var req domain.CreateFlagMetricRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}

	req.MetricName = strings.TrimSpace(req.MetricName)
	if req.MetricName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "metric_name é obrigatório"})
	}
	if len(req.MetricName) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "metric_name deve ter no máximo 100 caracteres"})
	}

	m, err := h.repo.Create(c.Context(), key, projectID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}

// Delete godoc
// DELETE /api/v1/flags/:key/metrics/:metric_name?project_id=<uuid>
// Response: 204 No Content
func (h *FlagMetricHandler) Delete(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}

	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	metricName := c.Params("metric_name")
	if metricName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "metric_name é obrigatório"})
	}

	if err := h.repo.Delete(c.Context(), key, projectID, metricName); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetLift godoc
// GET /api/v1/flags/:key/metrics/:metric_name/lift?project_id=<uuid>
// Response: FlagMetricLift
// BKL-1044: JOIN entre flag_evaluations e events para calcular lift real.
func (h *FlagMetricHandler) GetLift(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}

	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	metricName := c.Params("metric_name")
	if metricName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "metric_name é obrigatório"})
	}

	lift, err := h.repo.GetLift(c.Context(), key, projectID, metricName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if lift == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "vinculação flag→métrica não encontrada"})
	}
	return c.JSON(lift)
}
