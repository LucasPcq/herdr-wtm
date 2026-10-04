package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

const repo = "/nx/app"

func newDeps(h func(execx.Call) ([]byte, error)) (app.Deps, *execx.Fake, *bytes.Buffer) {
	f := &execx.Fake{Handler: h}
	out := &bytes.Buffer{}
	d := app.Deps{
		Wtm:    wtm.Client{Runner: f, Bin: "wtm"},
		Herdr:  herdr.Client{Runner: f, Bin: "herdr"},
		Git:    f,
		Config: config.Default(),
		Out:    out,
		In:     strings.NewReader("\n"),
		FS:     domain.FS{Normalize: filepath.Clean, Exists: func(string) bool { return false }},
		Log:    log.New(io.Discard, "", 0),
	}
	return d, f, out
}

func listJSON(wts ...domain.Worktree) []byte {
	if wts == nil {
		wts = []domain.Worktree{}
	}
	data, _ := json.Marshal(wts)
	return data
}

func workspacesJSON(ws ...domain.Workspace) []byte {
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"workspaces": ws}})
	return data
}

func openedJSON(id string) []byte {
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"workspace": map[string]string{"workspace_id": id}}})
	return data
}

func mainWT() domain.Worktree { return domain.Worktree{Branch: "main", Path: repo, IsParent: true} }

func wt(branch, path string) domain.Worktree { return domain.Worktree{Branch: branch, Path: path} }

func primaryWS() domain.Workspace {
	return domain.Workspace{ID: "w1", Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
}

func linkedWS(id, path string) domain.Workspace {
	return domain.Workspace{ID: id, Worktree: &domain.WorktreeInfo{CheckoutPath: path, RepoRoot: repo, IsLinked: true}}
}

func assertHas(t *testing.T, f *execx.Fake, line string) {
	t.Helper()
	if !slices.Contains(f.Lines(), line) {
		t.Fatalf("missing %q in %v", line, f.Lines())
	}
}

func assertNoPrefix(t *testing.T, f *execx.Fake, prefix string) {
	t.Helper()
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, prefix) {
			t.Fatalf("unexpected %q in %v", l, f.Lines())
		}
	}
}

var appRepo = domain.EventRepo{Root: repo, CommonDir: repo + "/.git"}

func snapshotEv(paths ...string) domain.Event {
	wts := []domain.EventWorktree{{Branch: "main", Path: repo, IsMain: true}}
	for _, p := range paths {
		wts = append(wts, domain.EventWorktree{Branch: filepath.Base(p), Path: p})
	}
	return domain.Event{V: 1, Type: domain.EventSnapshot, Repo: appRepo, Worktrees: wts}
}

func wtEv(typ, path, correlation string) domain.Event {
	return domain.Event{V: 1, Type: typ, Repo: appRepo, CorrelationID: correlation, Worktree: &domain.EventWorktree{Branch: filepath.Base(path), Path: path}}
}

func eventLines(evs ...domain.Event) []byte {
	var b bytes.Buffer
	for _, ev := range evs {
		data, _ := json.Marshal(ev)
		b.Write(append(data, '\n'))
	}
	return b.Bytes()
}

func panesJSON(cwds ...string) []byte {
	panes := make([]map[string]string, len(cwds))
	for i, c := range cwds {
		panes[i] = map[string]string{"cwd": c}
	}
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"panes": panes}})
	return data
}

// herdrState answers herdr's list calls with ws and cwds, and every open with w9.
func herdrState(ws []domain.Workspace, cwds ...string) func(execx.Call) ([]byte, error) {
	return func(c execx.Call) ([]byte, error) {
		switch {
		case c.Line() == "herdr workspace list":
			return workspacesJSON(ws...), nil
		case c.Line() == "herdr pane list":
			return panesJSON(cwds...), nil
		case strings.HasPrefix(c.Line(), "herdr worktree open"):
			return openedJSON("w9"), nil
		}
		return nil, nil
	}
}

func indexOf(t *testing.T, f *execx.Fake, line string) int {
	t.Helper()
	i := slices.Index(f.Lines(), line)
	if i < 0 {
		t.Fatalf("missing %q in %v", line, f.Lines())
	}
	return i
}

func focused(w domain.Workspace) domain.Workspace { w.Focused = true; return w }
