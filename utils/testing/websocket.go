package testing

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func WSTestSendReceive(t *testing.T, ws *websocket.Conn) {
	wsMsgs := []string{
		"Hi",
		"How r u?",
		"Ok, gotta go",
	}

	for _, msg := range wsMsgs {
		err := ws.WriteMessage(
			websocket.TextMessage,
			[]byte(msg),
		)

		require.NoError(t, err)

		_, rMsg, err := ws.ReadMessage()

		require.NoError(t, err)
		require.Equal(t, msg, string(rMsg))
	}
}
