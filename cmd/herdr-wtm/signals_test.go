package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

// Ctrl+C or a hangup while wtm runs must not kill herdr-wtm before it
// reconciles: the test process itself would die without the shield.
func TestShieldSignalsSurvivesInterruptAndHangup(t *testing.T) {
	stop := shieldSignals()
	defer stop()
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGHUP} {
		if err := syscall.Kill(os.Getpid(), sig); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(200 * time.Millisecond)
}
