// Package domain holds the types, constants and sentinel errors every layer shares.
package domain

// Worktree is the part of `wtm list --output json` the plugin reads.
type Worktree struct {
	Branch   string `json:"branch"`
	Path     string `json:"path"`
	IsParent bool   `json:"is_parent"`
}

// WorktreeInfo is herdr's view of the git checkout behind a workspace.
type WorktreeInfo struct {
	CheckoutPath string `json:"checkout_path"`
	RepoRoot     string `json:"repo_root"`
	IsLinked     bool   `json:"is_linked_worktree"`
}

// Workspace is one entry of `herdr workspace list`.
type Workspace struct {
	ID       string        `json:"workspace_id"`
	Label    string        `json:"label"`
	Focused  bool          `json:"focused"`
	Worktree *WorktreeInfo `json:"worktree"`
}

// HerdrContext is the part of HERDR_PLUGIN_CONTEXT_JSON the plugin reads.
type HerdrContext struct {
	WorkspaceID    string        `json:"workspace_id"`
	WorkspaceCWD   string        `json:"workspace_cwd"`
	FocusedPaneCWD string        `json:"focused_pane_cwd"`
	Worktree       *WorktreeInfo `json:"worktree"`
}

// Contracts is `wtm version --output json`: the binary's version and the
// version of each machine contract it speaks.
type Contracts struct {
	Version string `json:"version"`
	Events  int    `json:"events"`
}

// Event is one line of `wtm events --output json` (schema v1), reduced to the
// fields the plugin reads.
type Event struct {
	V             int             `json:"v"`
	Type          string          `json:"type"`
	Repo          EventRepo       `json:"repo"`
	Worktrees     []EventWorktree `json:"worktrees,omitempty"`
	Worktree      *EventWorktree  `json:"worktree,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	OK            *bool           `json:"ok,omitempty"`
	Hook          string          `json:"hook,omitempty"`
	ExitCode      *int            `json:"exit_code,omitempty"`
}

type EventRepo struct {
	Root      string `json:"root"`
	CommonDir string `json:"common_dir"`
}

type EventWorktree struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	IsMain bool   `json:"is_main"`
}

// FS is the filesystem as path rules see it, injected so rules stay pure.
type FS struct {
	Normalize func(string) string
	Exists    func(string) bool
}
