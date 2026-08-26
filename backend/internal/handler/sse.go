package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/collector"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/isp"
)

// SSEHub manages real-time Server-Sent Events subscribers.
type SSEHub struct {
	clients    map[chan string]bool
	register   chan chan string
	unregister chan chan string
	broadcast  chan string
	mu         sync.RWMutex
	pinger     *isp.Pinger
	speedtest  *isp.SpeedtestRunner
}

// NewSSEHub initializes the SSE hub.
func NewSSEHub(pinger *isp.Pinger, speedtest *isp.SpeedtestRunner) *SSEHub {
	return &SSEHub{
		clients:    make(map[chan string]bool),
		register:   make(chan chan string),
		unregister: make(chan chan string),
		broadcast:  make(chan string, 100),
		pinger:     pinger,
		speedtest:  speedtest,
	}
}

// Run starts the event dispatch loop.
func (h *SSEHub) Run() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- msg:
				default:
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.broadcastVitals()
		}
	}
}

func (h *SSEHub) broadcastVitals() {
	if h.pinger == nil {
		return
	}
	vitals := h.pinger.GetVitals()
	var speedRec interface{}
	if h.speedtest != nil {
		speedRec = h.speedtest.GetLastRecord()
	}

	payload := map[string]interface{}{
		"type": "vitals",
		"data": map[string]interface{}{
			"vitals":    vitals,
			"speedtest": speedRec,
		},
	}
	b, _ := json.Marshal(payload)
	h.broadcast <- string(b)
}

// BroadcastQuery sends a real-time DNS event to all connected dashboard clients.
func (h *SSEHub) BroadcastQuery(ev collector.DNSQueryEvent) {
	payload := map[string]interface{}{
		"type": "query",
		"data": ev,
	}
	b, _ := json.Marshal(payload)
	h.broadcast <- string(b)
}

// ServeHTTP handles incoming SSE connections on /api/v1/stream.
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string, 20)
	h.register <- clientChan

	defer func() {
		h.unregister <- clientChan
	}()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
