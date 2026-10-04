package app

import (
	"crypto/rand"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

type RunParams struct {
	Cmd    string
	Repo   string
	Origin string
}

// Run is the popup: it runs the wtm command, tagged so the watcher knows the
// user asked for it. The watcher opens, closes and focuses the workspaces.
func (d Deps) Run(p RunParams) error {
	if !rules.IsPopupCommand(p.Cmd) {
		return d.fail(fmt.Errorf("unknown command %q", p.Cmd))
	}
	if p.Repo == "" {
		return d.fail(errors.New("no repository given (" + domain.EnvRepo + " is empty)"))
	}
	branch, err := d.originBranch(p)
	if err != nil {
		return d.fail(err)
	}
	cmd := p.Cmd
	if cmd == domain.CmdMenu {
		chosen, err := d.Choose(menu.ChooseParams{Title: filepath.Base(p.Repo), Subtitle: branch, Items: menu.Items(branch)})
		if err != nil {
			return d.fail(err)
		}
		if d.Relaunch != nil && slices.Contains(domain.WtmCommands, chosen) {
			if err := d.Relaunch(PopupRequest{Cmd: chosen, Repo: p.Repo, Origin: p.Origin}); err != nil {
				return d.fail(err)
			}
			return nil
		}
		cmd = chosen
	}
	return d.runCommand(runCommandParams{Cmd: cmd, Repo: p.Repo, Branch: branch})
}

type runCommandParams struct {
	Cmd    string
	Repo   string
	Branch string
}

func (d Deps) runCommand(p runCommandParams) error {
	switch p.Cmd {
	case "":
		return nil
	case domain.CmdSync:
		return d.syncFromPopup(p.Repo)
	case domain.CmdOpen:
		return d.runOpen(p.Repo)
	}
	args := []string{p.Cmd}
	if p.Cmd == domain.CmdClean && p.Branch != "" {
		args = append(args, p.Branch)
	}
	err := d.runShielded(wtm.RunParams{Repo: p.Repo, Args: args, CorrelationID: newCorrelationID()})
	if err == nil || cancelled(err) {
		return nil
	}
	return d.fail(fmt.Errorf("wtm %s: %w", p.Cmd, err))
}

// originBranch reads wtm list only when the branch is needed: the menu's
// clean label, or clean itself.
func (d Deps) originBranch(p RunParams) (string, error) {
	if p.Origin == "" || (p.Cmd != domain.CmdMenu && p.Cmd != domain.CmdClean) {
		return "", nil
	}
	wts, err := d.Wtm.List(p.Repo)
	if err != nil {
		return "", err
	}
	return rules.BranchAt(rules.BranchAtParams{Worktrees: wts, Origin: p.Origin, FS: d.FS}), nil
}

func (d Deps) runOpen(repo string) error {
	path, err := d.Wtm.Resolve(repo)
	if cancelled(err) {
		return nil
	}
	if err != nil {
		return d.fail(err)
	}
	if path == "" {
		return nil
	}
	ws, err := d.Herdr.Workspaces()
	if err != nil {
		return d.fail(err)
	}
	if at, open := rules.WorkspaceAt(rules.WorkspaceAtParams{Workspaces: ws, Path: path, FS: d.FS}); open {
		if err := d.Herdr.Focus(at.ID); err != nil {
			return d.fail(err)
		}
		return nil
	}
	if _, err := d.Herdr.OpenWorktree(herdr.OpenParams{Repo: repo, Path: path, Focus: true}); err != nil {
		return d.fail(err)
	}
	return nil
}

// runShielded keeps SIGINT/SIGHUP off the plugin while wtm runs, and only
// then: the menu and the error prompt stay killable when the popup closes.
func (d Deps) runShielded(p wtm.RunParams) error {
	if d.Shield != nil {
		defer d.Shield()()
	}
	return d.Wtm.Run(p)
}

func newCorrelationID() string { return domain.CorrelationPrefix + rand.Text() }

// cancelled reports a wtm wizard or picker the user backed out of.
func cancelled(err error) bool {
	code, ok := execx.ExitCode(err)
	return ok && code == domain.WtmExitCancelled
}
