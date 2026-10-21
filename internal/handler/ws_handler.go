package handler

import (
	"log"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/pkg/wshub"
)

// WSHandler gerencia conexões WebSocket para o dashboard real-time
type WSHandler struct {
	hub           *wshub.Hub
	authenticate  func(token string) bool
	allowOrigins  []string
}

// NewWSHandler cria um novo WSHandler.
//
// dashboardOrigins é a lista (CSV) de origens permitidas no handshake —
// normalmente cfg.DashboardOrigins. Vazio = sem restrição de Origin
// (comportamento anterior à correção de segurança; configure em produção).
func NewWSHandler(hub *wshub.Hub, serviceToken string, dashboardOrigins string) *WSHandler {
	return &WSHandler{
		hub:          hub,
		authenticate: middleware.NewServiceOrJWTAuthenticator(serviceToken),
		allowOrigins: splitCSV(dashboardOrigins),
	}
}

// splitCSV converte uma lista CSV de origens em []string, removendo espaços.
// Vazio retorna nil — gofiber/contrib/websocket trata nil/vazio como "*".
func splitCSV(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
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
	// SEC: Origins restringe o handshake às origens do dashboard (PULSE_DASHBOARD_ORIGINS);
	// sem essa config, o cliente `fasthttp/websocket` padrão aceitaria qualquer Origin.
	api.Get("/ws", websocket.New(h.handleConnection, websocket.Config{Origins: h.allowOrigins}))
}

// handleConnection gerencia o ciclo de vida de uma conexão WebSocket
func (h *WSHandler) handleConnection(c *websocket.Conn) {
	// Autenticação via query param token.
	// Browsers não suportam headers customizados no WebSocket upgrade, então o token
	// viaja na query string — por isso aceitamos tanto o token mestre de serviço
	// (uso interno/service-to-service) quanto um JWT de sessão do dashboard (curta
	// duração, iss/aud validados) com a MESMA regra usada pela API REST equivalente
	// (ServiceTokenMiddleware). Isso evita expor o token mestre de serviço — que também
	// protege endpoints administrativos sensíveis (billing, license, user export) —
	// como único meio de autenticar o canal realtime do browser.
	token := c.Query("token", "")
	if !h.authenticate(token) {
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
