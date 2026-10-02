package app

import (
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
)

// Sync closes linked workspaces whose worktree is gone, for the invoking
// workspace's repository or, with all, for every repository herdr shows.
func (d Deps) Sync(all bool, ctx herdr.Context) error {
	ws, err := d.Herdr.Workspaces()
	if err != nil {
		return err
	}
	var repos []string
	if all {
		for _, w := range ws {
			if w.Worktree != nil && w.Worktree.RepoRoot != "" && !slices.Contains(repos, w.Worktree.RepoRoot) {
				repos = append(repos, w.Worktree.RepoRoot)
			}
		}
	} else {
		repo, err := d.resolveRepo(ctx)
		if err != nil {
			return err
		}
		repos = []string{repo}
	}

	var plan reconcile.Plan
	for _, repo := range repos {
		current, err := d.Wtm.List(repo)
		if err != nil {
			d.Log.Printf("sync: skipping %s: %v", repo, err)
			continue
		}
		plan.Close = append(plan.Close, reconcile.Stale(repo, current, ws, d.Exists).Close...)
	}
	d.apply("", plan)
	return nil
}
