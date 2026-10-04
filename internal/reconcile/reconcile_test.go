package reconcile_test

import (
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
			got := reconcile.Diff(tt.before, tt.after, tt.ws, gone)
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
	if got := reconcile.Diff(before, after, ws, exists); len(got.Close) != 0 {
		t.Fatalf("closed a workspace whose folder is still on disk: %+v", got)
	}
}

func TestDiffNeverClosesPrimary(t *testing.T) {
	got := reconcile.Diff([]domain.Worktree{wt(repo)}, nil, []domain.Workspace{primary("w1")}, gone)
	if len(got.Close) != 0 {
		t.Fatalf("closed primary: %+v", got)
	}
}

func TestStale(t *testing.T) {
	ws := []domain.Workspace{
		primary("w1"),
		linked("w2", "/nx/app.wt/gone"),
		linked("w3", "/nx/app.wt/alive"),
		linked("w4", "/nx/app.wt/outside-wtm"),
		{ID: "w5", Worktree: &domain.WorktreeInfo{CheckoutPath: "/nx/other.wt/gone", RepoRoot: "/nx/other", IsLinked: true}},
		{ID: "w6"},
	}
	current := []domain.Worktree{wt(repo), wt("/nx/app.wt/alive")}
	exists := func(p string) bool { return p == "/nx/app.wt/outside-wtm" }
	got := reconcile.Stale(repo, current, ws, exists)
	want := reconcile.Plan{Close: []string{"w2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPlanEmpty(t *testing.T) {
	if !(reconcile.Plan{}).Empty() || (reconcile.Plan{Close: []string{"w1"}}).Empty() {
		t.Fatal("Empty is wrong")
	}
}
