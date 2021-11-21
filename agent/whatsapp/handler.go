package whatsapp

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type WAHandler struct{}

func NewWaMsghandler() *WAHandler {
	return &WAHandler{}
}

func (h *WAHandler) onNewMessage(msgType int, msgBytes []byte) {
	msg := string(msgBytes)

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
