// Package reconcile computes the popup's before/after workspace plan.
// Diff is removed in 0.2.0 with the popup's before/after reconciliation.
package reconcile

import "github.com/LucasPcq/herdr-wtm/internal/domain"

// Plan lists worktree paths to open and workspace ids to close.
type Plan struct {
	Open  []string
	Close []string
}

func (p Plan) Empty() bool { return len(p.Open) == 0 && len(p.Close) == 0 }

type DiffParams struct {
	Before     []domain.Worktree
	After      []domain.Worktree
	Workspaces []domain.Workspace
	FS         domain.FS
}

// Diff opens worktrees that appeared between Before and After (unless already
// open) and closes linked workspaces whose worktree disappeared from wtm and
// from disk.
func Diff(p DiffParams) Plan {
	beforeSet, afterSet := pathSet(p.Before, p.FS), pathSet(p.After, p.FS)
	open := map[string]bool{}
	for _, w := range p.Workspaces {
		if w.Worktree != nil {
			open[p.FS.Normalize(w.Worktree.CheckoutPath)] = true
		}
	}
	var plan Plan
	for _, wt := range p.After {
		path := p.FS.Normalize(wt.Path)
		if !beforeSet[path] && !open[path] {
			plan.Open = append(plan.Open, wt.Path)
		}
	}
	for _, w := range p.Workspaces {
		if w.Worktree == nil || !w.Worktree.IsLinked {
			continue
		}
		path := p.FS.Normalize(w.Worktree.CheckoutPath)
		if beforeSet[path] && !afterSet[path] && !p.FS.Exists(w.Worktree.CheckoutPath) {
			plan.Close = append(plan.Close, w.ID)
		}
	}
	return plan
}

func pathSet(wts []domain.Worktree, fs domain.FS) map[string]bool {
	set := make(map[string]bool, len(wts))
	for _, wt := range wts {
		set[fs.Normalize(wt.Path)] = true
	}
	return set
}
