package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Run is the popup entrypoint: run the wtm command, then sync herdr workspaces
// with what changed.
func (d Deps) Run(cmd, repo, origin string) error {
	if !rules.IsPopupCommand(cmd) {
		return d.fail(fmt.Errorf("unknown command %q", cmd))
	}
	if repo == "" {
		return d.fail(errors.New("no repository given (" + domain.EnvRepo + " is empty)"))
	}
	before, err := d.Wtm.List(repo)
	if err != nil {
		return d.fail(err)
	}
	if cmd == domain.CmdMenu {
		chosen, err := d.Choose("wtm · "+filepath.Base(repo), menu.Items(rules.BranchAt(rules.BranchAtParams{Worktrees: before, Origin: origin, FS: d.FS})))
		if err != nil {
			return d.fail(err)
		}
		switch chosen {
		case "":
			return nil
		case domain.CmdSync:
			return d.syncRepo(repo)
		}
		cmd = chosen
	}
	if cmd == domain.CmdOpen {
		return d.runOpen(repo)
	}

	args := []string{cmd}
	if cmd == domain.CmdClean {
		if branch := rules.BranchAt(rules.BranchAtParams{Worktrees: before, Origin: origin, FS: d.FS}); branch != "" {
			args = append(args, branch)
		}
	}
	cmdErr := d.runShielded(repo, args)

	after, err := d.Wtm.List(repo)
	if err != nil {
		return d.fail(errors.Join(cmdErr, err))
	}
	ws, err := d.Herdr.Workspaces()
	if err != nil {
		return d.fail(errors.Join(cmdErr, err))
	}
	closed, opened := d.apply(repo, reconcile.Diff(reconcile.DiffParams{Before: before, After: after, Workspaces: ws, FS: d.FS}))
	if !opened && d.closedOrigin(ws, closed, origin) {
		if err := d.focusMain(repo, ws); err != nil {
			d.Log.Printf("focus main checkout: %v", err)
			d.notify(fmt.Sprintf("could not focus the main checkout: %v", err))
		}
	}
	if cmdErr != nil {
		return d.fail(fmt.Errorf("wtm %s: %w", cmd, cmdErr))
	}
	return nil
}

func (d Deps) runOpen(repo string) error {
	path, err := d.Wtm.Resolve(repo)
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
	target := d.FS.Normalize(path)
	for _, w := range ws {
		if w.Worktree != nil && d.FS.Normalize(w.Worktree.CheckoutPath) == target {
			if err := d.Herdr.Focus(w.ID); err != nil {
				return d.fail(err)
			}
			return nil
		}
	}
	if _, err := d.Herdr.OpenWorktree(herdr.OpenParams{Repo: repo, Path: path, Focus: true}); err != nil {
		return d.fail(err)
	}
	return nil
}

// runShielded runs the wtm command under the signal shield, and only it: the
// menu and the error prompt stay killable by a hangup when the popup closes.
func (d Deps) runShielded(repo string, args []string) error {
	if d.Shield != nil {
		defer d.Shield()()
	}
	return d.Wtm.Run(wtm.RunParams{Repo: repo, Args: args})
}

// closedOrigin reports whether one of the closed workspaces is the worktree
// the action was invoked from.
func (d Deps) closedOrigin(ws []domain.Workspace, closed []string, origin string) bool {
	if origin == "" {
		return false
	}
	target := d.FS.Normalize(origin)
	for _, w := range ws {
		if w.Worktree != nil && slices.Contains(closed, w.ID) && d.FS.Normalize(w.Worktree.CheckoutPath) == target {
			return true
		}
	}
	return false
}
