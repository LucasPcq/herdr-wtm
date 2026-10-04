package app

import (
	"errors"
	"fmt"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/gitx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

// Launch opens the plugin popup that will run cmd for the repository the
// action was invoked from.
func (d Deps) Launch(cmd string, ctx domain.HerdrContext) error {
	if cmd == domain.CmdBind {
		// Binding a key needs no repository: a small popup with the prompt.
		return d.Herdr.OpenPopup(herdr.PopupParams{
			Plugin: domain.PluginID, Entrypoint: domain.PopupEntrypoint, Width: domain.BindPopupWidth, Height: domain.BindPopupHeight,
			Env: map[string]string{domain.EnvCmd: domain.CmdBind},
		})
	}
	if !rules.IsPopupCommand(cmd) {
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
		Plugin:     domain.PluginID,
		Entrypoint: domain.PopupEntrypoint,
		Width:      d.Config.PopupWidth,
		Height:     d.Config.PopupHeight,
		Env:        map[string]string{domain.EnvCmd: cmd, domain.EnvRepo: repo, domain.EnvOrigin: origin},
	})
}

func (d Deps) resolveRepo(ctx domain.HerdrContext) (string, error) {
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
