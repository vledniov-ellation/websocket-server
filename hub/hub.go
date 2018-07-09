// Package hub defines the environment of creating a hub which works with multiple connected clients via websocket
// WARNING: This is not deployment-ready code until database usage implemented and the code is able to work on multiple distributed machines.
package hub

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/crunchyroll/cx-reactions/logging"
	"github.com/crunchyroll/cx-reactions/model"
)

const (
	broadcastDuration = 10 * time.Second
)

// Hub stores data about clients connected to it and message broadcasting to them
// TODO: Add tests CORE-108
type Hub struct {
	clients   map[*client]bool
	broadcast chan model.Emoji
	mux       sync.RWMutex
}

// NewHub instantiates a new empty hub
func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*client]bool),
		broadcast: make(chan model.Emoji),
	}
}

// Start launches the hub
func (h *Hub) Start() {
	go h.tick()
	go h.broadcastMessages()
}

// Shutdown gracefully shutdowns the hub
func (h *Hub) Shutdown() {
	wg := sync.WaitGroup{}
	var clients []*client

	h.mux.RLock()
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mux.RUnlock()

	wg.Add(len(clients))
	for _, c := range clients {
		go func(client *client) {
			client.Shutdown()
			wg.Done()
		}(c)
	}

	wg.Wait()
	close(h.broadcast)
}

// TODO: remove method after logging implemented CORE-109
func (h *Hub) tick() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			logging.Logger.Info("Currently connected clients", zap.Int("client_count", h.GetSubscribedNumber()))
		}
	}
}

func (h *Hub) broadcastMessages() {
	ticker := time.NewTicker(broadcastDuration)
	var messagesToBroadcast []model.Emoji

	for {
		select {
		case message, ok := <-h.broadcast:
			if !ok {
				logging.Logger.Info("Broadcast channel closed, quitting broadcast")
				return
			}
			messagesToBroadcast = append(messagesToBroadcast, message)
		case <-ticker.C:
			result := computeEmojiStats(messagesToBroadcast)
			result.Visitors = h.GetSubscribedNumber()
			logging.Logger.Info(
				"Sending broadcast",
				zap.Int("item_count", len(result.Items)),
				zap.Int("client_count", result.Visitors),
			)
			messagesToBroadcast = messagesToBroadcast[:0]

			h.mux.RLock()
			for client := range h.clients {
				client.send <- result
			}
			h.mux.RUnlock()
		}
	}
}

// RegisterConn registers a new connection in the hub
func (h *Hub) RegisterConn(conn *websocket.Conn) {
	id := uuid.New().String()
	logging.Logger.Debug("Registering client", zap.String("client_id", id))
	client := &client{
		hub:  h,
		conn: conn,
		send: make(chan model.EmojiStats),
		ID:   id,
	}
	h.mux.Lock()
	h.clients[client] = true
	h.mux.Unlock()

	// Client should start reading and writing
	go client.Run()
}

// RegisterMessage creates a message for broadcasting to clients
func (h *Hub) RegisterMessage(msg model.Emoji) {
	logging.Logger.Debug("Registered message", zap.String("message", msg.Type))
	h.broadcast <- msg
}

// GetSubscribedNumber returns the total number of connections in the hub
func (h *Hub) GetSubscribedNumber() int {
	h.mux.RLock()
	clientCount := len(h.clients)
	h.mux.RUnlock()
	return clientCount
}

// DisconnectSubscriber removes a connection from the hub
func (h *Hub) DisconnectSubscriber(subscriber *client) {
	logging.Logger.Info("Disconnecting Client", zap.String("client_id", subscriber.ID))
	h.mux.Lock()
	if _, ok := h.clients[subscriber]; ok {
		delete(h.clients, subscriber)
	}
	h.mux.Unlock()
}

// groupMessages groups the message by type and the count of such messages.
// e.g. {"smiling_face": 23, "crossed_swords": 12}
func computeEmojiStats(messages []model.Emoji) model.EmojiStats {
	grouped := map[string]int{}
	result := model.EmojiStats{Items: []model.Emoji{}}

	for _, msg := range messages {
		grouped[msg.Type]++
	}

	for msgType, count := range grouped {
		result.Items = append(result.Items, model.Emoji{Type: msgType, Count: count})
	}

	return result
}
