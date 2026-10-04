package app

import (
	"cmp"
	"errors"
	"strconv"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
)

// PopupRequest is what the popup runs: the command, its repository, and the
// worktree it was asked from.
type PopupRequest struct {
	Cmd    string
	Repo   string
	Origin string
}

// OpenPopup opens the popup that runs p.Cmd at that command's size. herdr shows
// one popup at a time and answers busy while the menu's is still closing, so
// it retries for a moment.
func (d Deps) OpenPopup(p PopupRequest) error {
	width, height := d.popupSize(p.Cmd)
	params := herdr.PopupParams{
		Plugin:     domain.PluginID,
		Entrypoint: domain.PopupEntrypoint,
		Width:      width,
		Height:     height,
		Env:        map[string]string{domain.EnvCmd: p.Cmd, domain.EnvRepo: p.Repo, domain.EnvOrigin: p.Origin},
	}
	delay := cmp.Or(d.PopupRetryDelay, domain.PopupRetryDelay)
	var err error
	for attempt := range domain.PopupOpenAttempts {
		if attempt > 0 {
			time.Sleep(delay)
		}
		err = d.Herdr.OpenPopup(params)
		if !errors.Is(err, domain.ErrHerdrBusy) {
			return err
		}
	}
	return err
}

// popupSize is the menu's fixed size, else the configured one, else the
// command's own.
func (d Deps) popupSize(cmd string) (width, height string) {
	if cmd == domain.CmdMenu {
		return strconv.Itoa(domain.MenuPopupCols), strconv.Itoa(domain.MenuPopupRows)
	}
	width, height = domain.CommandPopupWidth, domain.CommandPopupHeight
	if cmd == domain.CmdUI {
		width, height = domain.DashboardPopupSize, domain.DashboardPopupSize
	}
	return cmp.Or(d.Config.PopupWidth, width), cmp.Or(d.Config.PopupHeight, height)
}
