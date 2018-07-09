package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/crunchyroll/cx-reactions/app"
	"github.com/crunchyroll/cx-reactions/logging"
)

var configFile = flag.String("config", "config/config.yaml", "Config file path")

func main() {
	flag.Parse()
	wsApp := app.Init(*configFile)
	go wsApp.Start()

	c := make(chan os.Signal)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signal.Notify(c, signals...)
	sig := <-c
	logging.Logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	wsApp.Shutdown()
}
