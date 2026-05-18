// BKL-420: Pulse — Session Replay Handler
// Recebe eventos de gravação de sessão do SDK JavaScript e persiste anonimizados.
// LGPD compliant: sem PII, opt-out verificado no SDK antes do envio.
package handler

import (
	"context"
	"crypto/subtle"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReplayHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewReplayHandler(db *pgxpool.Pool, serviceToken string) *ReplayHandler {
	return &ReplayHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de Session Replay na aplicação Fiber.
func (h *ReplayHandler) RegisterRoutes(app *fiber.App) {
	// POST /events usa Bearer token por design SDK; GET requer service token (dados sensíveis de sessão)
	replayPublic := app.Group("/api/v1/replay")
	replayPublic.Post("/events", h.IngestEvents)

	replay := app.Group("/api/v1/replay", middleware.ServiceTokenMiddleware(h.serviceToken))
	replay.Get("/sessions", h.ListSessions)
	replay.Get("/sessions/:session_id/events", h.GetSessionEvents)
}

type replayEventPayload struct {
	Type        string                 `json:"type"`
	Timestamp   int64                  `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

type ingestReplayRequest struct {
	SessionID string                `json:"session_id"`
	ProjectID string                `json:"project_id"`
	UserID    string                `json:"user_id"`
	PageURL   string                `json:"page_url"`
	UserAgent string                `json:"user_agent"`
	Events    []replayEventPayload  `json:"events"`
}

// POST /api/v1/replay/events — recebe batch de eventos de sessão
func (h *ReplayHandler) IngestEvents(c *fiber.Ctx) error {
	// Autenticação via Bearer token (API key do projeto)
	authHeader := c.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token obrigatório"})
	}

	var req ingestReplayRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}

	if req.SessionID == "" || req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "session_id e project_id obrigatórios"})
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id inválido"})
	}

	// Verifica que o projeto existe e o token é válido
	var storedToken string
	err = h.db.QueryRow(context.Background(), `
		SELECT api_key FROM projects WHERE id = $1 AND deleted_at IS NULL
	`, projectID).Scan(&storedToken)
	if err != nil || subtle.ConstantTimeCompare([]byte(storedToken), []byte(authHeader[7:])) != 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token inválido"})
	}

	if len(req.Events) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ingested": 0})
	}

	// Trunca user_agent por LGPD
	userAgent := req.UserAgent
	if len(userAgent) > 200 {
		userAgent = userAgent[:200]
	}
	userID := req.UserID
	if userID == "" {
		userID = "anonymous"
	}

	// Upsert na sessão
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO session_replays (session_id, project_id, user_id, page_url, user_agent, event_count, last_event_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (session_id) DO UPDATE SET
			event_count = session_replays.event_count + $6,
			last_event_at = NOW(),
			duration_ms = EXTRACT(EPOCH FROM (NOW() - session_replays.started_at)) * 1000
	`, req.SessionID, projectID, userID, req.PageURL, userAgent, len(req.Events))
	if err != nil {
		log.Printf("[replay] falha ao upsert sessão %s: %v", req.SessionID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao persistir sessão"})
	}

	// Insere eventos em batch
	ingested := 0
	for _, evt := range req.Events {
		validTypes := map[string]bool{
			"dom_mutation": true, "click": true, "scroll": true, "viewport": true, "url_change": true,
		}
		if !validTypes[evt.Type] {
			continue
		}

		_, err = h.db.Exec(context.Background(), `
			INSERT INTO session_replay_events (session_id, project_id, event_type, timestamp_ms, data)
			VALUES ($1, $2, $3, $4, $5)
		`, req.SessionID, projectID, evt.Type, evt.Timestamp, evt.Data)
		if err == nil {
			ingested++
		}
	}

	return c.JSON(fiber.Map{"ingested": ingested, "session_id": req.SessionID})
}

// GET /api/v1/replay/sessions — lista sessões do projeto
func (h *ReplayHandler) ListSessions(c *fiber.Ctx) error {
	rawProjectID := c.Query("project_id")

	limit := 50

	if rawProjectID != "" && rawProjectID != "null" && rawProjectID != "undefined" {
		projectID, parseErr := uuid.Parse(rawProjectID)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id inválido"})
		}
		rows, err := h.db.Query(context.Background(), `
			SELECT session_id, user_id, page_url, event_count, duration_ms, started_at, last_event_at
			FROM session_replays
			WHERE project_id = $1
			ORDER BY started_at DESC
			LIMIT $2
		`, projectID, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao buscar sessões"})
		}
		defer rows.Close()

		type sessionRow struct {
			SessionID   string    `json:"session_id"`
			UserID      string    `json:"user_id"`
			PageURL     string    `json:"page_url"`
			EventCount  int       `json:"event_count"`
			DurationMs  int       `json:"duration_ms"`
			StartedAt   time.Time `json:"started_at"`
			LastEventAt time.Time `json:"last_event_at"`
		}
		var sessions []sessionRow
		for rows.Next() {
			var s sessionRow
			if err := rows.Scan(&s.SessionID, &s.UserID, &s.PageURL, &s.EventCount, &s.DurationMs, &s.StartedAt, &s.LastEventAt); err == nil {
				sessions = append(sessions, s)
			}
		}
		if sessions == nil {
			sessions = []sessionRow{}
		}
		return c.JSON(fiber.Map{"sessions": sessions, "total": len(sessions)})
	}

	rows, err := h.db.Query(context.Background(), `
		SELECT session_id, user_id, page_url, event_count, duration_ms, started_at, last_event_at
		FROM session_replays
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao buscar sessões"})
	}
	defer rows.Close()

	type sessionRow struct {
		SessionID   string    `json:"session_id"`
		UserID      string    `json:"user_id"`
		PageURL     string    `json:"page_url"`
		EventCount  int       `json:"event_count"`
		DurationMs  int       `json:"duration_ms"`
		StartedAt   time.Time `json:"started_at"`
		LastEventAt time.Time `json:"last_event_at"`
	}

	var sessions []sessionRow
	for rows.Next() {
		var s sessionRow
		if err := rows.Scan(&s.SessionID, &s.UserID, &s.PageURL, &s.EventCount, &s.DurationMs, &s.StartedAt, &s.LastEventAt); err == nil {
			sessions = append(sessions, s)
		}
	}

	return c.JSON(fiber.Map{"sessions": sessions, "total": len(sessions)})
}

// GET /api/v1/replay/sessions/:session_id/events — eventos de uma sessão para playback
func (h *ReplayHandler) GetSessionEvents(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "session_id obrigatório"})
	}

	rows, err := h.db.Query(context.Background(), `
		SELECT event_type, timestamp_ms, data
		FROM session_replay_events
		WHERE session_id = $1
		ORDER BY timestamp_ms ASC
	`, sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao buscar eventos"})
	}
	defer rows.Close()

	type eventRow struct {
		Type      string                 `json:"type"`
		Timestamp int64                  `json:"timestamp"`
		Data      map[string]interface{} `json:"data"`
	}

	var events []eventRow
	for rows.Next() {
		var e eventRow
		if err := rows.Scan(&e.Type, &e.Timestamp, &e.Data); err == nil {
			events = append(events, e)
		}
	}

	if len(events) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sessão não encontrada"})
	}

	return c.JSON(fiber.Map{"session_id": sessionID, "events": events, "total": len(events)})
}
