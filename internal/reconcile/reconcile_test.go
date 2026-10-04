package reconcile_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
)

const repo = "/nx/app"

func wt(path string) domain.Worktree { return domain.Worktree{Branch: "b", Path: path} }

func linked(id, path string) domain.Workspace {
	return domain.Workspace{ID: id, Worktree: &domain.WorktreeInfo{CheckoutPath: path, RepoRoot: repo, IsLinked: true}}
}

func primary(id string) domain.Workspace {
	return domain.Workspace{ID: id, Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name          string
		before, after []domain.Worktree
		ws            []domain.Workspace
		want          reconcile.Plan
	}{
		{
			name:   "created worktree is opened",
			before: []domain.Worktree{wt(repo)},
			after:  []domain.Worktree{wt(repo), wt("/nx/app.wt/feat a")},
			ws:     []domain.Workspace{primary("w1")},
			want:   reconcile.Plan{Open: []string{"/nx/app.wt/feat a"}},
		},
		{
			name:   "removed worktree's workspace is closed",
			before: []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []domain.Worktree{wt(repo)},
			ws:     []domain.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{Close: []string{"w2"}},
		},
		{
			name:   "ui session adds and removes",
			before: []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []domain.Worktree{wt(repo), wt("/nx/app.wt/b")},
			ws:     []domain.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a/")},
			want:   reconcile.Plan{Open: []string{"/nx/app.wt/b"}, Close: []string{"w2"}},
		},
		{
			name:   "already open worktree is not opened twice",
			before: []domain.Worktree{wt(repo)},
			after:  []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			ws:     []domain.Workspace{linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{},
		},
		{
			name:   "workspace without worktree info is ignored",
			before: []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []domain.Worktree{wt(repo)},
			ws:     []domain.Workspace{{ID: "w9"}},
			want:   reconcile.Plan{},
		},
		{
			name:   "nothing changed",
			before: []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []domain.Worktree{wt(repo), wt("/nx/app.wt/a")},
			ws:     []domain.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcile.Diff(reconcile.DiffParams{Before: tt.before, After: tt.after, Workspaces: tt.ws, FS: fsWith(gone)})
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func gone(string) bool { return false }

func TestDiffKeepsWorkspaceWhoseFolderStillExists(t *testing.T) {
	before := []domain.Worktree{wt(repo), wt("/nx/app.wt/a")}
	after := []domain.Worktree{wt(repo)}
	ws := []domain.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a")}
	exists := func(p string) bool { return p == "/nx/app.wt/a" }
	if got := reconcile.Diff(reconcile.DiffParams{Before: before, After: after, Workspaces: ws, FS: fsWith(exists)}); len(got.Close) != 0 {
		t.Fatalf("closed a workspace whose folder is still on disk: %+v", got)
	}
}

func TestDiffNeverClosesPrimary(t *testing.T) {
	got := reconcile.Diff(reconcile.DiffParams{Before: []domain.Worktree{wt(repo)}, Workspaces: []domain.Workspace{primary("w1")}, FS: fsWith(gone)})
	if len(got.Close) != 0 {
		t.Fatalf("closed primary: %+v", got)
	}
}

func TestPlanEmpty(t *testing.T) {
	if !(reconcile.Plan{}).Empty() || (reconcile.Plan{Close: []string{"w1"}}).Empty() {
		t.Fatal("Empty is wrong")
	}
}

func fsWith(exists func(string) bool) domain.FS {
	return domain.FS{Normalize: filepath.Clean, Exists: exists}
}
