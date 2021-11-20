package browser

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClientID(t *testing.T) {
	t.Run("Should get Client ID", func(t *testing.T) {
		clientId, err := GetClientID()

		require.NoError(t, err)

		_, err = base64.StdEncoding.DecodeString(clientId)

		require.NoError(t, err)
	})
}
