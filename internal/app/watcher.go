package app

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

// Watcher applies wtm events to herdr's workspaces. One stream goroutine calls
// Handle; only notes is shared, with its flush timer.
type Watcher struct {
	d     Deps
	paths map[string][]string
	notes *notes
}

type WatcherParams struct {
	Deps  Deps
	Quiet time.Duration
}

func NewWatcher(p WatcherParams) *Watcher {
	return &Watcher{
		d:     p.Deps,
		paths: map[string][]string{},
		notes: newNotes(notesParams{Quiet: p.Quiet, Send: p.Deps.notify}),
	}
}

func (w *Watcher) Flush() { w.notes.Flush() }

func (w *Watcher) Handle(ev domain.Event) {
	switch ev.Type {
	case domain.EventSnapshot:
		w.onSnapshot(ev)
	case domain.EventCreated, domain.EventRelocated:
		w.onCreated(ev)
	case domain.EventProvisioned:
		w.onProvisioned(ev)
	case domain.EventRemoved:
		w.onRemoved(ev)
	}
}

func (w *Watcher) onSnapshot(ev domain.Event) {
	paths := make([]string, 0, len(ev.Worktrees))
	for _, wt := range ev.Worktrees {
		paths = append(paths, wt.Path)
	}
	w.paths[ev.Repo.CommonDir] = paths
	ws, shown := w.shown(ev.Repo)
	if !shown {
		return
	}
	for _, id := range rules.Stale(rules.StaleParams{RepoRoot: ev.Repo.Root, Current: paths, Workspaces: ws, FS: w.d.FS}) {
		w.close(id)
	}
}

func (w *Watcher) onCreated(ev domain.Event) {
	if ev.Worktree == nil || ev.Worktree.IsMain {
		return
	}
	w.track(ev.Repo.CommonDir, ev.Worktree.Path)
	ws, shown := w.shown(ev.Repo)
	if !shown {
		return
	}
	if _, open := rules.WorkspaceAt(rules.WorkspaceAtParams{Workspaces: ws, Path: ev.Worktree.Path, FS: w.d.FS}); open {
		return
	}
	name := filepath.Base(ev.Worktree.Path)
	// Focus follows the user's own action only: an agent's worktree never steals it.
	focus := rules.IsOwnCorrelation(ev.CorrelationID)
	if _, err := w.d.Herdr.OpenWorktree(herdr.OpenParams{Repo: ev.Repo.Root, Path: ev.Worktree.Path, Focus: focus}); err != nil {
		w.fail(fmt.Sprintf("open %s: %v", name, err))
		return
	}
	w.notes.Opened(name)
}

func (w *Watcher) onProvisioned(ev domain.Event) {
	msg, failed := rules.ProvisionFailure(ev)
	if !failed {
		return
	}
	if _, shown := w.shown(ev.Repo); !shown {
		return
	}
	w.fail(msg)
}

func (w *Watcher) onRemoved(ev domain.Event) {
	if ev.Worktree == nil {
		return
	}
	w.untrack(ev.Repo.CommonDir, ev.Worktree.Path)
	if w.d.FS.Exists(ev.Worktree.Path) {
		return
	}
	ws, err := w.d.Herdr.Workspaces()
	if err != nil {
		w.d.Log.Printf("watch: %v", err)
		return
	}
	at, open := rules.WorkspaceAt(rules.WorkspaceAtParams{Workspaces: ws, Path: ev.Worktree.Path, FS: w.d.FS})
	if !open || !at.Worktree.IsLinked {
		return
	}
	// Closing the focused workspace first would leave the user wherever herdr lands.
	if at.Focused && rules.IsOwnCorrelation(ev.CorrelationID) {
		if err := w.d.focusMain(ev.Repo.Root, ws); err != nil {
			w.fail(fmt.Sprintf("focus the main checkout: %v", err))
		}
	}
	w.close(at.ID)
}

// shown fetches the workspaces and reports whether herdr shows repo.
func (w *Watcher) shown(repo domain.EventRepo) ([]domain.Workspace, bool) {
	ws, err := w.d.Herdr.Workspaces()
	if err != nil {
		w.d.Log.Printf("watch: %v", err)
		return nil, false
	}
	cwds, err := w.d.Herdr.PaneCWDs()
	if err != nil {
		w.d.Log.Printf("watch: %v", err)
	}
	return ws, rules.RepoShown(rules.RepoShownParams{
		RepoRoot: repo.Root, WorktreePaths: w.paths[repo.CommonDir], Workspaces: ws, PaneCWDs: cwds, FS: w.d.FS,
	})
}

func (w *Watcher) close(id string) {
	if err := w.d.Herdr.Close(id); err != nil {
		w.fail(fmt.Sprintf("close %s: %v", id, err))
		return
	}
	w.notes.Closed()
}

func (w *Watcher) fail(msg string) {
	w.d.Log.Printf("watch: %s", msg)
	w.notes.Failed(msg)
}

func (w *Watcher) track(commonDir, path string) {
	if !slices.Contains(w.paths[commonDir], path) {
		w.paths[commonDir] = append(w.paths[commonDir], path)
	}
}

func (w *Watcher) untrack(commonDir, path string) {
	w.paths[commonDir] = slices.DeleteFunc(w.paths[commonDir], func(p string) bool { return p == path })
}

// focusMain brings the user to repoRoot's main checkout, opening its workspace
// when none is open.
func (d Deps) focusMain(repoRoot string, ws []domain.Workspace) error {
	if main, ok := rules.MainWorkspace(rules.MainWorkspaceParams{Workspaces: ws, RepoRoot: repoRoot, FS: d.FS}); ok {
		return d.Herdr.Focus(main.ID)
	}
	_, err := d.Herdr.OpenWorktree(herdr.OpenParams{Repo: repoRoot, Path: repoRoot, Focus: true})
	return err
}
