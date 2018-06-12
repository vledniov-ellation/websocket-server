package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/crunchyroll/cx-reactions/model"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read next pong message from peer.
	pongWait = 20 * time.Second

	// Sends ping to peer in this interval. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Core interface {
	RegisterMessage(model.Emoji)
	DisconnectSubscriber(*Client)
}

type Client struct {
	Hub  Core
	Conn *websocket.Conn
	Send chan model.EmojiStats
	ID   int
}

func (c *Client) readPipe(stop chan struct{}) {
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		select {
		case <-stop:
			return
		default:
		}
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Print("Unexpected websocket connection lost: " + err.Error())
			} else {
				log.Print("Client closed connection: " + err.Error())
			}
			break
		}
		var incoming model.Emoji
		if err = json.Unmarshal(message, &incoming); err == nil {
			log.Print("Could not unmrshal message")
			break
		}
		c.Hub.RegisterMessage(incoming)
	}
}

func (c *Client) writePipe(stop chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case message, ok := <-c.Send:
			if !ok {
				log.Print("Client sending channel was closed")
				return
			}

			msg, err := json.Marshal(message)
			if err != nil {
				log.Println("Could not marshal message: ", message)
				return
			}

			err = c.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Print("Could not write message: " + err.Error())
				return
			}

		case <-ticker.C:
			if err := c.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	c.Hub.DisconnectSubscriber(c)
	c.Conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait))
	c.Conn.Close()
	close(c.Send)
}

func (c *Client) Run() {
	done := make(chan struct{}, 2)
	stop := make(chan struct{})

	go func() {
		c.readPipe(stop)
		done <- struct{}{}
	}()
	go func() {
		c.writePipe(stop)
		done <- struct{}{}
	}()

	<-done
	close(stop)
	<-done

	c.Close()
}
