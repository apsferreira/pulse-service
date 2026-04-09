package wshub

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// Client representa uma conexão WebSocket ativa
type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	ProjectID string
}

// BroadcastMessage é a mensagem enviada pelo hub para os clientes
type BroadcastMessage struct {
	// ProjectID vazio significa broadcast para todos os clientes
	ProjectID string
	Data      []byte
}

// Hub gerencia todas as conexões WebSocket ativas
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// New cria um novo Hub
func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run inicia o loop de eventos do hub. Deve ser executado em goroutine separada.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("wshub: cliente registrado project_id=%s total_clientes=%d", client.ProjectID, h.clientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("wshub: cliente desregistrado project_id=%s total_clientes=%d", client.ProjectID, h.clientCount())

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.ProjectID == msg.ProjectID || msg.ProjectID == "" {
					select {
					case client.send <- msg.Data:
					default:
						// Canal cheio — remove cliente lento
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast envia uma mensagem para todos os clientes de um projeto
func (h *Hub) Broadcast(projectID string, data []byte) {
	h.broadcast <- BroadcastMessage{ProjectID: projectID, Data: data}
}

// BroadcastAll envia uma mensagem para todos os clientes conectados
func (h *Hub) BroadcastAll(data []byte) {
	h.broadcast <- BroadcastMessage{ProjectID: "", Data: data}
}

// clientCount retorna o número de clientes conectados (sem lock — caller deve ter o lock)
func (h *Hub) clientCount() int {
	return len(h.clients)
}

// writePump envia mensagens do canal send para a conexão WebSocket
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("wshub: erro ao escrever mensagem project_id=%s", c.ProjectID)
			return
		}
	}
}

// NewClient cria um Client e inicia a goroutine de escrita
func NewClient(conn *websocket.Conn, projectID string, hub *Hub) *Client {
	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 64),
		ProjectID: projectID,
	}
	hub.register <- client
	go client.writePump()
	return client
}

// Unregister remove o cliente do hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
