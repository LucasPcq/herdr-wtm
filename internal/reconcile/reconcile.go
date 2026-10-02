// Package reconcile computes which herdr workspaces to open or close from wtm
// worktree snapshots. It performs no I/O beyond path normalization.
package reconcile

import (
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Plan lists worktree paths to open and workspace ids to close.
type Plan struct {
	Open  []string
	Close []string
}

func (p Plan) Empty() bool { return len(p.Open) == 0 && len(p.Close) == 0 }

// Diff opens worktrees that appeared between before and after (unless already
// open) and closes linked workspaces whose worktree disappeared.
func Diff(before, after []wtm.Worktree, ws []herdr.Workspace) Plan {
	beforeSet, afterSet := pathSet(before), pathSet(after)
	open := map[string]bool{}
	for _, w := range ws {
		if w.Worktree != nil {
			open[Normalize(w.Worktree.CheckoutPath)] = true
		}
	}
	var plan Plan
	for _, wt := range after {
		p := Normalize(wt.Path)
		if !beforeSet[p] && !open[p] {
			plan.Open = append(plan.Open, wt.Path)
		}
	}
	for _, w := range ws {
		if w.Worktree == nil || !w.Worktree.IsLinked {
			continue
		}
		p := Normalize(w.Worktree.CheckoutPath)
		if beforeSet[p] && !afterSet[p] {
			plan.Close = append(plan.Close, w.ID)
		}
	}
	return plan
}

// Stale closes repoRoot's linked workspaces whose checkout is neither a current
// wtm worktree nor present on disk.
func Stale(repoRoot string, current []wtm.Worktree, ws []herdr.Workspace, exists func(string) bool) Plan {
	root := Normalize(repoRoot)
	cur := pathSet(current)
	var plan Plan
	for _, w := range ws {
		if w.Worktree == nil || !w.Worktree.IsLinked || Normalize(w.Worktree.RepoRoot) != root {
			continue
		}
		if cur[Normalize(w.Worktree.CheckoutPath)] || exists(w.Worktree.CheckoutPath) {
			continue
		}
		plan.Close = append(plan.Close, w.ID)
	}
	return plan
}

func pathSet(wts []wtm.Worktree) map[string]bool {
	set := make(map[string]bool, len(wts))
	for _, wt := range wts {
		set[Normalize(wt.Path)] = true
	}
	return set
}
