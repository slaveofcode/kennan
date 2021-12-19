package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	qrterminal "github.com/mdp/qrterminal/v3"
	"github.com/slaveofcode/kennan/agent/whatsapp"
	"github.com/slaveofcode/kennan/utils/content"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

func createKeys() ([32]byte, [32]byte) {
	var pubKey, secKey [32]byte
	_, err := io.ReadFull(rand.Reader, secKey[:])
	if err != nil {
		panic("Unable generate Public & Secret Key")
	}

	secKey[0] &= 248
	secKey[31] &= 127
	secKey[31] |= 64

	curve25519.ScalarBaseMult(&pubKey, &secKey)

	return pubKey, secKey
}

func main() {
	var sharedKeyBytes [32]byte

	// pubKey, baseKey, err := ed25519.GenerateKey(rand.Reader)
	// secKey := baseKey[:32]
	// if err != nil {
	// 	panic("Unable generate key:" + err.Error())
	// }

	pubKey, secKey := createKeys()

	waClient, err := whatsapp.NewWhatsAppAgent()

	if err != nil {
		panic("Unable create agent:" + err.Error())
	}

	err = waClient.Connect(context.Background())

	if err != nil {
		panic("Unable connect:" + err.Error())
	}

	defer waClient.Close()

	// listen qr
	go func() {
		for !waClient.Handler.IsAuthenticated() {
			qr := <-waClient.Handler.GetQRDataChan()
			waClient.Handler.SetOnQRScan(true)

			if err != nil {
				log.Println("Unable generate key")
				continue
			}

			// print QR
			base64PubKey := base64.StdEncoding.EncodeToString(pubKey[:])

			qrCode := qr.Ref + "," + string(base64PubKey) + "," + waClient.Auth.ClientID

			log.Println("QRCode:", qrCode)

			qrterminal.Generate(qrCode, qrterminal.L, os.Stdout)
		}
	}()

	// listen successful connection info
	go func() {
		defer func() {
			log.Println("OOppps...")
		}()
		connInfo := <-waClient.Handler.GetConnInfoChan()

		secretBytes, err := base64.StdEncoding.DecodeString(connInfo.Secret)

		if err != nil {
			log.Println("Unable decode secret connection info", err)
			return
		}

		if len(secretBytes) != 144 {
			log.Println("Invalid secret length")
			return
		}

		log.Println("Got secret key:", connInfo.Secret)

		var pubBytes [32]byte
		copy(pubBytes[:], secretBytes[:32])

		skey, err := curve25519.X25519(secKey[:], pubBytes[:])
		copy(sharedKeyBytes[:], skey)
		// curve25519.ScalarMult(&sharedKeyBytes, &secKey, &pubBytes)
		if err != nil {
			log.Println("Unable generate sharedKey", err)
			return
		}

		// empty?
		log.Println("len", len(sharedKeyBytes), "-", string(sharedKeyBytes[:]), "-", string(sharedKeyBytes[0]))

		hmacH := hmac.New(sha256.New, make([]byte, 32))
		hmacH.Write(sharedKeyBytes[:])

		r := hkdf.Expand(sha256.New, hmacH.Sum(nil), []byte(""))

		sharedKeyBytesExpanded := make([]byte, 80)
		_, err = io.ReadAtLeast(r, sharedKeyBytesExpanded, 80)
		if err != nil {
			log.Println("Unable read expanded sharedKey", err)
			return
		}

		validationKey := sharedKeyBytesExpanded[32:64]
		var waValidationKey [112]byte
		copy(waValidationKey[:32], secretBytes[:32])
		copy(waValidationKey[32:], secretBytes[64:])

		hmacHash := hmac.New(sha256.New, validationKey)
		hmacHash.Write(waValidationKey[:])

		log.Println("hmacHash.Sum(nil):", hmacHash.Sum(nil))
		log.Println("secretBytes[32:64]:", secretBytes[32:64])
		if !hmac.Equal(hmacHash.Sum(nil), secretBytes[32:64]) {
			log.Println("Invalid key")
			return
		}

		keyEncrypted := make([]byte, 96)
		copy(keyEncrypted[:16], sharedKeyBytesExpanded[64:])
		copy(keyEncrypted[16:], secretBytes[64:])

		cb, err := aes.NewCipher(sharedKeyBytesExpanded[:32])
		if err != nil {
			log.Println("Unable prepare AES decryptor", err)
			return
		}

		if len(keyEncrypted) < aes.BlockSize {
			log.Printf("ciphertext is shorter then block size: %d / %d", len(keyEncrypted), aes.BlockSize)
			return
		}

		iv := keyEncrypted[:aes.BlockSize]
		cipherText := keyEncrypted[aes.BlockSize:]

		cbc := cipher.NewCBCDecrypter(cb, iv)
		cbc.CryptBlocks(cipherText, cipherText)
		keyDecrypted, err := content.UnPad(cipherText)
		if err != nil {
			log.Println("Unable decrypt key", err)
			return
		}

		log.Println("Key decrypted:", string(keyDecrypted[:32]))
		// encKey, macKey
		waClient.SetKeys(keyDecrypted[:32], keyDecrypted[32:64])
	}()

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	// send close signal
	log.Println("Closing...")
}

// s1,["Conn",{"ref":"1@bsWS19xN0ppyp5XGZX8by6KxDQ6hbzdyBwnUOH7fRLYG5fWBEW0rhFxfywHx0acbzcQ0/5u+ucq/Jw==","wid":"6285716114426@c.us","connected":true,"isResponse":"false","serverToken":"1@A8JAMgrQrmDVjejdPc9J7KIg9gZ4Xs0e7qCf5vFUUmk2yqz9PZPDBgJFwKvAmbeJpboqXOlJo75WTw==","browserToken":"1@TGw9oIf++li43tim2eUbZJRhuCj3WXv407oJbvxgeDS0CgK0gyEq5owWmF2QeXUPMYPfPOstKadf89CWxiyDRbCXvI8I/WmYJvsrgEjJkQW/Tc+/DNGii6r/Vj3FVuxYR26TccyN5N5IQYPk65oERg==","clientToken":"Yp6bA/WSITyXoVtkO/t8tBfxkr+Vf6zZrTk+VdFFPCo=","lc":"ID","lg":"en","locales":"en-ID,id-ID","secret":"Mrl9ySLOrdcfgdELlo+sKjQuZnO3FlghMxFuLQ8tXngEOe59lX/jpLvHAgn+sLFmVKLfYy8mLBPXvroyRV922goQi0njN1npCQFIobJMpzpWjExD1tC5cMCR5jPzxlU/SovOB/jM4d6u7pLwWw8z6qw7VN+UmcTaUsoyc/gFWUsqEDLKNqdlEptZDFEh4WWx","protoVersion":[0,17],"binVersion":11,"battery":72,"plugged":false,"platform":"android","features":{"URL":true,"FLAGS":"EAEYASgBOAFAAUgBWAFgAWgBeAGYAQGgAQGwAQK4AQHIAQHYAQHgAQPoAQLwAQP4AQOAAgOIAgGgAgOoAgPAAgHQAgPYAgPgAgPwAgE="},"phone":{"wa_version":"2.21.22.26","mcc":"510","mnc":"010","os_version":"10","device_manufacturer":"HUAWEI","device_model":"HWYAL","os_build_number":"YAL-L61 10.1.0.250(C301E12R1P2)"},"pushname":"Kresna","tos":0}]
