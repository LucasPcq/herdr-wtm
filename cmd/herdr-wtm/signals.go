package main

import (
	"os"
	"os/signal"
	"syscall"
)

// shieldSignals keeps SIGINT and SIGHUP from killing herdr-wtm while wtm runs
// in the popup, so the worktrees wtm already changed still get reconciled.
// Catching (rather than ignoring) leaves the default handlers in place for the
// wtm child, which stays interruptible. The returned func restores defaults.
func shieldSignals() func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP)
	go func() {
		for range ch {
		}
	}()
	return func() { signal.Stop(ch) }
}
