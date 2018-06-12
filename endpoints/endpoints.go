package endpoints

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Hub interface {
	RegisterConn(conn *websocket.Conn)
	GetSubscribedNumber() int
}

// TODO: Extract configs CORE-107
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 8 * time.Second,
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

func NewRouter(h Hub) *mux.Router {
	router := mux.NewRouter()
	router.Methods(http.MethodGet).
		Path("/ws").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleWS(h, w, r)
		}).Name("websocket")

	router.Methods(http.MethodGet).
		Path("/stats").
		HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handleStats(h, w)
		})
	return router
}

// TODO: Add tests CORE-108
func handleWS(h Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading to websocket " + err.Error())
		return
	}
	h.RegisterConn(conn)
}

func handleStats(hub Hub, w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(map[string]int{"client_count": hub.GetSubscribedNumber()})
	if err != nil {
		// TODO: Refactor logging to use zap logger CORE-109
		log.Fatal("Could not encode response")
	}
}
