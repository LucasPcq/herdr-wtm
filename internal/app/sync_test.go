package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

var mainCtx = domain.HerdrContext{Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}

func syncHandler(ws []domain.Workspace, snapshot ...domain.Event) func(execx.Call) ([]byte, error) {
	return func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm events --output json --repo /nx/app" {
			return eventLines(append(snapshot, domain.Event{V: 1, Type: domain.EventReady})...), nil
		}
		return herdrState(ws)(c)
	}
}

func TestSyncClosesWhatTheSnapshotNoLongerHolds(t *testing.T) {
	d, f, _ := newDeps(syncHandler([]domain.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/gone"), linkedWS("w3", "/nx/app.wt/kept")}, snapshotEv("/nx/app.wt/kept")))
	started := false
	d.StartWatcher = func() error { started = true; return nil }
	if err := d.Sync(mainCtx); err != nil {
		t.Fatal(err)
	}
	if !started {
		t.Fatal("sync must make sure the watcher runs")
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w3")
	assertHas(t, f, "herdr notification show wtm --body closed 1 workspace(s)")
}

func TestSyncSaysWhenAlreadyInSync(t *testing.T) {
	d, f, _ := newDeps(syncHandler([]domain.Workspace{primaryWS()}, snapshotEv()))
	if err := d.Sync(mainCtx); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr notification show wtm --body workspaces already in sync")
}

func TestSyncFailsWithoutSnapshot(t *testing.T) {
	d, _, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm events --output json --repo /nx/app" {
			return nil, execx.ExitError{Code: 12}
		}
		return nil, nil
	})
	if err := d.Sync(mainCtx); err == nil {
		t.Fatal("want error")
	}
}

func TestSyncGivesUpWhenNoSnapshotComes(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "wtm")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	d, _, _ := newDeps(nil)
	d.Wtm = wtm.Client{Runner: execx.OS{}, Bin: script}
	d.SnapshotTimeout = 200 * time.Millisecond
	start := time.Now()
	err := d.Sync(domain.HerdrContext{Worktree: &domain.WorktreeInfo{CheckoutPath: dir, RepoRoot: dir}})
	if err == nil || !strings.Contains(err.Error(), "no snapshot") {
		t.Fatalf("err %v", err)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("sync waited %s", time.Since(start))
	}
}
