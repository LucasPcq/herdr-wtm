package app

import (
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

// Sync closes linked workspaces whose worktree is gone, for the invoking
// workspace's repository or, with all, for every repository herdr shows.
func (d Deps) Sync(all bool, ctx domain.HerdrContext) error {
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

	d.apply("", d.closeStale(repos, ws))
	return nil
}

// syncRepo is Sync for one repository, run from the menu: it always reports
// back so a manual sync never ends in silence.
func (d Deps) syncRepo(repo string) error {
	ws, err := d.Herdr.Workspaces()
	if err != nil {
		return d.fail(err)
	}
	plan := d.closeStale([]string{repo}, ws)
	if plan.Empty() {
		if err := d.Herdr.Notify("wtm", "workspaces already in sync"); err != nil {
			d.Log.Printf("notify: %v", err)
		}
		return nil
	}
	d.apply(repo, plan)
	return nil
}

// closeStale plans closing the stale workspaces of repos, skipping (and
// logging) a repository whose wtm list fails.
func (d Deps) closeStale(repos []string, ws []domain.Workspace) reconcile.Plan {
	var plan reconcile.Plan
	for _, repo := range repos {
		current, err := d.Wtm.List(repo)
		if err != nil {
			d.Log.Printf("sync: skipping %s: %v", repo, err)
			continue
		}
		plan.Close = append(plan.Close, rules.Stale(rules.StaleParams{RepoRoot: repo, Current: worktreePaths(current), Workspaces: ws, FS: d.FS})...)
	}
	return plan
}

func worktreePaths(wts []domain.Worktree) []string {
	paths := make([]string, len(wts))
	for i, wt := range wts {
		paths[i] = wt.Path
	}
	return paths
}
