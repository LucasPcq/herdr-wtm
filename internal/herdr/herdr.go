// Package herdr drives the herdr CLI.
package herdr

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

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
	Worktree *WorktreeInfo `json:"worktree"`
}

// PopupParams describes a plugin popup to open.
type PopupParams struct {
	Plugin     string
	Entrypoint string
	Width      string
	Height     string
	Env        map[string]string
}

// Client runs herdr CLI commands.
type Client struct {
	Runner execx.Runner
	Bin    string
}

func (c Client) Workspaces() ([]Workspace, error) {
	out, err := c.output("workspace", "list")
	if err != nil {
		return nil, fmt.Errorf("herdr workspace list: %w", err)
	}
	var resp struct {
		Result struct {
			Workspaces []Workspace `json:"workspaces"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parse herdr workspace list: %w", err)
	}
	return resp.Result.Workspaces, nil
}

// PaneCWDs returns the working directory of every pane.
func (c Client) PaneCWDs() ([]string, error) {
	out, err := c.output("pane", "list")
	if err != nil {
		return nil, fmt.Errorf("herdr pane list: %w", err)
	}
	var resp struct {
		Result struct {
			Panes []struct {
				CWD string `json:"cwd"`
			} `json:"panes"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parse herdr pane list: %w", err)
	}
	cwds := make([]string, 0, len(resp.Result.Panes))
	for _, p := range resp.Result.Panes {
		if p.CWD != "" {
			cwds = append(cwds, p.CWD)
		}
	}
	return cwds, nil
}

// OpenWorktree opens the worktree at path as a workspace of repo and returns its id.
func (c Client) OpenWorktree(repo, path string, focus bool) (string, error) {
	focusFlag := "--no-focus"
	if focus {
		focusFlag = "--focus"
	}
	out, err := c.output("worktree", "open", "--cwd", repo, "--path", path, focusFlag)
	if err != nil {
		return "", fmt.Errorf("herdr worktree open: %w", err)
	}
	var resp struct {
		Result struct {
			Workspace Workspace `json:"workspace"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parse herdr worktree open: %w", err)
	}
	return resp.Result.Workspace.ID, nil
}

func (c Client) Focus(id string) error {
	return c.run("workspace", "focus", id)
}

func (c Client) Close(id string) error {
	return c.run("workspace", "close", id)
}

func (c Client) Notify(title, body string) error {
	return c.run("notification", "show", title, "--body", body)
}

func (c Client) OpenPopup(p PopupParams) error {
	args := []string{"plugin", "pane", "open", "--plugin", p.Plugin, "--entrypoint", p.Entrypoint, "--placement", "popup"}
	if p.Width != "" {
		args = append(args, "--width", p.Width)
	}
	if p.Height != "" {
		args = append(args, "--height", p.Height)
	}
	keys := make([]string, 0, len(p.Env))
	for k, v := range p.Env {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		args = append(args, "--env", k+"="+p.Env[k])
	}
	return c.run(args...)
}

func (c Client) run(args ...string) error {
	if _, err := c.output(args...); err != nil {
		return fmt.Errorf("herdr %s: %w", args[0], err)
	}
	return nil
}

// DefaultConfig returns `herdr --default-config`, the documented defaults.
func (c Client) DefaultConfig() (string, error) {
	out, err := c.output("--default-config")
	if err != nil {
		return "", fmt.Errorf("herdr --default-config: %w", err)
	}
	return string(out), nil
}

// ReloadConfig asks the server to reload config.toml and fails unless herdr
// applied it without diagnostics.
func (c Client) ReloadConfig() error {
	out, err := c.output("server", "reload-config")
	if err != nil {
		return fmt.Errorf("herdr server reload-config: %w", err)
	}
	var resp struct {
		Result struct {
			Status      string            `json:"status"`
			Diagnostics []json.RawMessage `json:"diagnostics"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("parse herdr server reload-config: %w", err)
	}
	if resp.Result.Status != "applied" || len(resp.Result.Diagnostics) > 0 {
		return fmt.Errorf("herdr did not apply the config (status %q): %s", resp.Result.Status, out)
	}
	return nil
}

func (c Client) output(args ...string) ([]byte, error) {
	return c.Runner.Output(execx.Cmd{Name: c.Bin, Args: args})
}
