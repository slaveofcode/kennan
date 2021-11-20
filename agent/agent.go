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

type AgentURL string
type ClientID string

type Agent struct {
	url         AgentURL
	clientId    ClientID
	headers     http.Header
	wsockConn   *websocket.Conn
	wsockDialer *websocket.Dialer
}

func New(initTag string, clientId ClientID, addr AgentURL, headers http.Header) (*Agent, error) {
	// conn, _, err := websocket.DefaultDialer.DialContext(ctx, addr, headers)
	return &Agent{
		url:      addr,
		clientId: clientId,
		headers:  headers,
		wsockDialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
	}, nil
}

func NewDefault(args ...interface{}) (*Agent, error) {
	cid, err := browser.GetClientID()

	if err != nil {
		return nil, err
	}

	url := AgentURL(conf.GetServerRand())
	clientId := ClientID(cid)
	header := http.Header{
		"Origin": []string{HEADER_ORIGIN},
	}

	// overriding defaults
	if len(args) > 0 {
		for _, arg := range args {
			switch arg := arg.(type) {
			case ClientID:
				clientId = arg
			case AgentURL:
				url = arg
			case http.Header:
				header = arg
			}
		}
	}

	return New(
		strconv.FormatInt(time.Now().Unix(), 10),
		clientId,
		url,
		header,
	)
}

func (a *Agent) Dial(ctx context.Context) error {
	conn, _, err := a.wsockDialer.DialContext(ctx, string(a.url), a.headers)

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
