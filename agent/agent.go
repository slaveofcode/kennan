package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type AgentURL string

type WS struct {
	Conn   *websocket.Conn
	Dialer *websocket.Dialer
}

type Agent struct {
	url     AgentURL
	headers http.Header
	WS      *WS
}

func New(addr AgentURL, headers http.Header, handshakeTimeout time.Duration) *Agent {
	return &Agent{
		url:     addr,
		headers: headers,
		WS: &WS{
			Dialer: &websocket.Dialer{
				Proxy:            http.ProxyFromEnvironment,
				HandshakeTimeout: handshakeTimeout,
			},
		},
	}
}

func (a *Agent) Dial(ctx context.Context) error {
	conn, _, err := a.WS.Dialer.DialContext(ctx, string(a.url), a.headers)

	if err != nil {
		return err
	}

	a.WS.Conn = conn

	return nil
}

func (a Agent) Close() error {
	if a.WS.Conn != nil {
		return a.WS.Conn.Close()
	}

	return nil
}
