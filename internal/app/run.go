package app

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Run is the popup entrypoint: run the wtm command, then sync herdr workspaces
// with what changed.
func (d Deps) Run(cmd, repo, origin string) error {
	if !IsCommand(cmd) && cmd != CmdMenu {
		return d.fail(fmt.Errorf("unknown command %q", cmd))
	}
	if repo == "" {
		return d.fail(errors.New("no repository given (" + EnvRepo + " is empty)"))
	}
	before, err := d.Wtm.List(repo)
	if err != nil {
		return d.fail(err)
	}
	if cmd == CmdMenu {
		chosen, err := d.Choose("wtm · "+filepath.Base(repo), menu.Items(branchAt(before, origin)))
		if err != nil {
			return d.fail(err)
		}
		switch chosen {
		case "":
			return nil
		case CmdSync:
			return d.syncRepo(repo)
		}
		cmd = chosen
	}
	if cmd == "open" {
		return d.runOpen(repo)
	}

	args := []string{cmd}
	if cmd == "clean" {
		if branch := branchAt(before, origin); branch != "" {
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
	d.apply(repo, reconcile.Diff(before, after, ws, d.Exists))
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
	target := reconcile.Normalize(path)
	for _, w := range ws {
		if w.Worktree != nil && reconcile.Normalize(w.Worktree.CheckoutPath) == target {
			if err := d.Herdr.Focus(w.ID); err != nil {
				return d.fail(err)
			}
			return nil
		}
	}
	if _, err := d.Herdr.OpenWorktree(repo, path, true); err != nil {
		return d.fail(err)
	}
	return nil
}

// branchAt returns the branch of the non-primary worktree at origin, or "".
func branchAt(wts []wtm.Worktree, origin string) string {
	if origin == "" {
		return ""
	}
	target := reconcile.Normalize(origin)
	for _, wt := range wts {
		if !wt.IsParent && reconcile.Normalize(wt.Path) == target {
			return wt.Branch
		}
	}
	return ""
}

// runShielded runs the wtm command under the signal shield, and only it: the
// menu and the error prompt stay killable by a hangup when the popup closes.
func (d Deps) runShielded(repo string, args []string) error {
	if d.Shield != nil {
		defer d.Shield()()
	}
	return d.Wtm.Run(repo, args...)
}
