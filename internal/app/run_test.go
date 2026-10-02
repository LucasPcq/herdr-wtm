package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func TestRunCreateOpensNewWorktree(t *testing.T) {
	list := snapshots(listJSON(mainWT()), listJSON(mainWT(), wt("feat/new", "/nx/app.wt/feat new")))
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		case "herdr worktree open --cwd /nx/app --path /nx/app.wt/feat new --focus":
			return openedJSON("w7"), nil
		}
		return nil, nil
	})
	if err := d.Run("create", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm create")
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/feat new --focus")
	assertHas(t, f, "herdr notification show wtm --body opened feat new")
	for _, c := range f.Calls {
		if c.Name == "wtm" && c.Dir != repo {
			t.Fatalf("wtm ran outside the repo: %+v", c)
		}
	}
}

func TestRunCleanFromLinkedWorktreeClosesIt(t *testing.T) {
	list := snapshots(listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), listJSON(mainWT()))
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	if err := d.Run("clean", repo, "/nx/app.wt/a/"); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm clean feat/a")
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w1")
	assertHas(t, f, "herdr notification show wtm --body closed 1 workspace(s)")
}

func TestRunCleanFromMainUsesPicker(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT(), wt("feat/a", "/nx/app.wt/a")), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		}
		return nil, nil
	})
	if err := d.Run("clean", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm clean")
	assertNoPrefix(t, f, "wtm clean ")
	assertNoPrefix(t, f, "herdr workspace close")
}

func TestRunNoChangeIsSilent(t *testing.T) {
	d, f, out := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT()), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		}
		return nil, nil
	})
	if err := d.Run("prune", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "wtm prune")
	assertNoPrefix(t, f, "herdr notification")
	assertNoPrefix(t, f, "herdr worktree open")
	if out.Len() != 0 {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestRunReconcilesEvenWhenWtmFails(t *testing.T) {
	list := snapshots(listJSON(mainWT(), wt("a", "/nx/app.wt/a"), wt("b", "/nx/app.wt/b")), listJSON(mainWT(), wt("b", "/nx/app.wt/b")))
	d, f, out := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "wtm prune":
			return nil, errors.New("exit status 1")
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a"), linkedWS("w3", "/nx/app.wt/b")), nil
		}
		return nil, nil
	})
	if err := d.Run("prune", repo, ""); err == nil {
		t.Fatal("want error")
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w3")
	if !strings.Contains(out.String(), "exit status 1") || !strings.Contains(out.String(), "Press Enter") {
		t.Fatalf("output %q", out.String())
	}
}

func TestRunStopsWhenListFails(t *testing.T) {
	d, f, out := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return nil, errors.New("no .wtm config")
		}
		return nil, nil
	})
	if err := d.Run("create", repo, ""); err == nil {
		t.Fatal("want error")
	}
	assertNoPrefix(t, f, "wtm create")
	assertNoPrefix(t, f, "herdr")
	if !strings.Contains(out.String(), "no .wtm config") {
		t.Fatalf("output %q", out.String())
	}
}

func TestRunReportsHerdrFailuresAndContinues(t *testing.T) {
	list := snapshots(listJSON(mainWT(), wt("a", "/nx/app.wt/a"), wt("b", "/nx/app.wt/b")), listJSON(mainWT()))
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return list(), nil
		case "herdr workspace list":
			return workspacesJSON(linkedWS("w2", "/nx/app.wt/a"), linkedWS("w3", "/nx/app.wt/b")), nil
		case "herdr workspace close w2":
			return nil, errors.New("boom")
		}
		return nil, nil
	})
	if err := d.Run("ui", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w3")
	var note string
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, "herdr notification show wtm --body ") {
			note = l
		}
	}
	if !strings.Contains(note, "closed 1 workspace(s)") || !strings.Contains(note, "failed: close w2") {
		t.Fatalf("notification %q", note)
	}
}

func TestRunOpenFocusesAlreadyOpenWorkspace(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT(), wt("a", "/nx/app.wt/a")), nil
		case "wtm resolve":
			return []byte("/nx/app.wt/a\n"), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/a")), nil
		}
		return nil, nil
	})
	if err := d.Run("open", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace focus w2")
	assertNoPrefix(t, f, "herdr worktree open")
}

func TestRunOpenOpensClosedWorktree(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "wtm list --output json":
			return listJSON(mainWT(), wt("b", "/nx/app.wt/b")), nil
		case "wtm resolve":
			return []byte("/nx/app.wt/b\n"), nil
		case "herdr workspace list":
			return workspacesJSON(primaryWS()), nil
		case "herdr worktree open --cwd /nx/app --path /nx/app.wt/b --focus":
			return openedJSON("w8"), nil
		}
		return nil, nil
	})
	if err := d.Run("open", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr worktree open --cwd /nx/app --path /nx/app.wt/b --focus")
}

func TestRunOpenAborted(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm list --output json" {
			return listJSON(mainWT()), nil
		}
		return nil, nil
	})
	if err := d.Run("open", repo, ""); err != nil {
		t.Fatal(err)
	}
	assertNoPrefix(t, f, "herdr")
}

func TestRunUnknownCommand(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Run("rm -rf", repo, ""); err == nil {
		t.Fatal("want error")
	}
	assertNoPrefix(t, f, "wtm")
}

func TestRunWithoutRepo(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Run("create", "", ""); err == nil {
		t.Fatal("want error")
	}
	if len(f.Calls) != 0 {
		t.Fatalf("calls %v", f.Lines())
	}
}
