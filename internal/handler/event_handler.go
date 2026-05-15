package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/pkg/wshub"
	"github.com/institutoitinerante/pulse-service/internal/service"
	"github.com/redis/go-redis/v9"
)

type EventHandler struct {
	svc          service.EventServicer
	serviceToken string
	hub          *wshub.Hub
	rdb          *redis.Client
	billingURL   string
	identityRepo userIdentityUpsertRepo
}

type userIdentityUpsertRepo interface {
	UpsertByProductUserID(ctx context.Context, product, productUserID, canonicalUserID, email string) error
}

func NewEventHandler(svc service.EventServicer, serviceToken string) *EventHandler {
	return &EventHandler{svc: svc, serviceToken: serviceToken}
}

// WithRedis adiciona o cliente Redis ao EventHandler para quota enforcement (BKL-1077).
func (h *EventHandler) WithRedis(rdb *redis.Client) *EventHandler {
	h.rdb = rdb
	return h
}

// WithIdentityRepo injeta o repositório de identidades para associar anonymous_id → user_id no identify.
func (h *EventHandler) WithIdentityRepo(repo userIdentityUpsertRepo) *EventHandler {
	h.identityRepo = repo
	return h
}

// WithBillingURL configura a URL de upgrade exibida no HTTP 429 de quota esgotada.
func (h *EventHandler) WithBillingURL(url string) *EventHandler {
	h.billingURL = url
	return h
}

// WithHub adiciona o hub WebSocket ao EventHandler para broadcast de eventos em tempo real
func (h *EventHandler) WithHub(hub *wshub.Hub) *EventHandler {
	h.hub = hub
	return h
}

// broadcastEvent serializa e envia o evento para os clientes WS do projeto.
// BKL-1038: usa ProjectID (UUID do tenant) como chave do canal — nunca event.Product
// (string livre do request). Eventos sem project_id não são broadcastados para
// evitar que um cliente adivinhe o nome de produto de outro tenant.
func (h *EventHandler) broadcastEvent(event *domain.Event) {
	if h.hub == nil || event == nil {
		return
	}
	// Sem project_id não há como isolar o canal — não broadcast.
	if event.ProjectID == nil || *event.ProjectID == "" {
		return
	}
	data, err := json.Marshal(map[string]interface{}{
		"type":  "event",
		"event": event,
	})
	if err != nil {
		return
	}
	h.hub.Broadcast(*event.ProjectID, data)
}

// RegisterRoutes mounts all event routes onto the given Fiber app.
func (h *EventHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public endpoint — called from frontends, no auth
	// BKL-1077: quota enforcement (fail-open se Redis nil)
	quotaMW := middleware.QuotaEnforcement(h.rdb, h.billingURL)
	api.Post("/events", quotaMW, h.IngestEvent)

	// BKL-1041: batch ingest de até 500 eventos por chamada (SDK queue flush)
	api.Post("/events/batch", quotaMW, h.BatchIngestEvents)

	// Public endpoint — simplified event tracking
	api.Post("/events/track", quotaMW, h.TrackEvent)

	// Public endpoint — merge anonymous_id com user_id apos login
	api.Post("/events/identify", h.IdentifyUser)

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

	// BKL-150: broadcast para clientes WS do mesmo produto/projeto
	go h.broadcastEvent(event)

	return c.Status(fiber.StatusCreated).JSON(event)
}

// BatchIngestEvents godoc
// POST /api/v1/events/batch
// BKL-1041: persiste até 500 eventos em uma única operação (pgx CopyFrom).
func (h *EventHandler) BatchIngestEvents(c *fiber.Ctx) error {
	var req domain.BatchIngestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	if len(req.Events) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "events array cannot be empty"})
	}
	if len(req.Events) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "batch too large: max 500 events per request"})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	result, err := h.svc.BatchIngest(c.Context(), req, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetSummary godoc
// GET /api/v1/events/summary?project_id=<uuid>
func (h *EventHandler) GetSummary(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	summary, err := h.svc.GetSummary(c.Context(), projectID)
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

	// BKL-150: broadcast para clientes WS do mesmo produto/projeto
	go h.broadcastEvent(event)

	return c.Status(fiber.StatusCreated).JSON(event)
}

// IdentifyUser godoc
// POST /api/v1/events/identify — merge anonymous_id com user_id real e registra identidade cross-produto.
// Campos: anonymous_id, user_id, product (opcional), canonical_user_id (opcional), email (opcional)
func (h *EventHandler) IdentifyUser(c *fiber.Ctx) error {
	var req domain.IdentifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	if req.AnonymousID == "" || req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "anonymous_id e user_id sao obrigatorios"})
	}

	count, err := h.svc.Identify(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	// Registra identidade cross-produto se tiver repositório e product definido
	if h.identityRepo != nil && req.Product != "" {
		canonicalID := req.CanonicalUserID
		if canonicalID == "" {
			canonicalID = req.UserID
		}
		go func() {
			_ = h.identityRepo.UpsertByProductUserID(context.Background(), req.Product, req.UserID, canonicalID, req.Email)
		}()
	}

	return c.JSON(fiber.Map{"merged_events": count, "anonymous_id": req.AnonymousID, "user_id": req.UserID})
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
			if v > 1000 {
				v = 1000
			}
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
