package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/crunchyroll/cx-reactions/logging"
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

// Core defines an interface for the messaging operator the client belongs to
type core interface {
	RegisterMessage(model.Emoji)
	DisconnectSubscriber(*client)
}

// Client represents a subscriber via websocket
type client struct {
	hub  core
	conn *websocket.Conn
	send chan model.EmojiStats
	wg   sync.WaitGroup
	once sync.Once
	ID   string
}

func (c *client) readPipe() {
	logger := logging.Logger.With(zap.String("client_id", c.ID))
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		logger.Debug("Received message from client", zap.ByteString("message", message))

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				logger.Warn("Unexpected websocket connection lost: " + err.Error())
			} else {
				logger.Warn("Client closed connection: " + err.Error())
			}
			break
		}
		var incoming model.Emoji
		if err = json.Unmarshal(message, &incoming); err != nil {
			logger.Error("Could not unmarshal message", zap.Error(err))
			break
		}
		c.hub.RegisterMessage(incoming)
	}
}

func (c *client) writePipe() {
	logger := logging.Logger.With(zap.String("client_id", c.ID))
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				logger.Warn("Client sending channel was closed")
				return
			}
			logger.Debug("Sending message to client", zap.Any("message", message))

			msg, err := json.Marshal(message)
			if err != nil {
				logger.With(zap.Any("message", message)).
					Error("Could not marshal message to client: " + err.Error())

				return
			}

			err = c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Error("Could not write message to websocket: " + err.Error())
				return
			}

		case <-ticker.C:
			logger.Debug("Sending ping")
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				logger.Info("Ping failed", zap.Error(err))
				return
			}
		}
	}
}

// Shutdown executes a closing procedure for the client
func (c *client) Shutdown() {
	logger := logging.Logger.With(zap.String("client_id", c.ID))
	c.once.Do(func() {
		logger.Debug("Closing client resources")
		c.hub.DisconnectSubscriber(c)
		c.conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait))
		c.conn.Close()
		close(c.send)
		logger.Debug("Client closed")
	})
	c.wg.Wait()
	logger.Debug("Client shutdown")
}

// Run launched the reading and writing pipes through the websocket for the client
func (c *client) Run() {
	logger := logging.Logger.With(zap.String("client_id", c.ID))
	done := make(chan struct{}, 2)
	c.wg.Add(2)

	go func() {
		logger.Debug("Opening Read Pipe for client")
		c.readPipe()
		logger.Debug("Read Pipe stopped")
		c.wg.Done()
		done <- struct{}{}
	}()
	go func() {
		logger.Debug("Opening Write Pipe for client")
		c.writePipe()
		logger.Debug("Write Pipe stopped")
		c.wg.Done()
		done <- struct{}{}
	}()

	<-done
	c.Shutdown()
	<-done
}
