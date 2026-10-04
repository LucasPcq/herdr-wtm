package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// herdr's [[startup]] hooks are one-shot, so `watch --detach`
// re-executes herdr-wtm as `watch` in its own session and returns at once.

func detachWatch() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, "watch")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// No stdio: the watcher logs to herdr-wtm.log in the plugin state dir.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("detach watcher: %w", err)
	}
	return cmd.Process.Release()
}

// runWatch runs the watcher unless another one already holds the lock (a
// second startup hook, after a herdr handoff).
func runWatch(d app.Deps, stateDir string) error {
	if stateDir != "" {
		f, err := os.OpenFile(filepath.Join(stateDir, domain.WatchLockFile), os.O_CREATE|os.O_RDWR, domain.FileMode)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			d.Log.Printf("watch: another watcher is running")
			return nil
		}
		_ = f.Truncate(0)
		fmt.Fprintf(f, "%d\n", os.Getpid())
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	d.Log.Printf("watch: started (pid %d)", os.Getpid())
	return d.Watch(ctx)
}
