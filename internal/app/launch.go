package app

import (
	"errors"
	"fmt"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/gitx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

type LaunchParams struct {
	Cmd     string
	Context domain.HerdrContext
}

// Launch opens the popup that runs Cmd for the repository the action was
// invoked from, starting the watcher first so the popup's changes are followed.
func (d Deps) Launch(p LaunchParams) error {
	if p.Cmd == domain.CmdBind {
		return d.Herdr.OpenPopup(herdr.PopupParams{
			Plugin: domain.PluginID, Entrypoint: domain.PopupEntrypoint,
			Width: domain.BindPopupWidth, Height: domain.BindPopupHeight,
			Env: map[string]string{domain.EnvCmd: domain.CmdBind},
		})
	}
	if !rules.IsPopupCommand(p.Cmd) {
		return fmt.Errorf("unknown command %q", p.Cmd)
	}
	d.startWatcher()
	repo, err := d.resolveRepo(p.Context)
	if err != nil {
		return err
	}
	origin := ""
	if p.Context.Worktree != nil && p.Context.Worktree.IsLinked {
		origin = p.Context.Worktree.CheckoutPath
	}
	return d.Herdr.OpenPopup(herdr.PopupParams{
		Plugin:     domain.PluginID,
		Entrypoint: domain.PopupEntrypoint,
		Width:      d.Config.PopupWidth,
		Height:     d.Config.PopupHeight,
		Env:        map[string]string{domain.EnvCmd: p.Cmd, domain.EnvRepo: repo, domain.EnvOrigin: origin},
	})
}

func (d Deps) resolveRepo(hctx domain.HerdrContext) (string, error) {
	if hctx.Worktree != nil && hctx.Worktree.RepoRoot != "" {
		return hctx.Worktree.RepoRoot, nil
	}
	cwd := hctx.FocusedPaneCWD
	if cwd == "" {
		cwd = hctx.WorkspaceCWD
	}
	if cwd == "" {
		return "", errors.New("herdr gave no working directory for this workspace")
	}
	return gitx.RepoRoot(d.Git, cwd)
}
