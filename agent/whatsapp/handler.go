package whatsapp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/slaveofcode/kennan/utils/content"
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

	h.isAuth = true
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

	parts := strings.SplitN(msg, ",", 2)
	if len(parts) < 2 {
		return
	}

	msgTag := parts[0]
	msgContent := parts[1]
	isJSONContent := content.IsJSON([]byte(msgContent))

	if msgTag == "s1" {
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

	// check sx
	rex := regexp.MustCompile("^s(.+)*$")
	log.Println("flag:", msgTag)
	if rex.MatchString(msgTag) {
		h.HandleInfo(msgContent)
		return
	}

	if h.IsAuthenticated() {
		if isJSONContent {
			log.Println("===================== MSG Start [JSON] ==========================")
			log.Println("MSG:", msgContent)
			log.Println("===================== MSG End [JSON] ==========================")
			return
		}

		decryptedMsg, err := h.DecryptMessage(msgContent)

		if err != nil {
			log.Println("===================== ERR Begin ==========================")
			log.Println("Error decrypt:", err)
			log.Println("===================== ERR End ==========================")
			return
		}

		log.Println("===================== MSG Start ==========================")
		log.Println("MSG:", decryptedMsg)
		log.Println("===================== MSG End ==========================")
		return
	}
}

func (h *WAHandler) HandleInfo(infoData string) {
	var infoParts []interface{}
	err := json.Unmarshal([]byte(infoData), &infoParts)

	if err != nil {
		log.Println("===================== Info Err Begin ==========================")
		log.Println("Err:", err)
		log.Println("===================== Info Err End  ==========================")
		return
	}

	if len(infoParts) < 2 {
		log.Println("===================== Info Err Begin ==========================")
		log.Printf("Invalid info, got %d length of data\n", len(infoParts))
		log.Println("===================== Info Err End  ==========================")
		return
	}

	infoType, ok := infoParts[0].(string)
	if !ok {
		log.Println("===================== Info Err Begin ==========================")
		log.Println("Non string info type", infoParts[0])
		log.Println("===================== Info Err End  ==========================")
		return
	}

	jsonInfo, err := json.Marshal(infoParts[1])
	if err != nil {
		log.Println("===================== Info Err Begin ==========================")
		log.Println("Non JSON Info:", err)
		log.Println("===================== Info Err End  ==========================")
		return
	}

	switch infoType {
	case "Conn":
		log.Println("Connection:", jsonInfo)
		return
	case "Props":
		log.Println("Properties:", jsonInfo)
		return
	case "Presence":
		log.Println("Presence:", jsonInfo)
		return
	}

	log.Println("===================== Info Begin ==========================")
	log.Println("MSG:", infoData)
	log.Println("===================== Info End  ==========================")
}

func (h *WAHandler) DecryptMessage(msg string) (string, error) {
	// validate message
	msgBytes := []byte(msg)
	hmacHash := hmac.New(sha256.New, h.MacKey)
	validationKey := msgBytes[32:]
	hmacHash.Write(validationKey)
	if !hmac.Equal(hmacHash.Sum(nil), msgBytes[:32]) {
		log.Println("Invalid key")
		return "", fmt.Errorf("invalid key validation")
	}

	msgContentEncrypted := msgBytes[:32]

	// decrypt message
	cb, err := aes.NewCipher(h.EncKey)
	if err != nil {
		return "", fmt.Errorf("unable prepare aes decryptor: %w", err)
	}

	if len(msgContentEncrypted) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext is shorter then block size: %d / %d", len(msgContentEncrypted), aes.BlockSize)
	}

	iv := msgContentEncrypted[:aes.BlockSize]
	cipherText := msgContentEncrypted[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(cb, iv)
	cbc.CryptBlocks(cipherText, cipherText)

	return string(cipherText), nil
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
