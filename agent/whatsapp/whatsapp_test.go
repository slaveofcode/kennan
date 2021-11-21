package whatsapp

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	agt "github.com/slaveofcode/kennan/agent"
	ktesting "github.com/slaveofcode/kennan/utils/testing"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	testServer := httptest.NewTLSServer(&ktesting.MockWSHandler{})
	testServer.URL = "ws" + strings.TrimPrefix(testServer.URL, "http") // wss:// or ws://

	t.Run("Create new default whatsapp agent", func(t *testing.T) {
		isConnected := false

		// hijack origin handler
		originHandler := testServer.Config.Handler
		testServer.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			isConnected = true
			originHandler.ServeHTTP(rw, r)
		})

		agent, err := NewWhatsAppAgent(agt.AgentURL(testServer.URL))

		require.NoError(t, err, "Error creating new Agent")

		agent.WS.Dialer.TLSClientConfig = &tls.Config{
			RootCAs: ktesting.AssignRootCAs(t, testServer),
		}

		err = agent.Dial(context.Background())

		require.NoError(t, err, "Error dial websocket")
		require.True(t, isConnected)

		defer agent.Close()

		// ws test message
		ktesting.WSTestSendReceive(t, agent.WS.Conn)
	})

	t.Cleanup(func() {
		testServer.Close()
	})
}
