package whatsapp

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Run("Should detect Pong response", func(t *testing.T) {
		h := NewWaMsghandler()

		now := time.Now().Unix()
		ts := strconv.FormatInt(now, 10)

		isPong, lastSeen, err := h.isPong([]byte(`!` + ts))

		require.NoError(t, err, "Should no error at pong reading")

		require.True(t, isPong, "Pong should true")
		require.True(t, lastSeen.Unix() == now, "Time doesn't match")
	})
}
