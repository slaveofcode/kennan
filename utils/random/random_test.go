package random

import (
	"testing"
)

func TestRandom(t *testing.T) {
	t.Run("Should get a plain random number", func(t *testing.T) {
		iter := 100
		min := 1
		max := 8

		for i := 0; i < iter; i++ {
			num := GetPlainInt(min, max)

			if num < 1 || num > 8 {
				t.Errorf("Got lowest than %d or higher than %d", min, max)
			}
		}
	})
}
