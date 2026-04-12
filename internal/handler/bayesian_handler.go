package handler

// BKL-1001: Endpoint de análise Bayesiana para experimentos A/B.
// GET /api/v1/experiments/:id/bayesian
// Aceita os dados de conversão via query params para simplicidade de consumo
// (não altera estado, semanticamente GET).

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// BayesianHandler expõe o endpoint de análise Bayesiana.
type BayesianHandler struct{}

func NewBayesianHandler() *BayesianHandler {
	return &BayesianHandler{}
}

// RegisterRoutes monta a rota de Bayesian analysis.
func (h *BayesianHandler) RegisterRoutes(app *fiber.App) {
	// Endpoint público — sem auth, apenas matemática (igual ao /experiments/analyze)
	app.Get("/api/v1/experiments/:id/bayesian", h.Analyze)
}

// Analyze godoc
// GET /api/v1/experiments/:id/bayesian
// Query params: a_conversions, a_total, b_conversions, b_total
// Response: BayesianResult com distribuições Beta e P(B > A)
// BKL-1001
func (h *BayesianHandler) Analyze(c *fiber.Ctx) error {
	parseInt := func(key string) (int, error) {
		v := c.Query(key)
		if v == "" {
			return 0, fiber.NewError(fiber.StatusBadRequest, key+" obrigatório")
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return 0, fiber.NewError(fiber.StatusBadRequest, key+" deve ser inteiro >= 0")
		}
		return n, nil
	}

	aConv, err := parseInt("a_conversions")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	aTotal, err := parseInt("a_total")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	bConv, err := parseInt("b_conversions")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	bTotal, err := parseInt("b_total")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	if aTotal <= 0 || bTotal <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "a_total e b_total devem ser > 0"})
	}
	if aConv > aTotal {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "a_conversions não pode ser maior que a_total"})
	}
	if bConv > bTotal {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "b_conversions não pode ser maior que b_total"})
	}

	result := service.AnalyzeBayesian(service.BayesianInput{
		AConversions: aConv,
		ATotal:       aTotal,
		BConversions: bConv,
		BTotal:       bTotal,
	})

	return c.JSON(result)
}
