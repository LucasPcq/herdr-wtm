package app_test

import (
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
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
