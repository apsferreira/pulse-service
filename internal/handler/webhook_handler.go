package handler

// BKL-1078: Webhooks para eventos de produto do Pulse.
// Registra URLs externas e dispara payloads assinados com HMAC-SHA256
// quando eventos relevantes ocorrem (event.tracked, flag.evaluated, health_score.dropped).
// Compatível com n8n, Make, Zapier, Slack, Datadog.
//
// Migration 043 adiciona:
//   - webhook_type (slack|datadog|custom) — formatação específica por destino
//   - success + duration_ms em webhook_deliveries — métricas de latência

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookType representa o destino do webhook para formatação correta do payload.
type WebhookType string

const (
	WebhookTypeSlack    WebhookType = "slack"
	WebhookTypeDatadog  WebhookType = "datadog"
	WebhookTypeCustom   WebhookType = "custom"
)

// WebhookHandler gerencia o registro e disparo de webhooks.
type WebhookHandler struct {
	db           *pgxpool.Pool
	projectRepo  *repository.ProjectRepository
	serviceToken string
	httpClient   *http.Client
}

func NewWebhookHandler(db *pgxpool.Pool, serviceToken string) *WebhookHandler {
	return &WebhookHandler{
		db:           db,
		serviceToken: serviceToken,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// WithProjectRepo injeta o ProjectRepository para enforcement de limites por plano.
func (h *WebhookHandler) WithProjectRepo(repo *repository.ProjectRepository) *WebhookHandler {
	h.projectRepo = repo
	return h
}

// RegisterRoutes registra as rotas de webhooks.
func (h *WebhookHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/webhooks", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/", h.List)
	api.Post("/", h.Create)
	api.Delete("/:id", h.Delete)
	api.Post("/:id/test", h.Test)
	api.Get("/:id/deliveries", h.ListDeliveries)
}

// WebhookRow representa um webhook no banco.
type WebhookRow struct {
	ID          string      `json:"id"`
	ProjectID   string      `json:"project_id"`
	URL         string      `json:"url"`
	Events      []string    `json:"events"`
	WebhookType WebhookType `json:"webhook_type"`
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
}

// WebhookDeliveryRow representa uma entrega no histórico.
type WebhookDeliveryRow struct {
	ID          int        `json:"id"`
	WebhookID   string     `json:"webhook_id"`
	EventType   string     `json:"event_type"`
	StatusCode  int        `json:"status_code"`
	Success     bool       `json:"success"`
	DurationMs  *int       `json:"duration_ms,omitempty"`
	Error       *string    `json:"error,omitempty"`
	DeliveredAt time.Time  `json:"delivered_at"`
}

// WebhookCreateRequest é o payload para POST /api/v1/webhooks.
type WebhookCreateRequest struct {
	ProjectID   string      `json:"project_id"`
	URL         string      `json:"url"`
	Events      []string    `json:"events"` // ex: ["event.tracked","flag.evaluated"]
	WebhookType WebhookType `json:"webhook_type"` // slack|datadog|custom (default: custom)
}

// List GET /api/v1/webhooks?project_id=...
func (h *WebhookHandler) List(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT id, project_id, url, events, webhook_type, is_active, created_at
		FROM webhooks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar webhooks"})
	}
	defer rows.Close()

	var webhooks []WebhookRow
	for rows.Next() {
		var w WebhookRow
		if err := rows.Scan(&w.ID, &w.ProjectID, &w.URL, &w.Events, &w.WebhookType, &w.IsActive, &w.CreatedAt); err != nil {
			continue
		}
		webhooks = append(webhooks, w)
	}
	if webhooks == nil {
		webhooks = []WebhookRow{}
	}

	return c.JSON(fiber.Map{"webhooks": webhooks, "total": len(webhooks)})
}

// Create POST /api/v1/webhooks
func (h *WebhookHandler) Create(c *fiber.Ctx) error {
	var req WebhookCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.ProjectID == "" || req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id e url são obrigatórios"})
	}

	// billing: webhooks disponíveis apenas a partir do plano growth
	if h.projectRepo != nil {
		proj, perr := h.projectRepo.GetByID(c.Context(), req.ProjectID)
		if perr == nil && proj != nil {
			switch proj.Plan {
			case "free":
				return upgradeGate(c, proj.Plan, "growth")
			case "starter":
				return upgradeGate(c, proj.Plan, "growth")
			}
		}
	}
	if len(req.Events) == 0 {
		req.Events = []string{"event.tracked", "flag.evaluated", "health_score.dropped"}
	}
	if req.WebhookType == "" {
		req.WebhookType = WebhookTypeCustom
	}
	if req.WebhookType != WebhookTypeSlack && req.WebhookType != WebhookTypeDatadog && req.WebhookType != WebhookTypeCustom {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "webhook_type deve ser slack, datadog ou custom"})
	}

	// Gera secret aleatório para assinatura HMAC
	secret, err := generateSecret()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao gerar secret"})
	}

	var id string
	err = h.db.QueryRow(c.Context(), `
		INSERT INTO webhooks (project_id, url, secret, events, webhook_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.ProjectID, req.URL, secret, req.Events, string(req.WebhookType)).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao criar webhook"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           id,
		"project_id":   req.ProjectID,
		"url":          req.URL,
		"events":       req.Events,
		"webhook_type": req.WebhookType,
		"secret":       secret, // retornado apenas na criação — não armazenado em plain text no frontend
		"is_active":    true,
		"message":      "Webhook criado. Guarde o secret — ele não será exibido novamente.",
	})
}

// Delete DELETE /api/v1/webhooks/:id
func (h *WebhookHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id obrigatório"})
	}

	result, err := h.db.Exec(c.Context(), `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao remover webhook"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "webhook não encontrado"})
	}

	return c.JSON(fiber.Map{"message": "webhook removido"})
}

// Test POST /api/v1/webhooks/:id/test — dispara payload de teste e registra a entrega
func (h *WebhookHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	var url, secret, wType string
	err := h.db.QueryRow(c.Context(), `
		SELECT url, secret, webhook_type FROM webhooks WHERE id = $1
	`, id).Scan(&url, &secret, &wType)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "webhook não encontrado"})
	}

	payload := map[string]interface{}{
		"event":     "webhook.test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      map[string]string{"message": "Este é um payload de teste do Pulse"},
	}

	start := time.Now()
	dispatchErr := h.dispatchWithType(context.Background(), url, secret, WebhookType(wType), "webhook.test", payload)
	durationMs := int(time.Since(start).Milliseconds())

	// Registra entrega do teste
	go h.recordDelivery(context.Background(), id, "webhook.test", payload, dispatchErr, durationMs)

	if dispatchErr != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("falha ao entregar: %v", dispatchErr)})
	}

	return c.JSON(fiber.Map{"message": "payload de teste entregue com sucesso", "duration_ms": durationMs})
}

// ListDeliveries GET /api/v1/webhooks/:id/deliveries — histórico das últimas 50 entregas
func (h *WebhookHandler) ListDeliveries(c *fiber.Ctx) error {
	webhookID := c.Params("id")
	if webhookID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id obrigatório"})
	}

	// Verifica se o webhook existe
	var exists bool
	if err := h.db.QueryRow(c.Context(), `SELECT EXISTS(SELECT 1 FROM webhooks WHERE id = $1)`, webhookID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "webhook não encontrado"})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT id, webhook_id, event_type, status_code, success, duration_ms, error, delivered_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY delivered_at DESC
		LIMIT 50
	`, webhookID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar entregas"})
	}
	defer rows.Close()

	var deliveries []WebhookDeliveryRow
	for rows.Next() {
		var d WebhookDeliveryRow
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventType, &d.StatusCode, &d.Success, &d.DurationMs, &d.Error, &d.DeliveredAt); err != nil {
			continue
		}
		deliveries = append(deliveries, d)
	}
	if deliveries == nil {
		deliveries = []WebhookDeliveryRow{}
	}

	return c.JSON(fiber.Map{"deliveries": deliveries, "total": len(deliveries)})
}

// Dispatch envia um payload para todos os webhooks ativos de um projeto que escutam o evento.
// Deve ser chamado de forma assíncrona (go h.Dispatch(...)) pelos handlers de outros domínios.
func (h *WebhookHandler) Dispatch(ctx context.Context, projectID, eventType string, data interface{}) {
	rows, err := h.db.Query(ctx, `
		SELECT id, url, secret, webhook_type FROM webhooks
		WHERE project_id = $1 AND is_active = true AND $2 = ANY(events)
	`, projectID, eventType)
	if err != nil {
		log.Printf("[webhook] dispatch query error: %v", err)
		return
	}
	defer rows.Close()

	payload := map[string]interface{}{
		"event":      eventType,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"project_id": projectID,
		"data":       data,
	}

	for rows.Next() {
		var id, url, secret, wType string
		if err := rows.Scan(&id, &url, &secret, &wType); err != nil {
			continue
		}
		go h.deliverAndLog(ctx, id, url, secret, WebhookType(wType), eventType, payload)
	}
}

// deliverAndLog tenta entregar o webhook e registra o resultado com métricas de latência.
func (h *WebhookHandler) deliverAndLog(ctx context.Context, webhookID, url, secret string, wType WebhookType, eventType string, payload map[string]interface{}) {
	start := time.Now()
	err := h.dispatchWithType(ctx, url, secret, wType, eventType, payload)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		log.Printf("[webhook] delivery failed webhook_id=%s event=%s duration_ms=%d: %v", webhookID, eventType, durationMs, err)
	}

	h.recordDelivery(ctx, webhookID, eventType, payload, err, durationMs)
}

// recordDelivery persiste o resultado da entrega em webhook_deliveries.
func (h *WebhookHandler) recordDelivery(ctx context.Context, webhookID, eventType string, payload interface{}, deliveryErr error, durationMs int) {
	statusCode := 0
	errMsg := ""
	success := deliveryErr == nil

	if httpErr, ok := deliveryErr.(interface{ StatusCode() int }); ok {
		statusCode = httpErr.StatusCode()
	}
	if deliveryErr != nil {
		errMsg = deliveryErr.Error()
	}
	if success {
		statusCode = 200
	}

	data, _ := json.Marshal(payload)
	_, _ = h.db.Exec(ctx, `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status_code, success, duration_ms, error)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
	`, webhookID, eventType, string(data), statusCode, success, durationMs, errMsg)
}

// dispatchWithType faz o HTTP POST formatando o payload de acordo com o tipo do webhook.
func (h *WebhookHandler) dispatchWithType(ctx context.Context, url, secret string, wType WebhookType, eventType string, payload interface{}) error {
	var body []byte
	var err error

	switch wType {
	case WebhookTypeSlack:
		body, err = h.formatSlack(eventType, payload)
	case WebhookTypeDatadog:
		body, err = h.formatDatadog(eventType, payload)
	default:
		// custom — payload canônico do Pulse com HMAC
		body, err = json.Marshal(payload)
	}
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	// Assina o payload com HMAC-SHA256 (todos os tipos recebem a assinatura)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pulse-Event", eventType)
	req.Header.Set("X-Pulse-Signature", sig)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook: status %d", resp.StatusCode)
	}
	return nil
}

// dispatch mantém compatibilidade com chamadas internas que não passam tipo.
// Usa formatação custom (payload canônico Pulse).
func (h *WebhookHandler) dispatch(ctx context.Context, url, secret, eventType string, payload interface{}) error {
	return h.dispatchWithType(ctx, url, secret, WebhookTypeCustom, eventType, payload)
}

// formatSlack formata o payload no formato de Incoming Webhook do Slack.
// Converte o evento Pulse para um bloco de texto legível.
func (h *WebhookHandler) formatSlack(eventType string, payload interface{}) ([]byte, error) {
	data, _ := json.MarshalIndent(payload, "", "  ")
	slackMsg := map[string]interface{}{
		"text": fmt.Sprintf("*[Pulse]* Evento: `%s`", eventType),
		"blocks": []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*[Pulse]* Evento: `%s`\n```%s```", eventType, string(data)),
				},
			},
		},
	}
	return json.Marshal(slackMsg)
}

// formatDatadog formata o payload como evento da API de Events do Datadog.
// Ref: https://docs.datadoghq.com/api/latest/events/#post-an-event
func (h *WebhookHandler) formatDatadog(eventType string, payload interface{}) ([]byte, error) {
	text, _ := json.Marshal(payload)
	ddEvent := map[string]interface{}{
		"title":      fmt.Sprintf("Pulse: %s", eventType),
		"text":       string(text),
		"alert_type": "info",
		"tags":       []string{"source:pulse", fmt.Sprintf("event:%s", eventType)},
	}
	return json.Marshal(ddEvent)
}

// generateSecret gera um secret aleatório de 32 bytes em hex.
func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
