package handler

// BKL-1078: Webhooks para eventos de produto do Pulse.
// Registra URLs externas e dispara payloads assinados com HMAC-SHA256
// quando eventos relevantes ocorrem (event.tracked, flag.evaluated, health_score.dropped).
// Compatível com n8n, Make, Zapier.

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
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookHandler gerencia o registro e disparo de webhooks.
type WebhookHandler struct {
	db           *pgxpool.Pool
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

// RegisterRoutes registra as rotas de webhooks.
func (h *WebhookHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/webhooks", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/", h.List)
	api.Post("/", h.Create)
	api.Delete("/:id", h.Delete)
	api.Post("/:id/test", h.Test)
}

// WebhookRow representa um webhook no banco.
type WebhookRow struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookCreateRequest é o payload para POST /api/v1/webhooks.
type WebhookCreateRequest struct {
	ProjectID string   `json:"project_id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"` // ex: ["event.tracked","flag.evaluated"]
}

// List GET /api/v1/webhooks?project_id=...
func (h *WebhookHandler) List(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT id, project_id, url, events, is_active, created_at
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
		if err := rows.Scan(&w.ID, &w.ProjectID, &w.URL, &w.Events, &w.IsActive, &w.CreatedAt); err != nil {
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
	if len(req.Events) == 0 {
		req.Events = []string{"event.tracked", "flag.evaluated", "health_score.dropped"}
	}

	// Gera secret aleatório para assinatura HMAC
	secret, err := generateSecret()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao gerar secret"})
	}

	var id string
	err = h.db.QueryRow(c.Context(), `
		INSERT INTO webhooks (project_id, url, secret, events)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.ProjectID, req.URL, secret, req.Events).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao criar webhook"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         id,
		"project_id": req.ProjectID,
		"url":        req.URL,
		"events":     req.Events,
		"secret":     secret, // retornado apenas na criação — não armazenado em plain text no frontend
		"is_active":  true,
		"message":    "Webhook criado. Guarde o secret — ele não será exibido novamente.",
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

// Test POST /api/v1/webhooks/:id/test — dispara payload de teste
func (h *WebhookHandler) Test(c *fiber.Ctx) error {
	id := c.Params("id")

	var url, secret string
	err := h.db.QueryRow(c.Context(), `SELECT url, secret FROM webhooks WHERE id = $1`, id).Scan(&url, &secret)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "webhook não encontrado"})
	}

	payload := map[string]interface{}{
		"event":     "webhook.test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      map[string]string{"message": "Este é um payload de teste do Pulse"},
	}

	if err := h.dispatch(context.Background(), url, secret, "webhook.test", payload); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("falha ao entregar: %v", err)})
	}

	return c.JSON(fiber.Map{"message": "payload de teste entregue com sucesso"})
}

// Dispatch envia um payload para todos os webhooks ativos de um projeto que escutam o evento.
// Deve ser chamado de forma assíncrona (go h.Dispatch(...)) pelos handlers de outros domínios.
func (h *WebhookHandler) Dispatch(ctx context.Context, projectID, eventType string, data interface{}) {
	rows, err := h.db.Query(ctx, `
		SELECT id, url, secret FROM webhooks
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
		var id, url, secret string
		if err := rows.Scan(&id, &url, &secret); err != nil {
			continue
		}
		go h.deliverAndLog(ctx, id, url, secret, eventType, payload)
	}
}

// deliverAndLog tenta entregar o webhook e registra o resultado.
func (h *WebhookHandler) deliverAndLog(ctx context.Context, webhookID, url, secret, eventType string, payload map[string]interface{}) {
	err := h.dispatch(ctx, url, secret, eventType, payload)
	statusCode := 0
	errMsg := ""

	if httpErr, ok := err.(interface{ StatusCode() int }); ok {
		statusCode = httpErr.StatusCode()
	}
	if err != nil {
		errMsg = err.Error()
		log.Printf("[webhook] delivery failed webhook_id=%s event=%s: %v", webhookID, eventType, err)
	}

	data, _ := json.Marshal(payload)
	_, _ = h.db.Exec(ctx, `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status_code, error)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
	`, webhookID, eventType, string(data), statusCode, errMsg)
}

// dispatch faz o HTTP POST com assinatura HMAC-SHA256.
func (h *WebhookHandler) dispatch(ctx context.Context, url, secret, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	// Assina o payload com HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
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

// generateSecret gera um secret aleatório de 32 bytes em hex.
func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
