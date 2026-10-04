// Package rules holds the plugin's decisions as pure functions: stdlib and
// domain only, the filesystem injected as a domain.FS.
package rules

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

type WorkspaceAtParams struct {
	Workspaces []domain.Workspace
	Path       string
	FS         domain.FS
}

// WorkspaceAt returns the workspace whose checkout is Path.
func WorkspaceAt(p WorkspaceAtParams) (domain.Workspace, bool) {
	target := p.FS.Normalize(p.Path)
	return find(p.Workspaces, func(w domain.Workspace) bool {
		return w.Worktree != nil && p.FS.Normalize(w.Worktree.CheckoutPath) == target
	})
}

type MainWorkspaceParams struct {
	Workspaces []domain.Workspace
	RepoRoot   string
	FS         domain.FS
}

// MainWorkspace returns the workspace of RepoRoot's main checkout.
func MainWorkspace(p MainWorkspaceParams) (domain.Workspace, bool) {
	root := p.FS.Normalize(p.RepoRoot)
	return find(p.Workspaces, func(w domain.Workspace) bool {
		return w.Worktree != nil && !w.Worktree.IsLinked && p.FS.Normalize(w.Worktree.CheckoutPath) == root
	})
}

type RepoShownParams struct {
	RepoRoot      string
	WorktreePaths []string
	Workspaces    []domain.Workspace
	PaneCWDs      []string
	FS            domain.FS
}

// RepoShown reports whether herdr shows RepoRoot: a workspace herdr knows as
// one of its worktrees, or a pane inside one of them, since herdr reports no
// worktree for a workspace it did not open as one.
func RepoShown(p RepoShownParams) bool {
	root := p.FS.Normalize(p.RepoRoot)
	if slices.ContainsFunc(p.Workspaces, func(w domain.Workspace) bool {
		return w.Worktree != nil && p.FS.Normalize(w.Worktree.RepoRoot) == root
	}) {
		return true
	}
	dirs := append([]string{p.RepoRoot}, p.WorktreePaths...)
	return slices.ContainsFunc(p.PaneCWDs, func(cwd string) bool {
		return within(withinParams{Path: cwd, Dirs: dirs, FS: p.FS})
	})
}

type withinParams struct {
	Path string
	Dirs []string
	FS   domain.FS
}

func within(p withinParams) bool {
	path := p.FS.Normalize(p.Path)
	return slices.ContainsFunc(p.Dirs, func(dir string) bool {
		dir = p.FS.Normalize(dir)
		return path == dir || strings.HasPrefix(path, dir+string(filepath.Separator))
	})
}

type BranchAtParams struct {
	Worktrees []domain.Worktree
	Origin    string
	FS        domain.FS
}

// BranchAt returns the branch of the linked worktree at Origin, or "".
func BranchAt(p BranchAtParams) string {
	if p.Origin == "" {
		return ""
	}
	target := p.FS.Normalize(p.Origin)
	i := slices.IndexFunc(p.Worktrees, func(wt domain.Worktree) bool {
		return !wt.IsParent && p.FS.Normalize(wt.Path) == target
	})
	if i < 0 {
		return ""
	}
	return p.Worktrees[i].Branch
}

func find(ws []domain.Workspace, match func(domain.Workspace) bool) (domain.Workspace, bool) {
	i := slices.IndexFunc(ws, match)
	if i < 0 {
		return domain.Workspace{}, false
	}
	return ws[i], true
}
