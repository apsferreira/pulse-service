package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

type MonitorHandler struct {
	svc          *service.MonitorService
	serviceToken string
}

func NewMonitorHandler(svc *service.MonitorService, serviceToken string) *MonitorHandler {
	return &MonitorHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes mounts all monitor routes onto the given Fiber app.
func (h *MonitorHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public endpoints
	api.Get("/monitors", h.ListMonitors)
	api.Get("/monitors/:id/history", h.GetHistory)

	// Admin endpoints — require service token
	admin := api.Group("/monitors", middleware.ServiceTokenMiddleware(h.serviceToken))
	admin.Post("/", h.CreateMonitor)
	admin.Put("/:id", h.UpdateMonitor)
	admin.Delete("/:id", h.DeleteMonitor)
}

// ListMonitors godoc
// GET /api/v1/monitors
func (h *MonitorHandler) ListMonitors(c *fiber.Ctx) error {
	statuses, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(statuses)
}

// CreateMonitor godoc
// POST /api/v1/monitors
func (h *MonitorHandler) CreateMonitor(c *fiber.Ctx) error {
	var req domain.CreateMonitorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	m, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}

// UpdateMonitor godoc
// PUT /api/v1/monitors/:id
func (h *MonitorHandler) UpdateMonitor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "missing monitor id"})
	}

	var req domain.UpdateMonitorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	m, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if m == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "monitor not found"})
	}
	return c.JSON(m)
}

// DeleteMonitor godoc
// DELETE /api/v1/monitors/:id
func (h *MonitorHandler) DeleteMonitor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "missing monitor id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetHistory godoc
// GET /api/v1/monitors/:id/history
func (h *MonitorHandler) GetHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "missing monitor id"})
	}

	checks, err := h.svc.GetHistory(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(checks)
}
