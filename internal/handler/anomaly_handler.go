// Package handler — AnomalyHandler: endpoint GET /api/v1/analytics/anomalies (BKL-865).
// Detecta anomalias em métricas de eventos usando média móvel + desvio padrão.
// Gemini 2.5 Flash gera explicação human-readable como fallback para explicação padrão.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// AnomalyHandler agrupa o endpoint de detecção de anomalias.
type AnomalyHandler struct {
	svc          *service.AnomalyService
	serviceToken string
}

// NewAnomalyHandler cria o handler de anomalias.
func NewAnomalyHandler(svc *service.AnomalyService, serviceToken string) *AnomalyHandler {
	return &AnomalyHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de anomalias no app Fiber.
func (h *AnomalyHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Get("/analytics/anomalies", h.Detect)
}

// Detect analisa uma métrica e retorna resultado de anomalia.
// GET /api/v1/analytics/anomalies?project_id=X&metric=page_view&window=7d
func (h *AnomalyHandler) Detect(c *fiber.Ctx) error {
	// Autenticação por service token (padrão do pulse-service)
	if token := c.Get("X-Service-Token"); token != h.serviceToken {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "não autorizado",
		})
	}

	projectID := c.Query("project_id")
	metric := c.Query("metric")
	window := c.Query("window", "7d")

	if metric == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "parâmetro 'metric' é obrigatório",
		})
	}

	result, err := h.svc.Detect(c.Context(), projectID, metric, window)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"data": result})
}
