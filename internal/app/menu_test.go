package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
)

// chooser returns a fake Chooser answering pick, recording what it was shown.
func chooser(pick string, err error, seen *[]menu.Item, title *string) func(string, []menu.Item) (string, error) {
	return func(t string, items []menu.Item) (string, error) {
		if seen != nil {
			*seen = items
		}
		if title != nil {
			*title = t
		}
		return pick, err
	}
}

func TestRunMenuRunsChosenCommand(t *testing.T) {
	d, f, _ := newDeps(nil)
	var title string
	d.Choose = chooser("create", nil, nil, &title)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
	if title != "wtm · app" {
		t.Fatalf("title %q", title)
	}
	assertHas(t, f, "wtm create")
	assertNoPrefix(t, f, "herdr")
}

func TestRunMenuLabelsCleanWithOriginBranch(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	var seen []menu.Item
	d.Choose = chooser("clean", nil, &seen, nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: "/nx/app.wt/a"}); err != nil {
		t.Fatal(err)
	}
	if seen[3].Label != "Clean this worktree (feat/a)" {
		t.Fatalf("label %q", seen[3].Label)
	}
	assertHas(t, f, "wtm clean feat/a")
}

func TestRunMenuCancelledDoesNothing(t *testing.T) {
	d, f, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT()), nil
		}
		return nil, nil
	})
	d.Choose = chooser("", nil, nil, nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
	if len(f.Lines()) != 0 {
		t.Fatalf("lines %v", f.Lines())
	}
	if out.Len() != 0 {
		t.Fatalf("output %q", out.String())
	}
}

func TestRunMenuChooserErrorIsShown(t *testing.T) {
	d, f, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT()), nil
		}
		return nil, nil
	})
	d.Choose = chooser("", errors.New("menu: could not open a TTY"), nil, nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(out.String(), "could not open a TTY") || !strings.Contains(out.String(), "Press Enter") {
		t.Fatalf("output %q", out.String())
	}
	assertNoPrefix(t, f, "herdr")
}

func TestRunMenuSyncClosesStale(t *testing.T) {
	d, f, _ := newDeps(syncHandler([]domain.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/gone")}, snapshotEv()))
	d.Choose = chooser("sync", nil, nil, nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm events --output json --repo /nx/app")
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "wtm sync")
}

func TestRunMenuSyncNothingStaleNotifies(t *testing.T) {
	d, f, _ := newDeps(syncHandler([]domain.Workspace{primaryWS()}, snapshotEv()))
	d.Choose = chooser("sync", nil, nil, nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr notification show wtm --body workspaces already in sync")
}

// The signal shield must only cover the wtm command: while the menu is shown,
// a hangup (popup closed) has to end the process.
func TestRunMenuShieldsOnlyTheWtmCommand(t *testing.T) {
	var events []string
	d, _, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Interactive {
			events = append(events, c.Line())
		}
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT()), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		}
		return nil, nil
	})
	d.Choose = func(string, []menu.Item) (string, error) {
		events = append(events, "menu")
		return "prune", nil
	}
	d.Shield = func() func() {
		events = append(events, "shield")
		return func() { events = append(events, "unshield") }
	}
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(events, ","); got != "menu,shield,wtm prune,unshield" {
		t.Fatalf("events %s", got)
	}
}

func TestRunMenuCancelledNeverShields(t *testing.T) {
	d, _, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT()), nil
		}
		return nil, nil
	})
	d.Choose = chooser("", nil, nil, nil)
	d.Shield = func() func() { t.Fatal("shield installed for a cancelled menu"); return func() {} }
	if err := d.Run(app.RunParams{Cmd: domain.CmdMenu, Repo: repo, Origin: ""}); err != nil {
		t.Fatal(err)
	}
}
