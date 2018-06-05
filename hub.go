package main

import (
	"time"
	"log"
	"sync"
	"fmt"
)

const broadcastDuration = 10 * time.Second

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	disconnect chan *Client
	mux 	   sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		broadcast: make(chan *Message),
		register: make(chan *Client),
		disconnect: make(chan *Client),
	}
}

func (h *Hub) run() {
	go h.Tick()
	go h.Broadcast()

	for {
		select {
		case client := <-h.register:
			log.Println("CLIENT REGISTERED")
			h.clients[client] = true
		case client := <-h.disconnect:
			log.Println("TRYING TO DISCONNECT CLIENT ", client.ID)
			h.mux.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mux.Unlock()
		}
	}
}

func (h *Hub) Tick() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			log.Print("CURRENTLY CONNECTED CLIENTS: ", len(h.clients))
		}
	}
}

func (h *Hub) Broadcast() {
	broadcastMutex := &sync.Mutex{}
	ticker := time.NewTicker(broadcastDuration)
	var messagesToBroadcast []Message

	for {
		select {
		case message := <-h.broadcast:
			broadcastMutex.Lock()
			messagesToBroadcast = append(messagesToBroadcast, *message)
			broadcastMutex.Unlock()
		case <-ticker.C:
			broadcastMutex.Lock()
			result := groupMessages(messagesToBroadcast)
			result.Visitors = len(h.clients)
			log.Println(fmt.Sprintf("SENDING BROADCAST %d EMOJI TYPES TO %d CLIENTS", len(result.Items), result.Visitors))
			messagesToBroadcast = messagesToBroadcast[:0]
			broadcastMutex.Unlock()

			h.mux.Lock()
			for client := range h.clients {
				client.send <- result
			}
			h.mux.Unlock()
		}
	}
}

func groupMessages(messages []Message) *Messages {
	result := Messages{Items: []*Message{}}

	for _, msg := range messages {
		emojiPresent := false
		for _, val := range result.Items {
			if msg.Type == val.Type {
				val.Count, emojiPresent = val.Count + 1, true
				break
			}
		}

		if !emojiPresent {
			result.Items = append(result.Items, &Message{Type: msg.Type, Count: 1})
		}
	}
	return &result
}
