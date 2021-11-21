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

type WAHandler struct {
	isAuth bool
	isOnQR bool
	QRData chan QRData
}

func NewWaMsghandler() *WAHandler {
	return &WAHandler{
		QRData: make(chan QRData),
	}
}

func (h *WAHandler) IsAuthenticated() bool {
	return h.isAuth
}

func (h *WAHandler) GetQRDataChan() chan QRData {
	return h.QRData
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

	log.Println("msg:", msg)
	// processing normal messages...
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
