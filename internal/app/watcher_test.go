package app_test

import (
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func newWatcher(h func(execx.Call) ([]byte, error)) (*app.Watcher, *execx.Fake, app.Deps) {
	d, f, _ := newDeps(h)
	return app.NewWatcher(app.WatcherParams{Deps: d}), f, d
}

func TestWatcherOpensCreatedWorktreeWithoutFocus(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS()}))
	w.Handle(snapshotEv())
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	w.Flush()
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/a --no-focus")
	assertHas(t, f, "herdr notification show wtm --body opened a")
}

func TestWatcherFocusesOwnCreation(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS()}))
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", "herdr-wtm:1"))
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/a --focus")
}

func TestWatcherIgnoresRepoNotShown(t *testing.T) {
	other := domain.Workspace{ID: "w8", Worktree: &domain.WorktreeInfo{CheckoutPath: "/other", RepoRoot: "/other"}}
	w, f, _ := newWatcher(herdrState([]domain.Workspace{other}, "/other/src"))
	w.Handle(snapshotEv())
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	w.Flush()
	assertNoPrefix(t, f, "herdr worktree open")
	assertNoPrefix(t, f, "herdr notification")
}

func TestWatcherSeesRepoThroughPaneCWD(t *testing.T) {
	plain := domain.Workspace{ID: "w5"}
	w, f, _ := newWatcher(herdrState([]domain.Workspace{plain}, "/nx/app/packages/web"))
	w.Handle(snapshotEv())
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/a --no-focus")
}

func TestWatcherSkipsAlreadyOpen(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a")}))
	w.Handle(snapshotEv("/nx/app.wt/a"))
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	w.Flush()
	assertNoPrefix(t, f, "herdr worktree open")
	assertNoPrefix(t, f, "herdr notification")
}

func TestWatcherIgnoresMainCheckoutCreation(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS()}))
	ev := wtEv(domain.EventCreated, repo, "")
	ev.Worktree.IsMain = true
	w.Handle(ev)
	assertNoPrefix(t, f, "herdr worktree open")
}

func TestWatcherClosesRemovedWorktreeKeepingFocus(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS(), focused(linkedWS("w2", "/nx/app.wt/a"))}))
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", ""))
	w.Flush()
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace focus")
	assertHas(t, f, "herdr notification show wtm --body closed 1 workspace(s)")
}

func TestWatcherFocusesMainBeforeClosingOwnFocusedRemoval(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS(), focused(linkedWS("w2", "/nx/app.wt/a"))}))
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", "herdr-wtm:1"))
	if indexOf(t, f, "herdr workspace focus w1") > indexOf(t, f, "herdr workspace close w2") {
		t.Fatalf("focus must come before close: %v", f.Lines())
	}
}

func TestWatcherOpensMainWhenNoneOpen(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{focused(linkedWS("w2", "/nx/app.wt/a"))}))
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", "herdr-wtm:1"))
	if indexOf(t, f, "herdr worktree open --cwd /nx/app --path /nx/app --focus") > indexOf(t, f, "herdr workspace close w2") {
		t.Fatalf("main must open before close: %v", f.Lines())
	}
}

func TestWatcherOwnRemovalOfUnfocusedWorkspaceKeepsFocus(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{focused(primaryWS()), linkedWS("w2", "/nx/app.wt/a")}))
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", "herdr-wtm:1"))
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace focus")
}

func TestWatcherKeepsRemovedWorktreeStillOnDisk(t *testing.T) {
	d, f, _ := newDeps(herdrState([]domain.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a")}))
	d.FS.Exists = func(p string) bool { return p == "/nx/app.wt/a" }
	w := app.NewWatcher(app.WatcherParams{Deps: d})
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", ""))
	assertNoPrefix(t, f, "herdr workspace close")
}

func TestWatcherSnapshotClosesStale(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/gone"), linkedWS("w3", "/nx/app.wt/kept")}))
	w.Handle(snapshotEv("/nx/app.wt/kept"))
	w.Flush()
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w3")
	assertNoPrefix(t, f, "herdr workspace close w1")
}

func TestWatcherNotifiesProvisionFailure(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS()}))
	no, code := false, 1
	ev := wtEv(domain.EventProvisioned, "/nx/app.wt/a", "")
	ev.OK, ev.Hook, ev.ExitCode = &no, "pnpm install", &code
	w.Handle(ev)
	w.Flush()
	assertHas(t, f, "herdr notification show wtm --body failed: a: on_create failed — pnpm install (exit 1)")
}

func TestWatcherBatchesNotifications(t *testing.T) {
	w, f, _ := newWatcher(herdrState([]domain.Workspace{primaryWS()}))
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/b", ""))
	w.Flush()
	assertHas(t, f, "herdr notification show wtm --body opened a, b")
}

func TestWatcherMatchesAliasedPaths(t *testing.T) {
	d, f, _ := newDeps(herdrState([]domain.Workspace{primaryWS(), linkedWS("w2", "/private/nx/app.wt/a")}))
	d.FS.Normalize = func(p string) string {
		if p == "/nx/app.wt/a" {
			return "/private/nx/app.wt/a"
		}
		return p
	}
	w := app.NewWatcher(app.WatcherParams{Deps: d})
	w.Handle(wtEv(domain.EventCreated, "/nx/app.wt/a", ""))
	w.Handle(wtEv(domain.EventRemoved, "/nx/app.wt/a", ""))
	assertNoPrefix(t, f, "herdr worktree open")
	assertHas(t, f, "herdr workspace close w2")
}
