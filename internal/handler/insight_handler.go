// BKL-864: Handler para insights narrativos semanais.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// InsightHandler expõe o endpoint de insight narrativo semanal.
type InsightHandler struct {
	svc          *service.WeeklyInsightService
	serviceToken string
}

// NewInsightHandler cria o handler.
func NewInsightHandler(svc *service.WeeklyInsightService, serviceToken string) *InsightHandler {
	return &InsightHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas de insight.
func (h *InsightHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/insights/weekly", h.GetLatest)
	api.Post("/insights/weekly/generate", h.TriggerGenerate)
}

// GetLatest retorna o insight narrativo mais recente.
// GET /api/v1/insights/weekly
func (h *InsightHandler) GetLatest(c *fiber.Ctx) error {
	report, err := h.svc.GetLatest(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "insight não disponível",
			"details": err.Error(),
		})
	}
	return c.JSON(report)
}

// TriggerGenerate força geração imediata do insight semanal (útil para teste/debug).
// POST /api/v1/insights/weekly/generate
func (h *InsightHandler) TriggerGenerate(c *fiber.Ctx) error {
	if err := h.svc.Generate(c.Context()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "falha ao gerar insight",
			"details": err.Error(),
		})
	}
	report, err := h.svc.GetLatest(c.Context())
	if err != nil {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"status": "gerado",
			"note":   "insight gerado mas não encontrado na busca subsequente",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(report)
}
