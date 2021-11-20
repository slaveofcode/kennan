package agent

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
	t *testing.T
}

func (h *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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

func rootCAs(t *testing.T, s *httptest.Server) *x509.CertPool {
	certs := x509.NewCertPool()
	for _, c := range s.TLS.Certificates {
		roots, err := x509.ParseCertificates(c.Certificate[len(c.Certificate)-1])
		if err != nil {
			t.Fatalf("error parsing server's root cert: %v", err)
		}
		for _, root := range roots {
			certs.AddCert(root)
		}
	}
	return certs
}

func TestAgent(t *testing.T) {

	server := httptest.NewTLSServer(&testHandler{})
	server.URL = "ws" + strings.TrimPrefix(server.URL, "http") // wss:// or ws://

	t.Run("Create new default agent", func(t *testing.T) {
		isConnected := false

		// hijack origin handler
		originHandler := server.Config.Handler
		server.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			isConnected = true
			originHandler.ServeHTTP(rw, r)
		})

		agent, err := New("xxxx", "yyyy", server.URL, http.Header{
			"Origin": []string{HEADER_ORIGIN},
		})

		require.NoError(t, err, "Error creating new Agent")

		agent.wsockDialer.TLSClientConfig = &tls.Config{
			RootCAs: rootCAs(t, server),
		}

		err = agent.Dial(context.Background())

		require.NoError(t, err, "Error dial websocket")
		require.True(t, isConnected)

		defer agent.Close()

		// send some message
		wsMsgs := []string{
			"Hi",
			"How r u?",
			"Ok, gotta go",
		}

		for _, msg := range wsMsgs {
			err := agent.wsockConn.WriteMessage(
				websocket.TextMessage,
				[]byte(msg),
			)

			require.NoError(t, err)

			_, rMsg, err := agent.wsockConn.ReadMessage()

			require.NoError(t, err)
			require.Equal(t, msg, string(rMsg))
		}
	})

	t.Run("Create new agent", func(t *testing.T) {

	})

	t.Cleanup(func() {
		server.Close()
	})
}
