package handler

// BKL-991: Streaming SSE para flag updates — elimina polling do frontend.
// Endpoint: GET /api/v1/projects/:project_id/flags/stream
//
// Quando uma flag é atualizada (via FlagSSEBroker.Notify), todos os subscribers
// do project_id recebem um evento SSE com o novo estado da flag.
//
// Formato do evento:
//   data: {"flag_key":"...","enabled":true,"updated_at":"..."}\n\n
//
// Heartbeat a cada 30s:
//   data: {"type":"heartbeat"}\n\n

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
)

// FlagUpdateEvent é o payload enviado via SSE quando uma flag é atualizada.
type FlagUpdateEvent struct {
	FlagKey   string    `json:"flag_key"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FlagSSEBroker gerencia os canais SSE por project_id.
// Thread-safe: sync.Map para o mapa externo, mutex para modificar as slices internas.
type FlagSSEBroker struct {
	mu          sync.Mutex
	subscribers map[string][]chan FlagUpdateEvent
}

// NewFlagSSEBroker cria um broker SSE.
func NewFlagSSEBroker() *FlagSSEBroker {
	return &FlagSSEBroker{
		subscribers: make(map[string][]chan FlagUpdateEvent),
	}
}

// Subscribe registra um novo canal de eventos para um project_id.
// Retorna o canal e uma função cleanup a ser chamada no disconnect.
func (b *FlagSSEBroker) Subscribe(projectID string) (chan FlagUpdateEvent, func()) {
	ch := make(chan FlagUpdateEvent, 16)

	b.mu.Lock()
	b.subscribers[projectID] = append(b.subscribers[projectID], ch)
	b.mu.Unlock()

	cleanup := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		chans := b.subscribers[projectID]
		filtered := chans[:0]
		for _, c := range chans {
			if c != ch {
				filtered = append(filtered, c)
			}
		}
		b.subscribers[projectID] = filtered
		close(ch)
	}
	return ch, cleanup
}

// Notify envia um evento para todos os subscribers de um project_id.
// Non-blocking por subscriber: se o canal estiver cheio, o evento é descartado.
func (b *FlagSSEBroker) Notify(projectID string, event FlagUpdateEvent) {
	b.mu.Lock()
	chans := make([]chan FlagUpdateEvent, len(b.subscribers[projectID]))
	copy(chans, b.subscribers[projectID])
	b.mu.Unlock()

	for _, ch := range chans {
		select {
		case ch <- event:
		default:
			log.Printf("[sse] canal cheio para project_id=%s — evento descartado", projectID)
		}
	}
}

// FlagSSEHandler serve o endpoint SSE de flag updates.
type FlagSSEHandler struct {
	broker       *FlagSSEBroker
	serviceToken string
}

// NewFlagSSEHandler cria o handler SSE.
func NewFlagSSEHandler(broker *FlagSSEBroker, serviceToken string) *FlagSSEHandler {
	return &FlagSSEHandler{broker: broker, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas SSE.
func (h *FlagSSEHandler) RegisterRoutes(app *fiber.App) {
	// Rota separada do grupo de projetos para não colidir com /:id/health-config
	app.Get("/api/v1/projects/:project_id/flags/stream",
		middleware.ServiceTokenMiddleware(h.serviceToken),
		h.Stream,
	)
}

// Stream godoc
// GET /api/v1/projects/:project_id/flags/stream
// Abre uma conexão SSE. Eventos chegam sempre que uma flag do projeto é atualizada.
// Heartbeat a cada 30s para manter a conexão viva através de proxies/load balancers.
func (h *FlagSSEHandler) Stream(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // desabilita buffering nginx/traefik

	ch, cleanup := h.broker.Subscribe(projectID)

	// fasthttp expõe SetBodyStreamWriter(func(w *bufio.Writer))
	// O *bufio.Writer aqui é do pacote bufio padrão do Go.
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cleanup()

		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		// Comentário inicial confirma conexão ao cliente
		_, _ = fmt.Fprintf(w, ": connected project_id=%s\n\n", projectID)
		_ = w.Flush()

		for {
			select {
			case <-c.Context().Done():
				return

			case event, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(event)
				if err != nil {
					log.Printf("[sse] marshal error: %v", err)
					continue
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}

			case <-heartbeat.C:
				if _, err := fmt.Fprint(w, "data: {\"type\":\"heartbeat\"}\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}
