// Package wtm drives the wtm CLI.
package wtm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// Worktree is the part of `wtm list --output json` the plugin uses.
type Worktree struct {
	Branch   string `json:"branch"`
	Path     string `json:"path"`
	IsParent bool   `json:"is_parent"`
}

// Client runs wtm commands.
type Client struct {
	Runner execx.Runner
	Bin    string
}

// List returns every worktree of the repository at repo.
func (c Client) List(repo string) ([]Worktree, error) {
	out, err := c.Runner.Output(repo, c.Bin, "list", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("wtm list: %w", err)
	}
	var wts []Worktree
	if err := json.Unmarshal(out, &wts); err != nil {
		return nil, fmt.Errorf("parse wtm list output: %w", err)
	}
	return wts, nil
}

// Run runs an interactive wtm command in repo.
func (c Client) Run(repo string, args ...string) error {
	return c.Runner.Interactive(repo, nil, c.Bin, args...)
}

// Resolve shows wtm's worktree picker and returns the chosen path, or "" when
// the user aborts.
func (c Client) Resolve(repo string) (string, error) {
	var out bytes.Buffer
	if err := c.Runner.Interactive(repo, &out, c.Bin, "resolve"); err != nil {
		return "", fmt.Errorf("wtm resolve: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}
