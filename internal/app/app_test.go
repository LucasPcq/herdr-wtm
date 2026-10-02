package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"slices"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

const repo = "/nx/app"

func newDeps(h func(execx.Call) ([]byte, error)) (app.Deps, *execx.Fake, *bytes.Buffer) {
	f := &execx.Fake{Handler: h}
	out := &bytes.Buffer{}
	d := app.Deps{
		Wtm:      wtm.Client{Runner: f, Bin: "wtm"},
		Herdr:    herdr.Client{Runner: f, Bin: "herdr"},
		Git:      f,
		Config:   config.Default(),
		PluginID: "lucaspcq.wtm",
		Out:      out,
		In:       strings.NewReader("\n"),
		Exists:   func(string) bool { return false },
		Log:      log.New(io.Discard, "", 0),
	}
	return d, f, out
}

func listJSON(wts ...wtm.Worktree) []byte {
	if wts == nil {
		wts = []wtm.Worktree{}
	}
	data, _ := json.Marshal(wts)
	return data
}

func workspacesJSON(ws ...herdr.Workspace) []byte {
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"workspaces": ws}})
	return data
}

func openedJSON(id string) []byte {
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"workspace": map[string]string{"workspace_id": id}}})
	return data
}

func mainWT() wtm.Worktree { return wtm.Worktree{Branch: "main", Path: repo, IsParent: true} }

func wt(branch, path string) wtm.Worktree { return wtm.Worktree{Branch: branch, Path: path} }

func primaryWS() herdr.Workspace {
	return herdr.Workspace{ID: "w1", Worktree: &herdr.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
}

func linkedWS(id, path string) herdr.Workspace {
	return herdr.Workspace{ID: id, Worktree: &herdr.WorktreeInfo{CheckoutPath: path, RepoRoot: repo, IsLinked: true}}
}

// snapshots answers successive `wtm list` calls with each snapshot in turn
// (the last one repeats).
func snapshots(lists ...[]byte) func() []byte {
	n := 0
	return func() []byte {
		i := min(n, len(lists)-1)
		n++
		return lists[i]
	}
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

func wtm0(path string) wtm.Worktree { return wtm.Worktree{Branch: "main", Path: path, IsParent: true} }
