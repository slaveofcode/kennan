package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdp/qrterminal/v3"
	"github.com/slaveofcode/kennan/agent/whatsapp"
)

func main() {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Unable generate key:" + err.Error())
	}

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
			log.Println("QR:", qr.Ref, qr.TTL)

			if err != nil {
				log.Println("Unable generate key")
				continue
			}

			// print QR
			base64PubKey := base64.StdEncoding.EncodeToString(pubKey)

			qrCode := qr.Ref + "," + string(base64PubKey) + "," + waClient.Auth.ClientID

			log.Println("QRCode:", qrCode)

			qrterminal.Generate(qrCode, qrterminal.L, os.Stdout)
		}
	}()

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	// send close signal
	log.Println("Closing...")
}
