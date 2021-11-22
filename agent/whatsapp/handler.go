package whatsapp

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"
)

type QRData struct {
	Ref string
	TTL time.Time
}

type ConnInfo struct {
	Secret string
}

type WAHandler struct {
	isAuth bool
	isOnQR bool

	QRData   chan QRData
	ConnInfo chan ConnInfo

	EncKey []byte
	MacKey []byte
}

func NewWaMsghandler() *WAHandler {
	return &WAHandler{
		QRData:   make(chan QRData),
		ConnInfo: make(chan ConnInfo),
	}
}

func (h *WAHandler) IsAuthenticated() bool {
	return h.isAuth
}

func (h *WAHandler) GetQRDataChan() chan QRData {
	return h.QRData
}

func (h *WAHandler) onReceiveKeys(encKey []byte, macKey []byte) {
	h.EncKey = encKey
	h.MacKey = macKey
}

func (h *WAHandler) onNewMessage(msgType int, msgBytes []byte) {
	msg := string(msgBytes)

	if !h.IsAuthenticated() {
		isPong, _, _ := h.isPong(msgBytes)

		if !isPong {
			go func() {
				parts := strings.SplitN(msg, ",", 2)
				if len(parts) > 1 {
					var qrData map[string]interface{}
					json.Unmarshal([]byte(parts[1]), &qrData)

					if qrData["time"] == nil || qrData["ttl"] == nil {
						// skip non QR data
						return
					}

					serverUnixTime := int64(qrData["time"].(float64))
					treshold := int64(qrData["ttl"].(float64))

					h.QRData <- QRData{
						Ref: qrData["ref"].(string),
						TTL: time.UnixMilli(serverUnixTime + treshold),
					}
				}
			}()
		}
	}

	// log.Println("===============================================")
	// log.Println("MSG:", msg)
	// log.Println("===============================================")

	parts := strings.SplitN(msg, ",", 2)
	if len(parts) < 2 {
		return
	}

	flag := parts[0]

	if flag == "s1" {
		go func() {
			var connParts []interface{}
			err := json.Unmarshal([]byte(parts[1]), &connParts)

			if err != nil {
				log.Println("Unable destruct connection info", err)
				return
			}

			if len(connParts) < 2 {
				log.Println("Invalid connection parts", connParts)
				return
			}

			connJson, ok := connParts[1].(map[string]interface{})

			if !ok {
				log.Println("Unable parse json connection info", err)
				return
			}

			h.ConnInfo <- ConnInfo{
				Secret: connJson["secret"].(string),
			}
		}()
	}
}

func (h *WAHandler) SetOnQRScan(stat bool) {
	h.isOnQR = stat
}

func (h *WAHandler) IsOnQRScan() bool {
	return h.isOnQR
}

func (h *WAHandler) GetConnInfoChan() chan ConnInfo {
	return h.ConnInfo
}

func (h *WAHandler) isPong(msgBytes []byte) (bool, *time.Time, error) {
	msg := string(msgBytes)
	if strings.HasPrefix(msg, "!") && len(msg) > 1 {
		unix, err := strconv.ParseInt(msg[1:], 10, 64)

		if err != nil {
			return true, nil, err
		}

		ts := time.Unix(unix, 0)
		return true, &ts, err
	}

	return false, nil, nil
}
