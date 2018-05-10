package main

import (
	"time"
	"log"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	disconnect chan *Client
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

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.disconnect:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
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