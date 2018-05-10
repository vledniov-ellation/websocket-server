package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read next pong message from peer.
	pongWait = 10 * time.Second

	// Sends ping to peer in this interval. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 8 * time.Second,
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {return true},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan string
}

func (c *Client) readPipe() {
	defer func() {
		c.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Print("Client closed connection: "+ err.Error())
			}
			break
		}
		var incoming Message
		json.Unmarshal(message, &incoming)
		c.hub.broadcast <- incoming.Body
	}
}

func (c *Client) writePipe() {
	ticker := time.NewTicker(pingPeriod)
	// Need to close connection in case something happened and we cannot send any messages
	// and the ticker failed to do his job
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				log.Print("Client sending channel was closed")
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, serializeMessage(message))
			if err != nil {
				log.Print("Could not write message: "+err.Error())
				return
			}

			// Send the rest of the queued messages to client
			messages := len(c.send)
			for i := 0; i < messages; i++ {
				err = c.conn.WriteMessage(websocket.TextMessage, serializeMessage(<-c.send))
				if err != nil {
					log.Print("Writing broadcast messages failed: "+err.Error())
					return
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	c.conn.Close()
	c.hub.disconnect <- c
}

func serializeMessage(message string) []byte {
	jsonMsg, err := json.Marshal(Message{Body: message})
	if err != nil {
		log.Fatal("Could not marshal message", message)
	}
	return jsonMsg
}

func handleWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	//log.Print(
	//	"UPGRADE ", r.Header.Get("Upgrade"),
	//	" CONNECTION ", r.Header.Get("Connection"),
	//	" CLIENT NUMBER ", len(hub.clients)+1,
	//)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading to websocket "+ err.Error())
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan string)}
	hub.register <-client

	// Client should start reading and writing
	go client.readPipe()
	go client.writePipe()
}
