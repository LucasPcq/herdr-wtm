package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/muesli/termenv"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

// detachPopup reopens the popup for a command chosen in the menu. The menu's
// popup must close first, and herdr shows one popup at a time, so the reopen
// runs from a process of its own that outlives the menu.
func detachPopup(p app.PopupRequest) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, domain.SubPopup)
	cmd.Env = append(os.Environ(),
		domain.EnvCmd+"="+p.Cmd, domain.EnvRepo+"="+p.Repo, domain.EnvOrigin+"="+p.Origin)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("reopen the popup: %w", err)
	}
	return cmd.Process.Release()
}

// paintSurface gives the popup a background of its own, derived from the
// terminal's, so it stands out from the panes behind it. Everything drawn in
// the popup, wtm's wizards included, sits on it.
func paintSurface() {
	out := termenv.NewOutput(os.Stdout)
	background, ok := out.BackgroundColor().(termenv.RGBColor)
	if !ok {
		return
	}
	surface, ok := rules.Surface(string(background))
	if !ok {
		return
	}
	out.SetBackgroundColor(out.Color(surface))
}
