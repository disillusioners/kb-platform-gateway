package sse

import (
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	running bool
}

func NewHub() *Hub {
	return &Hub{
		running: true,
	}
}

func (h *Hub) Run() {}

func (h *Hub) Broadcast(event SSEEvent) {
}

func (h *Hub) Send(client interface{}, event SSEEvent) {
}

func (h *Hub) AddClient(client interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
}

func (h *Hub) RemoveClient(client interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
}
