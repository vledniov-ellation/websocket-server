// WARNING: This is not deployment-ready code until database usage implemented and the code is able to work on multiple distributed machines.
package hub

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/crunchyroll/cx-reactions/model"
)

const (
	broadcastDuration = 10 * time.Second
)

// TODO: Add tests CORE-108
type Hub struct {
	clients   map[*Client]bool
	broadcast chan model.Emoji
	mux       sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan model.Emoji),
	}
}

func (h *Hub) Start() {
	go h.tick()
	go h.broadcastMessages()
}

// TODO: remove method after logging implemented CORE-109
func (h *Hub) tick() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			log.Print("CURRENTLY CONNECTED CLIENTS: ", h.GetSubscribedNumber())
		}
	}
}

func (h *Hub) broadcastMessages() {
	ticker := time.NewTicker(broadcastDuration)
	var messagesToBroadcast []model.Emoji

	for {
		select {
		case message := <-h.broadcast:
			messagesToBroadcast = append(messagesToBroadcast, message)
		case <-ticker.C:
			result := computeEmojiStats(messagesToBroadcast)
			result.Visitors = h.GetSubscribedNumber()
			log.Println(fmt.Sprintf("SENDING BROADCAST %d EMOJI TYPES TO %d CLIENTS", len(result.Items), result.Visitors))
			messagesToBroadcast = messagesToBroadcast[:0]

			h.mux.RLock()
			for client := range h.clients {
				client.Send <- result
			}
			h.mux.RUnlock()
		}
	}
}

func (h *Hub) RegisterConn(conn *websocket.Conn) {
	client := &Client{Hub: h, Conn: conn, Send: make(chan model.EmojiStats)}
	h.mux.Lock()
	h.clients[client] = true
	h.mux.Unlock()

	// Client should start reading and writing
	go client.Run()
}

func (h *Hub) RegisterMessage(msg model.Emoji) {
	h.broadcast <- msg
}

func (h *Hub) GetSubscribedNumber() int {
	h.mux.RLock()
	clientCount := len(h.clients)
	h.mux.RUnlock()
	return clientCount
}

func (h *Hub) DisconnectSubscriber(client *Client) {
	log.Println("TRYING TO DISCONNECT CLIENT ", client.ID)
	h.mux.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
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
