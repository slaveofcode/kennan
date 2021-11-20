package random

import (
	"math/rand"
	"time"
)

// GetPlainInt will return plain random integer (unsafe)
func GetPlainInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}
