package app

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/gitx"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
)

// POC (LUC-233): Watch keeps herdr workspaces in sync with `wtm events`.
// herdr runs no plugin daemon, so main detaches this process at startup.

// watchTick is how often Watch checks herdr is still up and looks for repos
// it does not stream yet.
const watchTick = 5 * time.Second

// Watch streams `wtm events` for every repository herdr shows a workspace of,
// and applies each change to the workspaces. It returns when herdr stops
// answering or ctx is cancelled.
func (d Deps) Watch(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var mu sync.Mutex // one herdr reconciliation at a time across repos
	watched := map[string]bool{}
	for {
		ws, err := d.Herdr.Workspaces()
		if err != nil {
			d.Log.Printf("watch: herdr is gone, stopping: %v", err)
			return nil
		}
		for _, repo := range d.watchRepos(ws, watched) {
			watched[repo] = true
			d.Log.Printf("watch: streaming %s", repo)
			go d.streamRepo(ctx, repo, &mu)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(watchTick):
		}
	}
}

// streamRepo reads repo's event stream, restarting it with a backoff when it
// ends: a fresh stream opens on a snapshot, so nothing is lost but latency. A
// stream that fails before its first ready (not a wtm repository, a wtm
// without `events`) is not retried.
func (d Deps) streamRepo(ctx context.Context, repo string, mu *sync.Mutex) {
	backoff := time.Second
	for ctx.Err() == nil {
		ready := false
		err := d.Wtm.Events(ctx, repo, func(ev domain.Event) {
			mu.Lock()
			defer mu.Unlock()
			ready = ready || ev.Type == "ready"
			d.handleEvent(repo, ev)
		})
		if ctx.Err() != nil {
			return
		}
		if !ready {
			d.Log.Printf("watch: %s not watched: %v", repo, err)
			return
		}
		d.Log.Printf("watch: %s stream ended (%v), retrying in %s", repo, err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 30*time.Second)
	}
}

func (d Deps) handleEvent(repo string, ev domain.Event) {
	d.Log.Printf("watch: %s %s", ev.Type, eventPath(ev))
	switch ev.Type {
	case "snapshot":
		ws, err := d.Herdr.Workspaces()
		if err != nil {
			d.Log.Printf("watch: %v", err)
			return
		}
		current := make([]domain.Worktree, 0, len(ev.Worktrees))
		for _, w := range ev.Worktrees {
			current = append(current, domain.Worktree{Branch: w.Branch, Path: w.Path, IsParent: w.IsMain})
		}
		d.apply(repo, reconcile.Stale(repo, current, ws, d.Exists))
	case "worktree.created":
		if ev.Worktree == nil || ev.Worktree.IsMain {
			return
		}
		ws, err := d.Herdr.Workspaces()
		if err != nil {
			d.Log.Printf("watch: %v", err)
			return
		}
		if workspaceAt(ws, ev.Worktree.Path) != "" {
			return
		}
		// Never steal focus: the change may come from an agent in another pane.
		if _, err := d.Herdr.OpenWorktree(repo, ev.Worktree.Path, false); err != nil {
			d.Log.Printf("watch: open %s: %v", ev.Worktree.Path, err)
		}
	case "worktree.removed":
		if ev.Worktree == nil || d.Exists(ev.Worktree.Path) {
			return
		}
		ws, err := d.Herdr.Workspaces()
		if err != nil {
			d.Log.Printf("watch: %v", err)
			return
		}
		if id := workspaceAt(ws, ev.Worktree.Path); id != "" && isLinked(ws, id) {
			d.apply(repo, reconcile.Plan{Close: []string{id}})
		}
	}
}

// watchRepos returns the repositories not yet in watched: those herdr reports
// on workspaces, plus those holding a pane's cwd, since herdr only reports a
// worktree for workspaces it opened as one. Paths seen are remembered in
// watched, so a pane outside any repository costs one git call.
func (d Deps) watchRepos(ws []domain.Workspace, watched map[string]bool) []string {
	var repos []string
	add := func(repo string) {
		if !watched[repo] && !slices.Contains(repos, repo) {
			repos = append(repos, repo)
		}
	}
	for _, w := range ws {
		if w.Worktree != nil && w.Worktree.RepoRoot != "" {
			add(w.Worktree.RepoRoot)
		}
	}
	cwds, err := d.Herdr.PaneCWDs()
	if err != nil {
		d.Log.Printf("watch: %v", err)
	}
	for _, cwd := range cwds {
		if watched["cwd:"+cwd] {
			continue
		}
		watched["cwd:"+cwd] = true
		if repo, err := gitx.RepoRoot(d.Git, cwd); err == nil {
			add(repo)
		}
	}
	return repos
}

// workspaceAt returns the id of the workspace whose checkout is path, or "".
func workspaceAt(ws []domain.Workspace, path string) string {
	target := reconcile.Normalize(path)
	for _, w := range ws {
		if w.Worktree != nil && reconcile.Normalize(w.Worktree.CheckoutPath) == target {
			return w.ID
		}
	}
	return ""
}

func isLinked(ws []domain.Workspace, id string) bool {
	i := slices.IndexFunc(ws, func(w domain.Workspace) bool { return w.ID == id })
	return i >= 0 && ws[i].Worktree != nil && ws[i].Worktree.IsLinked
}

func eventPath(ev domain.Event) string {
	if ev.Worktree != nil {
		return ev.Worktree.Path
	}
	return ev.Repo.Root
}
