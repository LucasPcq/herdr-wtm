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

func wtm0(path string) domain.Worktree {
	return domain.Worktree{Branch: "main", Path: path, IsParent: true}
}
