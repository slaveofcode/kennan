package conf

import (
	"fmt"

	"github.com/slaveofcode/kennan/utils/random"
)

// GetServerRand will return random websocket server address from 1-8
func GetServerRand() string {
	return fmt.Sprintf(
		"wss://w%d.web.whatsapp.com/ws",
		random.GetPlainInt(1, 8),
	)
}
