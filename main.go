package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/crunchyroll/cx-reactions/config"
	"github.com/crunchyroll/cx-reactions/endpoints"
	"github.com/crunchyroll/cx-reactions/hub"
	"github.com/crunchyroll/cx-reactions/logging"
)

var configFile = flag.String("config", "config/config.yaml", "Config file path")

func main() {
	flag.Parse()
	config.InitConfig(*configFile)
	err := logging.Init(logSettings())
	if err != nil {
		panic(fmt.Sprintf("Could not initialize logger: %v", err))
	}

	emojiHub := hub.NewHub()
	emojiHub.Start()
	router := endpoints.NewRouter(emojiHub)
	// TODO: implement graceful shutdown of the server/app CORE-110
	server := http.Server{
		Addr:         config.Listen(),
		Handler:      router,
		ReadTimeout:  config.ServerReadTimeout(),
		WriteTimeout: config.ServerWriteTimeout(),
	}

	logging.Logger.Info("Starting server")
	err = server.ListenAndServe()
	if err != nil {
		logging.Logger.Fatal("Error starting server: " + err.Error())
	}
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
