package realtime

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]*Client),
	}
}

func (h *Hub) Register(driverID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.clients[driverID]; ok {
		old.Close()
	}

	h.clients[driverID] = c
	log.Println("driver websocket connected:", driverID)
}

func (h *Hub) Unregister(driverID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if current, ok := h.clients[driverID]; ok && current == c {
		delete(h.clients, driverID)
		log.Println("driver websocket disconnected:", driverID)
	}
}

func (h *Hub) SendToDriver(ctx context.Context, driverID uuid.UUID, msg OutgoingMessage) bool {
	h.mu.RLock()
	client, ok := h.clients[driverID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	return client.Send(ctx, msg)
}
