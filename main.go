package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/crunchyroll/cx-reactions/endpoints"
	"github.com/crunchyroll/cx-reactions/hub"
)

// TODO: extract configs CORE-107
var addr = flag.String("addr", ":8080", "server address")

func main() {
	flag.Parse()
	emojiHub := hub.NewHub()
	emojiHub.Start()
	router := endpoints.NewRouter(emojiHub)
	// TODO: implement graceful shutdown of the server/app CORE-110
	server := http.Server{
		Addr:         *addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Print("Listening")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
	defer server.Close()
}
