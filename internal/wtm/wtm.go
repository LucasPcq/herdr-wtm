// Package wtm drives the wtm CLI. wtm's argv is spelled only here.
package wtm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

type Client struct {
	Runner execx.Runner
	Bin    string
}

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

type RunParams struct {
	Repo          string
	Args          []string
	CorrelationID string
}

// Run runs an interactive wtm command, its events tagged with CorrelationID.
func (c Client) Run(p RunParams) error {
	var env []string
	if p.CorrelationID != "" {
		env = []string{domain.EnvCorrelationID + "=" + p.CorrelationID}
	}
	return c.Runner.Interactive(execx.Cmd{Dir: p.Repo, Name: c.Bin, Args: p.Args, Env: env})
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

// Contracts reads which machine contracts the installed wtm speaks.
func (c Client) Contracts() (domain.Contracts, error) {
	out, err := c.Runner.Output(execx.Cmd{Name: c.Bin, Args: []string{"version", "--output", "json"}})
	if err != nil {
		return domain.Contracts{}, fmt.Errorf("wtm version: %w", err)
	}
	var contracts domain.Contracts
	if err := json.Unmarshal(out, &contracts); err != nil {
		return domain.Contracts{}, fmt.Errorf("parse wtm version output: %w", err)
	}
	return contracts, nil
}

type EventsParams struct {
	// Repo is the repository to follow; empty follows every repository wtm knows.
	Repo    string
	OnEvent func(domain.Event)
}

// Events reads `wtm events --output json` until it ends or ctx is cancelled.
// A line that does not decode is skipped, as the contract asks.
func (c Client) Events(ctx context.Context, p EventsParams) error {
	cmd := execx.Cmd{Dir: domain.GlobalStreamDir, Name: c.Bin, Args: []string{"events", "--output", "json"}}
	if p.Repo != "" {
		cmd.Dir = p.Repo
		cmd.Args = append(cmd.Args, "--repo", p.Repo)
	}
	err := c.Runner.Stream(ctx, cmd, func(line []byte) {
		var ev domain.Event
		if json.Unmarshal(line, &ev) == nil {
			p.OnEvent(ev)
		}
	})
	if err != nil {
		return fmt.Errorf("wtm events: %w", err)
	}
	return nil
}
