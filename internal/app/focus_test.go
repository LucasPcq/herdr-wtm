package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
)

// cleanWorld: worktree /nx/app.wt/a disappears during the command; ws lists the workspaces herdr shows.
func cleanWorld(ws []herdr.Workspace, failOn string) func(execx.Call) ([]byte, error) {
	list := snapshots(listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a"), wt("feat/b", "/nx/app.wt/b")), listJSON(mainWT(), wt("feat/b", "/nx/app.wt/b")))
	return func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(ws...), nil
		case "herdr worktree open --cwd /nx/app --path /nx/app --focus":
			if failOn == "open" {
				return nil, errors.New("boom")
			}
			return openedJSON("w9"), nil
		case "herdr workspace focus w1":
			if failOn == "focus" {
				return nil, errors.New("boom")
			}
		}
		return nil, nil
	}
}

func TestCleanOfOriginFocusesMainCheckout(t *testing.T) {
	d, f, _ := newDeps(cleanWorld([]herdr.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a")}, ""))
	if err := d.Run("clean", repo, "/nx/app.wt/a"); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertHas(t, f, "herdr workspace focus w1")
	lines := strings.Join(f.Lines(), "|")
	if strings.Index(lines, "workspace close w2") > strings.Index(lines, "workspace focus w1") {
		t.Fatalf("focus before close: %v", f.Lines())
	}
}

func TestCleanOfOriginOpensMainCheckoutWhenNotOpen(t *testing.T) {
	d, f, _ := newDeps(cleanWorld([]herdr.Workspace{linkedWS("w2", "/nx/app.wt/a")}, ""))
	if err := d.Run("clean", repo, "/nx/app.wt/a"); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app --focus")
}

func TestCleanOfAnotherWorktreeKeepsFocus(t *testing.T) {
	d, f, _ := newDeps(cleanWorld([]herdr.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a"), linkedWS("w3", "/nx/app.wt/b")}, ""))
	if err := d.Run("clean", repo, "/nx/app.wt/b"); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace focus")
	assertNoPrefix(t, f, "herdr worktree open")
}

func TestOriginNotClosedWhenCloseFails(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "herdr workspace close w2" {
			return nil, errors.New("busy")
		}
		return cleanWorld([]herdr.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a")}, "")(c)
	})
	if err := d.Run("clean", repo, "/nx/app.wt/a"); err != nil {
		t.Fatal(err)
	}
	assertNoPrefix(t, f, "herdr workspace focus")
}

func TestFocusMainFailureIsReported(t *testing.T) {
	d, f, _ := newDeps(cleanWorld([]herdr.Workspace{primaryWS(), linkedWS("w2", "/nx/app.wt/a")}, "focus"))
	if err := d.Run("clean", repo, "/nx/app.wt/a"); err != nil {
		t.Fatal(err)
	}
	var reported bool
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, "herdr notification show wtm --body could not focus the main checkout") {
			reported = true
		}
	}
	if !reported {
		t.Fatalf("no failure notification: %v", f.Lines())
	}
}

func TestNewWorkspaceKeepsFocusOverMainCheckout(t *testing.T) {
	list := snapshots(listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), listJSON(mainWT(), wt("feat/c", "/nx/app.wt/c")))
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a")), nil
		case "herdr worktree open --cwd /nx/app --path /nx/app.wt/c --focus":
			return openedJSON("w4"), nil
		}
		return nil, nil
	})
	if err := d.Run("ui", repo, "/nx/app.wt/a"); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace focus")
}
