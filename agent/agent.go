package agent

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slaveofcode/kennan/browser"
	"github.com/slaveofcode/kennan/conf"
)

const HEADER_ORIGIN = "https://web.whatsapp.com"

type Agent struct {
	url         string
	headers     http.Header
	wsockConn   *websocket.Conn
	wsockDialer *websocket.Dialer
}

func New(initTag, clientId, addr string, headers http.Header) (*Agent, error) {
	// conn, _, err := websocket.DefaultDialer.DialContext(ctx, addr, headers)
	return &Agent{
		url:     addr,
		headers: headers,
		wsockDialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
	}, nil
}

func NewDefault(ctx context.Context) (*Agent, error) {
	cid, err := browser.GetClientID()

	if err != nil {
		return nil, err
	}

	return New(
		strconv.FormatInt(time.Now().Unix(), 10),
		cid,
		conf.GetServerRand(),
		http.Header{
			"Origin": []string{HEADER_ORIGIN},
		},
	)
}

func (a *Agent) Dial(ctx context.Context) error {
	conn, _, err := a.wsockDialer.DialContext(ctx, a.url, a.headers)

	if err != nil {
		return err
	}

	a.wsockConn = conn

	return nil
}

func (a Agent) Close() error {
	if a.wsockConn != nil {
		return a.wsockConn.Close()
	}

	return nil
}
