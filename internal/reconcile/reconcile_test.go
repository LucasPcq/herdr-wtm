package reconcile_test

import (
	"reflect"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

const repo = "/nx/app"

func wt(path string) wtm.Worktree { return wtm.Worktree{Branch: "b", Path: path} }

func linked(id, path string) herdr.Workspace {
	return herdr.Workspace{ID: id, Worktree: &herdr.WorktreeInfo{CheckoutPath: path, RepoRoot: repo, IsLinked: true}}
}

func primary(id string) herdr.Workspace {
	return herdr.Workspace{ID: id, Worktree: &herdr.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name          string
		before, after []wtm.Worktree
		ws            []herdr.Workspace
		want          reconcile.Plan
	}{
		{
			name:   "created worktree is opened",
			before: []wtm.Worktree{wt(repo)},
			after:  []wtm.Worktree{wt(repo), wt("/nx/app.wt/feat a")},
			ws:     []herdr.Workspace{primary("w1")},
			want:   reconcile.Plan{Open: []string{"/nx/app.wt/feat a"}},
		},
		{
			name:   "removed worktree's workspace is closed",
			before: []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []wtm.Worktree{wt(repo)},
			ws:     []herdr.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{Close: []string{"w2"}},
		},
		{
			name:   "ui session adds and removes",
			before: []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []wtm.Worktree{wt(repo), wt("/nx/app.wt/b")},
			ws:     []herdr.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a/")},
			want:   reconcile.Plan{Open: []string{"/nx/app.wt/b"}, Close: []string{"w2"}},
		},
		{
			name:   "already open worktree is not opened twice",
			before: []wtm.Worktree{wt(repo)},
			after:  []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			ws:     []herdr.Workspace{linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{},
		},
		{
			name:   "workspace without worktree info is ignored",
			before: []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []wtm.Worktree{wt(repo)},
			ws:     []herdr.Workspace{{ID: "w9"}},
			want:   reconcile.Plan{},
		},
		{
			name:   "nothing changed",
			before: []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			after:  []wtm.Worktree{wt(repo), wt("/nx/app.wt/a")},
			ws:     []herdr.Workspace{primary("w1"), linked("w2", "/nx/app.wt/a")},
			want:   reconcile.Plan{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcile.Diff(tt.before, tt.after, tt.ws)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDiffNeverClosesPrimary(t *testing.T) {
	got := reconcile.Diff([]wtm.Worktree{wt(repo)}, nil, []herdr.Workspace{primary("w1")})
	if len(got.Close) != 0 {
		t.Fatalf("closed primary: %+v", got)
	}
}

func TestStale(t *testing.T) {
	ws := []herdr.Workspace{
		primary("w1"),
		linked("w2", "/nx/app.wt/gone"),
		linked("w3", "/nx/app.wt/alive"),
		linked("w4", "/nx/app.wt/outside-wtm"),
		{ID: "w5", Worktree: &herdr.WorktreeInfo{CheckoutPath: "/nx/other.wt/gone", RepoRoot: "/nx/other", IsLinked: true}},
		{ID: "w6"},
	}
	current := []wtm.Worktree{wt(repo), wt("/nx/app.wt/alive")}
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
