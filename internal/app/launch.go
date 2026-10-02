package app

import (
	"errors"
	"fmt"

	"github.com/LucasPcq/herdr-wtm/internal/gitx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
)

// Launch opens the plugin popup that will run cmd for the repository the
// action was invoked from.
func (d Deps) Launch(cmd string, ctx herdr.Context) error {
	if !IsCommand(cmd) && cmd != CmdMenu {
		return fmt.Errorf("unknown command %q", cmd)
	}
	repo, err := d.resolveRepo(ctx)
	if err != nil {
		return err
	}
	origin := ""
	if ctx.Worktree != nil && ctx.Worktree.IsLinked {
		origin = ctx.Worktree.CheckoutPath
	}
	return d.Herdr.OpenPopup(herdr.PopupParams{
		Plugin:     d.PluginID,
		Entrypoint: "run",
		Width:      d.Config.PopupWidth,
		Height:     d.Config.PopupHeight,
		Env:        map[string]string{EnvCmd: cmd, EnvRepo: repo, EnvOrigin: origin},
	})
}

func (d Deps) resolveRepo(ctx herdr.Context) (string, error) {
	if ctx.Worktree != nil && ctx.Worktree.RepoRoot != "" {
		return ctx.Worktree.RepoRoot, nil
	}
	cwd := ctx.FocusedPaneCWD
	if cwd == "" {
		cwd = ctx.WorkspaceCWD
	}
	if cwd == "" {
		return "", errors.New("herdr gave no working directory for this workspace")
	}
	return gitx.RepoRoot(d.Git, cwd)
}
