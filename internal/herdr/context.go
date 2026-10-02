package herdr

import (
	"encoding/json"
	"fmt"
)

// Context is the part of HERDR_PLUGIN_CONTEXT_JSON the plugin uses.
type Context struct {
	WorkspaceID    string        `json:"workspace_id"`
	WorkspaceCWD   string        `json:"workspace_cwd"`
	FocusedPaneCWD string        `json:"focused_pane_cwd"`
	Worktree       *WorktreeInfo `json:"worktree"`
}

// ParseContext decodes HERDR_PLUGIN_CONTEXT_JSON; an empty string is an empty context.
func ParseContext(raw string) (Context, error) {
	var ctx Context
	if raw == "" {
		return ctx, nil
	}
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return Context{}, fmt.Errorf("parse herdr plugin context: %w", err)
	}
	return ctx, nil
}
