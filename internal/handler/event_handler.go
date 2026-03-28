package handler

import (
	"strconv"
	"time"

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

	// Public endpoint — simplified event tracking
	api.Post("/events/track", h.TrackEvent)

	// Admin endpoints — require service token
	api.Get("/events/summary", middleware.ServiceTokenMiddleware(h.serviceToken), h.GetSummary)
	api.Get("/events/count", middleware.ServiceTokenMiddleware(h.serviceToken), h.CountEvents)
	api.Get("/events", middleware.ServiceTokenMiddleware(h.serviceToken), h.ListEvents)
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

// TrackEvent godoc
// POST /api/v1/events/track — simplified event tracking endpoint.
func (h *EventHandler) TrackEvent(c *fiber.Ctx) error {
	var req domain.TrackEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if req.EventName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "event_name is required"})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	event, err := h.svc.Track(c.Context(), req, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(event)
}

// CountEvents godoc
// GET /api/v1/events/count — returns event counts grouped by event_name, optionally filtered by period.
func (h *EventHandler) CountEvents(c *fiber.Ctx) error {
	eventName := c.Query("event_name")

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		t, err := time.Parse(time.RFC3339, sd)
		if err != nil {
			t, err = time.Parse("2006-01-02", sd)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid start_date format, use RFC3339 or YYYY-MM-DD"})
			}
		}
		startDate = &t
	}
	if ed := c.Query("end_date"); ed != "" {
		t, err := time.Parse(time.RFC3339, ed)
		if err != nil {
			t, err = time.Parse("2006-01-02", ed)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid end_date format, use RFC3339 or YYYY-MM-DD"})
			}
			// Set end_date to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		endDate = &t
	}

	result, err := h.svc.Count(c.Context(), eventName, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(result)
}

// ListEvents godoc
// GET /api/v1/events — returns events with pagination.
func (h *EventHandler) ListEvents(c *fiber.Ctx) error {
	eventName := c.Query("event_name")
	userID := c.Query("user_id")

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

	result, err := h.svc.List(c.Context(), eventName, userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(result)
}
