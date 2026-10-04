package wtm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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
	cmd := exec.CommandContext(ctx, c.Bin, "events", "--repo", repo, "--output", "json")
	cmd.Dir = repo
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("wtm events: %w", err)
	}
	sc := bufio.NewScanner(out)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var ev Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			continue
		}
		handle(ev)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wtm events: %w", err)
	}
	return sc.Err()
}
