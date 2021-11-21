package testing

import (
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

type MockWSHandler struct {
	t *testing.T
}

func (h *MockWSHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// default handlers
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		h.t.Errorf("Upgrade Error: %v", err)
		return
	}

	defer ws.Close()

	// receive & respond ws message
	for {
		msgType, msg, err := ws.ReadMessage()

		if err != nil {
			break
		}

		err = ws.WriteMessage(msgType, msg)
		if err != nil {
			break
		}
	}
}
