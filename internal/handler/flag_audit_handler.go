package handler

// BKL-1074: Audit log de feature flag evaluations.
// GET /api/v1/flags/audit — histórico paginado de avaliações.

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type FlagAuditHandler struct {
	repo         *repository.FlagEvaluationRepository
	serviceToken string
}

func NewFlagAuditHandler(repo *repository.FlagEvaluationRepository, serviceToken string) *FlagAuditHandler {
	return &FlagAuditHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta rotas de audit log de flags.
func (h *FlagAuditHandler) RegisterRoutes(app *fiber.App) {
	auth := middleware.ServiceTokenMiddleware(h.serviceToken)
	app.Get("/api/v1/flags/audit", auth, h.List)
}

// List godoc
// GET /api/v1/flags/audit?flag_key=&project_id=&limit=&offset=
func (h *FlagAuditHandler) List(c *fiber.Ctx) error {
	flagKey := c.Query("flag_key")
	projectID := c.Query("project_id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	items, total, err := h.repo.List(c.Context(), flagKey, projectID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(domain.FlagEvaluationListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
