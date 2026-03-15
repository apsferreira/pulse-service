package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

type FlagHandler struct {
	svc          *service.FlagService
	serviceToken string
}

func NewFlagHandler(svc *service.FlagService, serviceToken string) *FlagHandler {
	return &FlagHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes mounts all flag routes onto the given Fiber app.
func (h *FlagHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public endpoint — no auth required (called from SDKs)
	api.Get("/flags/evaluate", h.EvaluateFlags)

	// Admin endpoints — require service token
	admin := api.Group("/flags", middleware.ServiceTokenMiddleware(h.serviceToken))
	admin.Get("/", h.ListFlags)
	admin.Post("/", h.CreateFlag)
	admin.Put("/:id", h.UpdateFlag)
	admin.Delete("/:id", h.DeleteFlag)
}

// ListFlags godoc
// GET /api/v1/flags
// Query params: product (optional)
func (h *FlagHandler) ListFlags(c *fiber.Ctx) error {
	var product *string
	if p := c.Query("product"); p != "" {
		product = &p
	}

	flags, err := h.svc.List(c.Context(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if flags == nil {
		flags = []domain.FeatureFlag{}
	}
	return c.JSON(flags)
}

// CreateFlag godoc
// POST /api/v1/flags
func (h *FlagHandler) CreateFlag(c *fiber.Ctx) error {
	var req domain.CreateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if len(req.Name) < 2 || len(req.Name) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "name must be between 2 and 100 characters"})
	}

	flag, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(flag)
}

// UpdateFlag godoc
// PUT /api/v1/flags/:id
func (h *FlagHandler) UpdateFlag(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "missing flag id"})
	}

	var req domain.UpdateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	flag, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}
	return c.JSON(flag)
}

// DeleteFlag godoc
// DELETE /api/v1/flags/:id
func (h *FlagHandler) DeleteFlag(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "missing flag id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// EvaluateFlags godoc
// GET /api/v1/flags/evaluate
// Query params: product (required), user_id (optional), tenant_id (optional)
func (h *FlagHandler) EvaluateFlags(c *fiber.Ctx) error {
	product := c.Query("product")
	userID := c.Query("user_id")
	tenantID := c.Query("tenant_id")

	resp, err := h.svc.EvaluateFlags(c.Context(), product, userID, tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(resp)
}
