package endpoints

import (
	"encoding/json"
	"net/http"

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

// NewRouter defines the router that will handle all the routes in the project
func NewRouter(h Hub, upgrader websocket.Upgrader) *mux.Router {
	router := mux.NewRouter()
	handler := &handlers{hub: h}
	router.Methods(http.MethodGet).
		Path("/ws").
		HandlerFunc(handler.websockets(upgrader)).Name("websocket")

	router.Methods(http.MethodGet).
		Path("/stats").
		HandlerFunc(handler.stats)
	return router
}

func (h handlers) websockets(upgrader websocket.Upgrader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Logger.Error("Error upgrading to websocket: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.hub.RegisterConn(conn)
	}
}

func (h handlers) stats(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]int{"client_count": h.hub.GetSubscribedNumber()})
	if err != nil {
		logging.Logger.Error("Could not encode response " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
