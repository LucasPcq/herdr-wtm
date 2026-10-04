package app

import (
	"errors"
	"testing"
)

func TestLivenessNeedsConsecutiveFailures(t *testing.T) {
	var l liveness
	boom := errors.New("timeout")
	steps := []struct {
		err  error
		gone bool
	}{{boom, false}, {boom, false}, {nil, false}, {boom, false}, {boom, false}, {boom, true}}
	for i, s := range steps {
		if got := l.gone(s.err); got != s.gone {
			t.Fatalf("step %d: gone %v, want %v", i, got, s.gone)
		}
	}
}
