package whatsapp

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slaveofcode/kennan/agent"
	"github.com/slaveofcode/kennan/browser"
	"github.com/slaveofcode/kennan/utils/whatsapp"
)

const WhatsAppSocketURL = "wss://web.whatsapp.com/ws"
const WSHandshakeTimeout = 15 * time.Second
const WebVersion = "[2, 2142, 12]"
const WebAgent = `["Kennan", "Chrome", "89.0.4389"]`

type ClientID string

type Handler interface {
	isPong([]byte) (bool, *time.Time, error)
	onNewMessage(int, []byte)
	onReceiveKeys([]byte, []byte)
	GetQRDataChan() chan QRData
	IsAuthenticated() bool
	SetOnQRScan(bool)
	IsOnQRScan() bool
	GetConnInfoChan() chan ConnInfo
}

type Auth struct {
	ClientID string
}

type Config struct {
	KeepAliveInterval time.Duration
	DoKeepAlive       bool
	LastKeepAliveResp time.Time
	WebAgent          string
	WebVersion        string
}

type WAInfo struct {
	EncKey   []byte
	MacKey   []byte
	LastPing *time.Time
}

type WhatsAppAgent struct {
	*agent.Agent
	Auth    *Auth
	Config  *Config
	Handler Handler
	WAInfo  *WAInfo
}

func NewWhatsAppAgent(args ...interface{}) (*WhatsAppAgent, error) {
	cid, err := browser.GetClientID()

	if err != nil {
		return nil, err
	}

	url := agent.AgentURL(WhatsAppSocketURL)
	header := http.Header{
		"Origin":          []string{"https://web.whatsapp.com"},
		"Accept-Encoding": []string{"gzip, deflate, br"},
		"Accept-Language": []string{"en-US,en;q=0.9"},
		"Cache-Control":   []string{"no-cache"},
		"Host":            []string{"web.whatsapp.com"},
		"Pragma":          []string{"no-cache"},
		"Sec-WebSocket-Extensions": []string{
			"permessage-deflate; client_max_window_bits",
		},
	}

	auth := &Auth{
		ClientID: string(ClientID(cid)),
	}
	config := &Config{
		DoKeepAlive:       true,
		KeepAliveInterval: time.Second * 20,
		WebAgent:          WebAgent,
		WebVersion:        WebVersion,
	}

	var handler Handler = NewWaMsghandler()

	// overriding defaults
	if len(args) > 0 {
		for _, arg := range args {
			switch arg := arg.(type) {
			case agent.AgentURL:
				url = arg
			case http.Header:
				header = arg
			case Handler:
				handler = arg
			case *Config:
				config = arg
			case *Auth:
				auth = arg
			}
		}
	}

	return &WhatsAppAgent{
		Agent: agent.New(
			url,
			header,
			WSHandshakeTimeout,
		),
		Handler: handler,
		Auth:    auth,
		Config:  config,
		WAInfo:  &WAInfo{},
	}, nil
}

func (wa *WhatsAppAgent) Connect(ctx context.Context) error {
	err := wa.Agent.Dial(ctx)

	if err != nil {
		log.Println("Error in dial")
		return err
	}

	go wa.startHandleMessages()
	go wa.sendKeepAlive()

	return err
}

func (wa *WhatsAppAgent) sendKeepAlive() {
	for wa.Config.DoKeepAlive {
		wa.Agent.WS.Conn.WriteMessage(websocket.TextMessage, []byte("?,,"))
		time.Sleep(wa.Config.KeepAliveInterval)
	}
}

func (wa *WhatsAppAgent) sendInit() error {
	tag := whatsapp.GenerateLoginTag()
	data := []byte(tag + `,["admin","init",` + wa.Config.WebVersion + `,` + wa.Config.WebAgent + `,"` + wa.Auth.ClientID + `",true]`)
	err := wa.WS.Conn.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		log.Println("Unable send auth:", err)
		return err
	}
	return nil
}

func (wa *WhatsAppAgent) startHandleMessages() {
	ws := wa.Agent.WS.Conn
	for {
		msgType, msgBytes, err := ws.ReadMessage()

		if err != nil {
			log.Println("Error read message:", err)
		}

		wa.Handler.onNewMessage(msgType, msgBytes)

		isPong, tStamp, err := wa.Handler.isPong(msgBytes)

		if err != nil {
			log.Println("Unable check pong response:", err)
		}

		if isPong {
			wa.WAInfo.LastPing = tStamp
			wa.sendInit()
		}
	}
}

func (wa *WhatsAppAgent) SetKeys(encKey []byte, macKey []byte) {
	wa.WAInfo.EncKey = encKey
	wa.WAInfo.MacKey = macKey

	wa.Handler.onReceiveKeys(encKey, macKey)
}
