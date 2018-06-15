package endpoints

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/crunchyroll/cx-reactions/logging"
)

// Hub defines interface for handling multiple subscriber entities.
type Hub interface {
	RegisterConn(conn *websocket.Conn)
	GetSubscribedNumber() int
}

type handlers struct {
	hub Hub
}

// TODO: Extract configs CORE-107
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 8 * time.Second,
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

// NewRouter defines the router that will handle all the routes in the project
func NewRouter(h Hub) *mux.Router {
	router := mux.NewRouter()
	handler := &handlers{hub: h}
	router.Methods(http.MethodGet).
		Path("/ws").
		HandlerFunc(handler.websockets).Name("websocket")

	router.Methods(http.MethodGet).
		Path("/stats").
		HandlerFunc(handler.stats)
	return router
}

// TODO: Add tests CORE-108
func (h handlers) websockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Logger.Error("Error upgrading to websocket: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.hub.RegisterConn(conn)
}

func (h handlers) stats(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]int{"client_count": h.hub.GetSubscribedNumber()})
	if err != nil {
		logging.Logger.Error("Could not encode response " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
