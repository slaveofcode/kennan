package conf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomAddress(t *testing.T) {
	attempt := 25
	minDiff := 5

	lastSvrAddr := GetServerRand()
	totalDiffs := 0
	for i := 0; i < attempt; i++ {
		addr := GetServerRand()
		if lastSvrAddr != addr {
			totalDiffs++
			lastSvrAddr = addr
		}
	}

	require.GreaterOrEqual(t, totalDiffs, minDiff)
}
