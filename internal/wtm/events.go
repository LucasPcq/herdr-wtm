package wtm

import (
	"context"
	"encoding/json"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// Event is one line of `wtm events --output json` (schema v1), reduced to the
// fields the plugin reads. Unknown types and fields are ignored, as the
// contract asks.
type Event struct {
	V         int             `json:"v"`
	Type      string          `json:"type"`
	Repo      EventRepo       `json:"repo"`
	Worktrees []EventWorktree `json:"worktrees"`
	Worktree  *EventWorktree  `json:"worktree"`
	FromPath  string          `json:"from_path"`
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

// Events runs `wtm events --output json` for repo and calls handle for each
// line until the stream ends or ctx is cancelled. It returns the process's
// exit error; a clean end of stream is nil.
func (c Client) Events(ctx context.Context, repo string, handle func(Event)) error {
	cmd := execx.Cmd{Dir: repo, Name: c.Bin, Args: []string{"events", "--repo", repo, "--output", "json"}}
	return c.Runner.Stream(ctx, cmd, func(line []byte) {
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return
		}
		handle(ev)
	})
}
