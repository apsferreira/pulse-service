// BKL-insights: Handler para insights acionáveis gerados pelo agente autônomo.
package handler

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/agent"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// ProjectInsightHandler expõe os endpoints de insights por projeto.
type ProjectInsightHandler struct {
	repo         *repository.ProjectInsightRepository
	insightAgent *agent.InsightsAgent
	serviceToken string
}

// NewProjectInsightHandler cria o handler.
func NewProjectInsightHandler(
	repo *repository.ProjectInsightRepository,
	insightAgent *agent.InsightsAgent,
	serviceToken string,
) *ProjectInsightHandler {
	return &ProjectInsightHandler{
		repo:         repo,
		insightAgent: insightAgent,
		serviceToken: serviceToken,
	}
}

// RegisterRoutes monta as rotas de project insights.
func (h *ProjectInsightHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/insights", h.List)
	api.Post("/insights/:id/dismiss", h.Dismiss)
	api.Post("/insights/run", h.TriggerRun)
}

// List retorna insights de um projeto.
// GET /api/v1/insights?project_id=<uuid>&dismissed=false
func (h *ProjectInsightHandler) List(c *fiber.Ctx) error {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "project_id obrigatório",
		})
	}

	dismissedStr := c.Query("dismissed", "false")
	dismissed := dismissedStr == "true"

	items, err := h.repo.List(c.Context(), projectID, dismissed)
	if err != nil {
		log.Printf("[project-insight] List erro: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error: "erro ao listar insights",
		})
	}

	return c.JSON(fiber.Map{
		"insights": items,
		"total":    len(items),
	})
}

// Dismiss marca um insight como dispensado.
// POST /api/v1/insights/:id/dismiss
// Body: { "dismissed_by": "user@email.com" }
func (h *ProjectInsightHandler) Dismiss(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "id obrigatório",
		})
	}

	var req domain.DismissInsightRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "payload inválido",
		})
	}
	req.DismissedBy = strings.TrimSpace(req.DismissedBy)
	if req.DismissedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "dismissed_by obrigatório",
		})
	}
	if len(req.DismissedBy) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "dismissed_by excede 255 caracteres",
		})
	}

	if err := h.repo.Dismiss(c.Context(), id, req.DismissedBy); err != nil {
		log.Printf("[project-insight] Dismiss id=%s erro: %v", id, err)
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error: "insight não encontrado ou já dispensado",
		})
	}

	return c.JSON(fiber.Map{"status": "dispensado"})
}

// TriggerRun dispara uma execução manual do agente de insights.
// POST /api/v1/insights/run
// Requer service token — não exposto ao frontend direto.
func (h *ProjectInsightHandler) TriggerRun(c *fiber.Ctx) error {
	if err := h.insightAgent.RunNow(c.Context()); err != nil {
		log.Printf("[project-insight] TriggerRun erro: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error: "erro ao executar agente de insights",
		})
	}
	return c.JSON(fiber.Map{
		"status": "análise concluída",
	})
}
