package main

import (
	"os"
	"os/signal"
	"syscall"
)

// shieldSignals keeps SIGINT and SIGHUP from killing herdr-wtm while wtm runs
// in the popup; the wtm child keeps the default handlers and stays
// interruptible.
func shieldSignals() func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP)
	go func() {
		for range ch {
		}
	}()
	return func() { signal.Stop(ch) }
}
