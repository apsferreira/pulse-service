package handler

import (
	"crypto/subtle"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/pkg/wshub"
)

// WSHandler gerencia conexões WebSocket para o dashboard real-time
type WSHandler struct {
	hub          *wshub.Hub
	serviceToken string
}

// NewWSHandler cria um novo WSHandler
func NewWSHandler(hub *wshub.Hub, serviceToken string) *WSHandler {
	return &WSHandler{hub: hub, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas WebSocket no app Fiber
func (h *WSHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Upgrade middleware — rejeita conexões não-WebSocket antes de chegar ao handler
	api.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// GET /api/v1/ws?project_id=xxx&token=xxx
	// Autenticação via query param token — browsers não suportam headers customizados no WS upgrade.
	api.Get("/ws", websocket.New(h.handleConnection))
}

// handleConnection gerencia o ciclo de vida de uma conexão WebSocket
func (h *WSHandler) handleConnection(c *websocket.Conn) {
	// Autenticação via query param token (SERVICE_TOKEN)
	// Browsers não suportam headers customizados no WebSocket upgrade
	token := c.Query("token", "")
	if subtle.ConstantTimeCompare([]byte(token), []byte(h.serviceToken)) != 1 {
		log.Printf("ws: autenticação falhou — token inválido")
		_ = c.Close()
		return
	}

	projectID := c.Query("project_id", "")

	client := wshub.NewClient(c, projectID, h.hub)
	defer h.hub.Unregister(client)

	// Lê mensagens do cliente (necessário para detectar disconnect e pong frames)
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: conexão encerrada inesperadamente project_id=%s", projectID)
			}
			break
		}
	}
}
