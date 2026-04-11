package handler

// BKL-1043: Endpoint de análise estatística de experimentos A/B.
// BKL-1073: Expõe p-value + CI para consumo pela UI do dashboard.

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

type ExperimentHandler struct{}

func NewExperimentHandler() *ExperimentHandler {
	return &ExperimentHandler{}
}

// RegisterRoutes monta as rotas de experimentos.
func (h *ExperimentHandler) RegisterRoutes(app *fiber.App) {
	// Endpoint público de análise estatística — não requer auth (apenas matemática)
	app.Post("/api/v1/experiments/analyze", h.Analyze)
}

// Analyze godoc
// POST /api/v1/experiments/analyze
// Body: ExperimentInput (control/treatment conversions + totals)
// Response: ExperimentResult com p-value, chi-squared, IC, lift e recomendação
// BKL-1043, BKL-1073
func (h *ExperimentHandler) Analyze(c *fiber.Ctx) error {
	var input service.ExperimentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "payload inválido — envie control_conversions, control_total, treatment_conversions, treatment_total",
		})
	}

	if input.ControlTotal <= 0 || input.TreatmentTotal <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "control_total e treatment_total devem ser > 0",
		})
	}
	if input.ControlConversions < 0 || input.ControlConversions > input.ControlTotal {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "control_conversions deve estar entre 0 e control_total",
		})
	}
	if input.TreatmentConversions < 0 || input.TreatmentConversions > input.TreatmentTotal {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "treatment_conversions deve estar entre 0 e treatment_total",
		})
	}

	result := service.AnalyzeExperiment(input)
	return c.JSON(result)
}
