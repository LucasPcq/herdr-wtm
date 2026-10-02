package app_test

import (
	"errors"
	"strings"
	"testing"

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
	list := snapshots(listJSON(mainWT()), listJSON(mainWT(), wt("feat/new", "/nx/app.wt/new")))
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		case "herdr worktree open --cwd /nx/app --path /nx/app.wt/new --focus":
			return openedJSON("w7"), nil
		}
		return nil, nil
	})
	var title string
	d.Choose = chooser("create", nil, nil, &title)
	if err := d.Run("menu", repo, ""); err != nil {
		t.Fatal(err)
	}
	if title != "wtm · app" {
		t.Fatalf("title %q", title)
	}
	assertHas(t, f, "wtm create")
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/new --focus")
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
	if err := d.Run("menu", repo, "/nx/app.wt/a"); err != nil {
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
	if err := d.Run("menu", repo, ""); err != nil {
		t.Fatal(err)
	}
	if strings.Join(f.Lines(), "|") != "wtm list --output json" {
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
	if err := d.Run("menu", repo, ""); err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(out.String(), "could not open a TTY") || !strings.Contains(out.String(), "Press Enter") {
		t.Fatalf("output %q", out.String())
	}
	assertNoPrefix(t, f, "herdr")
}

func TestRunMenuSyncClosesStale(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT()), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/gone")), nil
		}
		return nil, nil
	})
	d.Choose = chooser("sync", nil, nil, nil)
	if err := d.Run("menu", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "wtm sync")
}

func TestRunMenuSyncNothingStaleNotifies(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT()), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		}
		return nil, nil
	})
	d.Choose = chooser("sync", nil, nil, nil)
	if err := d.Run("menu", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr notification show wtm --body workspaces already in sync")
}
