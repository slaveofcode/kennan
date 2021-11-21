package main

import (
	"context"
	"time"

	"github.com/slaveofcode/kennan/agent/whatsapp"
)

func main() {
	waClient, err := whatsapp.NewWhatsAppAgent()

	if err != nil {
		panic("Unable create agent:" + err.Error())
	}

	err = waClient.Connect(context.Background())

	if err != nil {
		panic("Unable connect:" + err.Error())
	}

	time.Sleep(time.Second * 60)

	defer waClient.Close()
}
