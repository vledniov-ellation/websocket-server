package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/crunchyroll/cx-reactions/config"
	"github.com/crunchyroll/cx-reactions/endpoints"
	"github.com/crunchyroll/cx-reactions/hub"
	"github.com/crunchyroll/cx-reactions/logging"
)

// Reactions represents reactions application
type Reactions struct {
	server http.Server
	hub    *hub.Hub
}

// Init initializes config, logger and the app
func Init(configFilePath string) *Reactions {
	config.Init(configFilePath)
	err := logging.Init(logSettings())
	if err != nil {
		panic(fmt.Sprintf("Could not initialize logger: %v", err))
	}
	return &Reactions{hub: hub.NewHub()}
}

// Start starts reactions app
func (rx *Reactions) Start() {
	rx.hub.Start()
	upgrader := websocket.Upgrader{
		HandshakeTimeout: config.HandshakeTimeout(),
		ReadBufferSize:   config.ReadBufferSize(),
		WriteBufferSize:  config.WriteBufferSize(),
		CheckOrigin:      func(r *http.Request) bool { return true },
	}

	router := endpoints.NewRouter(rx.hub, upgrader)
	rx.server = http.Server{
		Addr:         config.Listen(),
		Handler:      router,
		ReadTimeout:  config.ServerReadTimeout(),
		WriteTimeout: config.ServerWriteTimeout(),
	}

	logging.Logger.Info("Starting server")
	err := rx.server.ListenAndServe()
	if err == http.ErrServerClosed {
		logging.Logger.Info("http server is closed")
	} else {
		logging.Logger.Fatal("http server failed", zap.Error(err))
	}
}

// Shutdown gracefully shuts down reactions app
func (rx *Reactions) Shutdown() {
	defer logging.Logger.Sync()
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout())
	defer cancel()
	err := rx.server.Shutdown(ctx)
	if err != nil {
		logging.Logger.Error("error closing server", zap.Error(err))
	}
	rx.hub.Shutdown()

	logging.Logger.Info("reactions service was shut down")
}

// LogSettings represent the settings for logger
func logSettings() logging.Settings {
	return logging.Settings{
		Level:         config.LogLevel(),
		Output:        config.LogOutput(),
		LogCaller:     config.ShouldLogCaller(),
		LogStacktrace: config.ShouldLogStacktrace(),
	}
}
