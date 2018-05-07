package main

import (
	"flag"
	"net/http"

	"log"
	"github.com/gorilla/mux"
	"time"
)

var addr = flag.String("addr", ":8080", "server address")

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	router := newRouter(hub)
	server := http.Server{
		Addr: *addr,
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Print("Listening")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
	defer server.Close()
}

func newRouter(hub *Hub) *mux.Router {
	router := mux.NewRouter()
	router.Methods(http.MethodGet).
		Path("/ws").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleWS(hub, w, r)
		}).Name("websocket")

	return router
}