package main

import (
	"time"
	"log"
	"sync"
)

const broadcastDuration = 5 * time.Second

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	disconnect chan *Client
	mux 	   sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		broadcast: make(chan string),
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
			if _, ok := h.clients[client]; ok {
				h.mux.Lock()
				delete(h.clients, client)
				close(client.send)
				h.mux.Unlock()
			}
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
	var messagesToBroadcast []string

	for {
		select {
		case message := <-h.broadcast:
			broadcastMutex.Lock()
			messagesToBroadcast = append(messagesToBroadcast, message)
			broadcastMutex.Unlock()
		case <-ticker.C:
			broadcastMutex.Lock()
			messagesBuffer := append([]string{}, messagesToBroadcast...)
			messagesToBroadcast = messagesToBroadcast[:0]
			broadcastMutex.Unlock()

			h.mux.Lock()
			for _, message := range messagesBuffer {
				for client := range h.clients {
					client.send <- message
				}
			}
			h.mux.Unlock()
		}
	}
}