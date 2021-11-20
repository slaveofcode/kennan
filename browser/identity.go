package browser

import (
	"crypto/rand"
	"encoding/base64"
)

// GetClientID will return 16 random bytes string encoded with base64
func GetClientID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}
