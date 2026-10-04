// Package wtm drives the wtm CLI.
package wtm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// Client runs wtm commands.
type Client struct {
	Runner execx.Runner
	Bin    string
}

// List returns every worktree of the repository at repo.
func (c Client) List(repo string) ([]domain.Worktree, error) {
	out, err := c.Runner.Output(execx.Cmd{Dir: repo, Name: c.Bin, Args: []string{"list", "--output", "json"}})
	if err != nil {
		return nil, fmt.Errorf("wtm list: %w", err)
	}
	var wts []domain.Worktree
	if err := json.Unmarshal(out, &wts); err != nil {
		return nil, fmt.Errorf("parse wtm list output: %w", err)
	}
	return wts, nil
}

// Run runs an interactive wtm command in repo.
func (c Client) Run(repo string, args ...string) error {
	return c.Runner.Interactive(execx.Cmd{Dir: repo, Name: c.Bin, Args: args})
}

// Resolve shows wtm's worktree picker and returns the chosen path, or "" when
// the user aborts.
func (c Client) Resolve(repo string) (string, error) {
	var out bytes.Buffer
	if err := c.Runner.Interactive(execx.Cmd{Dir: repo, Stdout: &out, Name: c.Bin, Args: []string{"resolve"}}); err != nil {
		return "", fmt.Errorf("wtm resolve: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}
