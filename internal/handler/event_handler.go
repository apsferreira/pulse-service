package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

type EventHandler struct {
	svc          *service.EventService
	serviceToken string
}

func NewEventHandler(svc *service.EventService, serviceToken string) *EventHandler {
	return &EventHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes mounts all event routes onto the given Fiber app.
func (h *EventHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public endpoint — called from frontends, no auth
	api.Post("/events", h.IngestEvent)

	// Admin endpoint — requires service token
	api.Get("/events/summary", middleware.ServiceTokenMiddleware(h.serviceToken), h.GetSummary)
}

// IngestEvent godoc
// POST /api/v1/events
func (h *EventHandler) IngestEvent(c *fiber.Ctx) error {
	var req domain.IngestEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	event, err := h.svc.Ingest(c.Context(), req, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(event)
}

// GetSummary godoc
// GET /api/v1/events/summary
func (h *EventHandler) GetSummary(c *fiber.Ctx) error {
	summary, err := h.svc.GetSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(summary)
}
