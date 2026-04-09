package handler

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// validKeyPattern allows only alphanumeric, underscores, hyphens and dots.
var validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,100}$`)

type FlagHandler struct {
	svc          *service.FlagService
	serviceToken string
}

func NewFlagHandler(svc *service.FlagService, serviceToken string) *FlagHandler {
	return &FlagHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes mounts all flag routes onto the given Fiber app.
func (h *FlagHandler) RegisterRoutes(app *fiber.App) {
	// Public endpoints — no auth required (consumed by SDKs)
	app.Get("/flags", h.GetFlagsMap)
	app.Post("/flags/evaluate-batch", h.EvaluateBatch) // BKL-093: batch evaluation

	// Admin CRUD endpoints — require service token
	api := app.Group("/api/v1/flags", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/", h.CreateFlag)
	api.Get("/", h.ListFlags)
	api.Get("/:key", h.GetFlagByKey)
	api.Put("/:key", h.UpdateFlag)
	api.Delete("/:key", h.DeleteFlag)
}

// GetFlagsMap godoc
// GET /flags — public, returns {flags: {key: enabled}} for SDK consumption.
// Query params opcionais: user_id, tenant_id — quando fornecidos, avalia rollout e targeting.
func (h *FlagHandler) GetFlagsMap(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	tenantID := c.Query("tenant_id")

	// Se tem contexto, avalia com rollout + targeting
	if userID != "" || tenantID != "" {
		resp, err := h.svc.EvaluateFlags(c.Context(), userID, tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to evaluate flags"})
		}
		return c.JSON(resp)
	}

	// Sem contexto — retorna mapa simples (cache Redis)
	resp, err := h.svc.GetFlagsMap(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to load flags"})
	}
	return c.JSON(resp)
}

// ListFlags godoc
// GET /api/v1/flags — returns all flags with full details.
func (h *FlagHandler) ListFlags(c *fiber.Ctx) error {
	flags, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to list flags"})
	}
	if flags == nil {
		flags = []domain.FeatureFlag{}
	}
	return c.JSON(flags)
}

// GetFlagByKey godoc
// GET /api/v1/flags/:key
func (h *FlagHandler) GetFlagByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	flag, err := h.svc.GetByKey(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to get flag"})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}
	return c.JSON(flag)
}

// CreateFlag godoc
// POST /api/v1/flags
func (h *FlagHandler) CreateFlag(c *fiber.Ctx) error {
	var req domain.CreateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if !validKeyPattern.MatchString(req.Key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "key must be 2-100 chars, alphanumeric/underscore/hyphen/dot only"})
	}

	flag, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(flag)
}

// UpdateFlag godoc
// PUT /api/v1/flags/:key
func (h *FlagHandler) UpdateFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	var req domain.UpdateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	flag, err := h.svc.Update(c.Context(), key, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to update flag"})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}
	return c.JSON(flag)
}

// DeleteFlag godoc
// DELETE /api/v1/flags/:key
func (h *FlagHandler) DeleteFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	if err := h.svc.Delete(c.Context(), key); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to delete flag"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// EvaluateBatch godoc
// POST /flags/evaluate-batch — BKL-093: avalia N flags de uma vez para SDKs.
// Body: {"keys": ["flag-a", "flag-b"], "user_id": "...", "tenant_id": "..."}
// Response: {"flags": {"flag-a": true, "flag-b": false}}
func (h *FlagHandler) EvaluateBatch(c *fiber.Ctx) error {
	var req struct {
		Keys     []string `json:"keys"`
		UserID   string   `json:"user_id"`
		TenantID string   `json:"tenant_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	if len(req.Keys) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "keys must not be empty"})
	}
	if len(req.Keys) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "max 50 keys per batch"})
	}

	// Evaluate all flags with context
	allFlags, err := h.svc.EvaluateFlags(c.Context(), req.UserID, req.TenantID)
	if err != nil {
		// Fail-open: return all false defaults
		result := make(map[string]bool, len(req.Keys))
		for _, k := range req.Keys {
			result[k] = false
		}
		return c.JSON(domain.FlagsMapResponse{Flags: result})
	}

	// Filter to only requested keys
	result := make(map[string]bool, len(req.Keys))
	for _, k := range req.Keys {
		result[k] = allFlags.Flags[k]
	}
	return c.JSON(domain.FlagsMapResponse{Flags: result})
}
