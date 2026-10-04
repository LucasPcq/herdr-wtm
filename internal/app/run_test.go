package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func wtmCall(t *testing.T, f *execx.Fake, line string) execx.Call {
	t.Helper()
	for _, c := range f.Calls {
		if c.Line() == line {
			return c
		}
	}
	t.Fatalf("missing %q in %v", line, f.Lines())
	return execx.Call{}
}

func TestRunTagsCommandAndLeavesWorkspacesToTheWatcher(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdCreate, Repo: repo}); err != nil {
		t.Fatal(err)
	}
	c := wtmCall(t, f, "wtm create")
	if c.Dir != repo || len(c.Env) != 1 || !strings.HasPrefix(c.Env[0], "WTM_CORRELATION_ID=herdr-wtm:") {
		t.Fatalf("call %+v", c)
	}
	assertNoPrefix(t, f, "wtm list")
	assertNoPrefix(t, f, "herdr")
}

func TestRunGivesEachCommandItsOwnCorrelation(t *testing.T) {
	d, f, _ := newDeps(nil)
	_ = d.Run(app.RunParams{Cmd: domain.CmdPrune, Repo: repo})
	_ = d.Run(app.RunParams{Cmd: domain.CmdPrune, Repo: repo})
	if f.Calls[0].Env[0] == f.Calls[1].Env[0] {
		t.Fatalf("same correlation twice: %v", f.Calls[0].Env)
	}
}

func TestRunCleanFromLinkedWorktreeNamesItsBranch(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdClean, Repo: repo, Origin: "/nx/app.wt/a/"}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm clean feat/a")
	assertNoPrefix(t, f, "herdr")
}

func TestRunCleanFromMainUsesWtmPicker(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Run(app.RunParams{Cmd: domain.CmdClean, Repo: repo}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm clean")
	assertNoPrefix(t, f, "wtm list")
}

func TestRunShowsWtmFailureInPopup(t *testing.T) {
	d, _, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm prune" {
			return nil, errors.New("exit status 1")
		}
		return nil, nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdPrune, Repo: repo}); err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(out.String(), "exit status 1") || !strings.Contains(out.String(), "Press Enter") {
		t.Fatalf("output %q", out.String())
	}
}

func TestRunRefusesUnknownCommand(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Run(app.RunParams{Cmd: "bogus", Repo: repo}); err == nil {
		t.Fatal("want error")
	}
	assertNoPrefix(t, f, "wtm")
}

func TestRunOpenFocusesExistingWorkspace(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm resolve":
			return []byte("/nx/app.wt/a\n"), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdOpen, Repo: repo}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace focus w2")
}

func TestRunOpenOpensWithFocus(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm resolve":
			return []byte("/nx/app.wt/a\n"), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		}
		return openedJSON("w9"), nil
	})
	if err := d.Run(app.RunParams{Cmd: domain.CmdOpen, Repo: repo}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/a --focus")
}
