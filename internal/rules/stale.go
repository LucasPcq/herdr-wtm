package rules

import (
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

type StaleParams struct {
	RepoRoot   string
	Current    []string
	Workspaces []domain.Workspace
	FS         domain.FS
}

// Stale returns the ids of RepoRoot's linked workspaces whose checkout is
// neither a current worktree nor on disk: a folder still there may hold work.
func Stale(p StaleParams) []string {
	root := p.FS.Normalize(p.RepoRoot)
	current := make([]string, len(p.Current))
	for i, path := range p.Current {
		current[i] = p.FS.Normalize(path)
	}
	var ids []string
	for _, w := range p.Workspaces {
		if w.Worktree == nil || !w.Worktree.IsLinked || p.FS.Normalize(w.Worktree.RepoRoot) != root {
			continue
		}
		if slices.Contains(current, p.FS.Normalize(w.Worktree.CheckoutPath)) || p.FS.Exists(w.Worktree.CheckoutPath) {
			continue
		}
		ids = append(ids, w.ID)
	}
	return ids
}
