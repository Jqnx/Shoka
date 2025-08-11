package websocket

import (
	"Shoka/internal/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	clientsMux sync.RWMutex
	upgrader   websocket.Upgrader
	log        logger.Logger
}

func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		log:     logger,
		clients: make(map[*Client]bool),
		// TODO: change CheckOrigin for prod
		// https://pkg.go.dev/github.com/gorilla/websocket#Upgrader
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

func (h *Hub) AddClient(client *Client) {
	h.clientsMux.Lock()
	h.clients[client] = true
	h.clientsMux.Unlock()
}

func (h *Hub) RemoveClient(client *Client) {
	h.clientsMux.Lock()
	delete(h.clients, client)
	h.clientsMux.Unlock()
}

func (h *Hub) Broadcast(message []byte) {
	h.clientsMux.RLock()
	clientsCopy := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clientsCopy = append(clientsCopy, client)
	}
	h.clientsMux.RUnlock()

	var deadClients []*Client
	for _, client := range clientsCopy {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			deadClients = append(deadClients, client)
			client.Close()
		}
	}

	if len(deadClients) > 0 {
		h.clientsMux.Lock()
		for _, client := range deadClients {
			delete(h.clients, client)
		}
		h.clientsMux.Unlock()
	}
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("failed to upgrade websocket", "error", err)
		return
	}

	client := &Client{conn: conn}
	defer client.Close()

	h.AddClient(client)

	defer h.RemoveClient(client)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				h.log.Error("issue with websocket", "error", err)
			}
			break
		}
	}
}
