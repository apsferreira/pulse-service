package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
)

type CostHandler struct {
	repo         domain.CostRepository
	serviceToken string
}

func NewCostHandler(repo domain.CostRepository, serviceToken string) *CostHandler {
	return &CostHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas de cost tracking.
// Todas requerem X-Service-Token ou Authorization Bearer JWT.
func (h *CostHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))

	api.Post("/costs", h.RecordCost)
	api.Get("/costs/overview", h.GetOverview)
	api.Get("/costs/records", h.ListRecords)
	api.Get("/costs/total", h.GetTotal)
}

// parsePeriod converte o query param "days" em from/to. Default: 30 dias.
func parsePeriod(c *fiber.Ctx) (from, to time.Time) {
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	to = time.Now().UTC()
	from = to.AddDate(0, 0, -days)
	return
}

// RecordCost godoc
// POST /api/v1/costs
// Ingestao de custo vinda de servicos externos (SocialMake, attend-service, etc.).
func (h *CostHandler) RecordCost(c *fiber.Ctx) error {
	var req domain.RecordCostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "corpo da requisicao invalido"})
	}
	if req.Product == "" || req.Operation == "" || req.Provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "product, operation e provider sao obrigatorios",
		})
	}

	cost := &domain.APICost{
		ID:        uuid.New(),
		Product:   req.Product,
		Operation: req.Operation,
		Provider:  req.Provider,
		Model:     req.Model,
		TokensIn:  req.TokensIn,
		TokensOut: req.TokensOut,
		Images:    req.Images,
		CostUSD:   req.CostUSD,
		CostBRL:   req.CostUSD * domain.USDToBRL,
		Metadata:  req.Metadata,
	}

	if err := h.repo.Record(c.Context(), cost); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao registrar custo"})
	}
	return c.Status(fiber.StatusCreated).JSON(cost)
}

// GetOverview godoc
// GET /api/v1/costs/overview?product=socialmake&days=30
func (h *CostHandler) GetOverview(c *fiber.Ctx) error {
	product := c.Query("product")
	from, to := parsePeriod(c)

	overview, err := h.repo.GetOverview(c.Context(), product, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao buscar overview de custos"})
	}
	return c.JSON(overview)
}

// ListRecords godoc
// GET /api/v1/costs/records?product=socialmake&days=30&limit=100
func (h *CostHandler) ListRecords(c *fiber.Ctx) error {
	product := c.Query("product")
	from, to := parsePeriod(c)

	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	records, err := h.repo.ListByProduct(c.Context(), product, from, to, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao listar registros de custo"})
	}
	return c.JSON(fiber.Map{
		"records": records,
		"count":   len(records),
	})
}

// GetTotal godoc
// GET /api/v1/costs/total?days=30
func (h *CostHandler) GetTotal(c *fiber.Ctx) error {
	from, to := parsePeriod(c)

	total, err := h.repo.TotalByPeriod(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao calcular total de custos"})
	}
	return c.JSON(fiber.Map{
		"from":      from,
		"to":        to,
		"total_usd": total,
		"total_brl": total * domain.USDToBRL,
	})
}
