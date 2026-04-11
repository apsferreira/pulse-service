package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// HealthScoreHandler expõe os endpoints de health score por tenant.
// BKL-801: fundação para churn prevention.
type HealthScoreHandler struct {
	svc          *service.HealthScoreService
	serviceToken string
}

func NewHealthScoreHandler(svc *service.HealthScoreService, serviceToken string) *HealthScoreHandler {
	return &HealthScoreHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de health score.
func (h *HealthScoreHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/health-scores", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/", h.ListAll)       // BKL-1083: visão admin de todos os tenants
	api.Get("/:tenant_id", h.GetHealthScore)
}

// ListAll GET /api/v1/health-scores
// Retorna health scores de todos os tenants, ordenados por score ASC (BKL-1083).
func (h *HealthScoreHandler) ListAll(c *fiber.Ctx) error {
	scores, err := h.svc.ListAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao listar health scores"})
	}

	type item struct {
		*domain.TenantHealthScore
		Status string `json:"status"`
	}
	result := make([]item, len(scores))
	for i, hs := range scores {
		result[i] = item{TenantHealthScore: hs, Status: hs.Status()}
	}
	return c.JSON(fiber.Map{"scores": result, "total": len(result)})
}

// GetHealthScore GET /api/v1/health-scores/:tenant_id
// Retorna o health score atual do tenant, calculado pelo cron diário.
func (h *HealthScoreHandler) GetHealthScore(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "tenant_id obrigatório"})
	}

	hs, err := h.svc.GetByTenantID(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao buscar health score"})
	}

	return c.JSON(domain.HealthScoreResponse{
		TenantHealthScore: hs,
		Status:            hs.Status(),
	})
}
