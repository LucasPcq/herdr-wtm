package app_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func popupLine(t *testing.T, f *execx.Fake) string {
	t.Helper()
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, "herdr plugin pane open") {
			return l
		}
	}
	t.Fatalf("no popup opened: %v", f.Lines())
	return ""
}

func TestLaunchSizesThePopupForItsCommand(t *testing.T) {
	cases := map[string]string{
		domain.CmdMenu:   "--width 52 --height 21",
		domain.CmdCreate: "--width 100 --height 30",
		domain.CmdUI:     "--width 90% --height 90%",
	}
	for cmd, size := range cases {
		d, f, _ := newDeps(nil)
		if err := d.Launch(app.LaunchParams{Cmd: cmd, Context: mainCtx}); err != nil {
			t.Fatal(err)
		}
		if l := popupLine(t, f); !strings.Contains(l, size) {
			t.Errorf("%s: %s", cmd, l)
		}
	}
}

func TestLaunchConfiguredSizeAppliesToCommandsNotTheMenu(t *testing.T) {
	for cmd, size := range map[string]string{domain.CmdCreate: "--width 120 --height 40", domain.CmdMenu: "--width 52 --height 21"} {
		d, f, _ := newDeps(nil)
		d.Config.PopupWidth, d.Config.PopupHeight = "120", "40"
		if err := d.Launch(app.LaunchParams{Cmd: cmd, Context: mainCtx}); err != nil {
			t.Fatal(err)
		}
		if l := popupLine(t, f); !strings.Contains(l, size) {
			t.Errorf("%s: %s", cmd, l)
		}
	}
}

func busyThen(n int) (func(execx.Call) ([]byte, error), *int) {
	calls := 0
	return func(c execx.Call) ([]byte, error) {
		if !strings.HasPrefix(c.Line(), "herdr plugin pane open") {
			return nil, nil
		}
		calls++
		if calls <= n {
			return nil, errors.New(`exit status 1: {"error":{"code":"ui_busy"}}`)
		}
		return nil, nil
	}, &calls
}

func TestOpenPopupWaitsForTheMenuToClose(t *testing.T) {
	h, calls := busyThen(2)
	d, _, _ := newDeps(h)
	d.PopupRetryDelay = time.Millisecond
	if err := d.OpenPopup(app.PopupRequest{Cmd: domain.CmdCreate, Repo: repo}); err != nil {
		t.Fatal(err)
	}
	if *calls != 3 {
		t.Fatalf("calls %d", *calls)
	}
}

func TestOpenPopupGivesUpWhenHerdrStaysBusy(t *testing.T) {
	h, calls := busyThen(1000)
	d, _, _ := newDeps(h)
	d.PopupRetryDelay = time.Millisecond
	if err := d.OpenPopup(app.PopupRequest{Cmd: domain.CmdCreate, Repo: repo}); !errors.Is(err, domain.ErrHerdrBusy) {
		t.Fatalf("err %v", err)
	}
	if *calls != domain.PopupOpenAttempts {
		t.Fatalf("calls %d", *calls)
	}
}

func TestRunMenuReopensThePopupForTheChosenCommand(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	d.Choose = chooser(domain.CmdClean, nil, nil, nil)
	var got app.PopupRequest
	d.Relaunch = func(p app.PopupRequest) error { got = p; return nil }
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: "/nx/app.wt/a"}); err != nil {
		t.Fatal(err)
	}
	if got != (app.PopupRequest{Cmd: domain.CmdClean, Repo: repo, Origin: "/nx/app.wt/a"}) {
		t.Fatalf("relaunch %+v", got)
	}
	assertNoPrefix(t, f, "wtm clean")
}

func TestRunCancelledWizardClosesSilently(t *testing.T) {
	d, _, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm create" {
			return nil, execx.ExitError{Code: domain.WtmExitCancelled}
		}
		return nil, nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdCreate, Repo: repo}); err != nil {
		t.Fatalf("err %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("output %q", out.String())
	}
}

func TestRunOpenCancelledPickerClosesSilently(t *testing.T) {
	d, _, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm resolve" {
			return nil, execx.ExitError{Code: domain.WtmExitCancelled}
		}
		return nil, nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdOpen, Repo: repo}); err != nil || out.Len() != 0 {
		t.Fatalf("err %v output %q", err, out.String())
	}
}
