package rules_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

const repo = "/nx/app"

// testFS spells paths through aliases (as symlinks would) and reports onDisk as existing.
func testFS(aliases map[string]string, onDisk ...string) domain.FS {
	return domain.FS{
		Normalize: func(p string) string {
			p = filepath.Clean(p)
			if a, ok := aliases[p]; ok {
				return a
			}
			return p
		},
		Exists: func(p string) bool { return slices.Contains(onDisk, p) },
	}
}

func primary() domain.Workspace {
	return domain.Workspace{ID: "w1", Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
}

func linked(id, path string) domain.Workspace {
	return domain.Workspace{ID: id, Worktree: &domain.WorktreeInfo{CheckoutPath: path, RepoRoot: repo, IsLinked: true}}
}

func TestWorkspaceAtNormalizes(t *testing.T) {
	fs := testFS(map[string]string{"/tmp/app.wt/a": "/private/tmp/app.wt/a"})
	ws := []domain.Workspace{linked("w2", "/private/tmp/app.wt/a/")}
	got, ok := rules.WorkspaceAt(rules.WorkspaceAtParams{Workspaces: ws, Path: "/tmp/app.wt/a", FS: fs})
	if !ok || got.ID != "w2" {
		t.Fatalf("got %+v ok %v", got, ok)
	}
}

func TestWorkspaceAtIgnoresPlainWorkspaces(t *testing.T) {
	ws := []domain.Workspace{{ID: "w5"}}
	if _, ok := rules.WorkspaceAt(rules.WorkspaceAtParams{Workspaces: ws, Path: repo, FS: testFS(nil)}); ok {
		t.Fatal("a workspace without worktree info matches no path")
	}
}

func TestMainWorkspace(t *testing.T) {
	ws := []domain.Workspace{linked("w2", "/nx/app.wt/a"), primary()}
	got, ok := rules.MainWorkspace(rules.MainWorkspaceParams{Workspaces: ws, RepoRoot: repo, FS: testFS(nil)})
	if !ok || got.ID != "w1" {
		t.Fatalf("got %+v ok %v", got, ok)
	}
	if _, ok := rules.MainWorkspace(rules.MainWorkspaceParams{Workspaces: ws[:1], RepoRoot: repo, FS: testFS(nil)}); ok {
		t.Fatal("no main workspace open")
	}
}

func TestRepoShownByWorkspaceRepoRoot(t *testing.T) {
	p := rules.RepoShownParams{RepoRoot: repo, Workspaces: []domain.Workspace{linked("w2", "/nx/app.wt/a")}, FS: testFS(nil)}
	if !rules.RepoShown(p) {
		t.Fatal("a linked workspace of the repo shows it")
	}
}

func TestRepoShownByPaneCWD(t *testing.T) {
	p := rules.RepoShownParams{
		RepoRoot: repo, WorktreePaths: []string{repo, "/nx/app.wt/a"},
		Workspaces: []domain.Workspace{{ID: "w5"}}, PaneCWDs: []string{"/nx/app.wt/a/src"}, FS: testFS(nil),
	}
	if !rules.RepoShown(p) {
		t.Fatal("a pane inside a worktree shows the repo")
	}
}

func TestRepoShownRootCountsWithoutSnapshot(t *testing.T) {
	p := rules.RepoShownParams{RepoRoot: repo, PaneCWDs: []string{repo}, FS: testFS(nil)}
	if !rules.RepoShown(p) {
		t.Fatal("a pane at the repo root shows it")
	}
}

func TestRepoShownNotBySiblingPrefix(t *testing.T) {
	p := rules.RepoShownParams{RepoRoot: repo, WorktreePaths: []string{repo}, PaneCWDs: []string{"/nx/app2"}, FS: testFS(nil)}
	if rules.RepoShown(p) {
		t.Fatal("/nx/app2 is not inside /nx/app")
	}
}

func TestStale(t *testing.T) {
	ws := []domain.Workspace{
		primary(),
		linked("w2", "/nx/app.wt/gone"),
		linked("w3", "/nx/app.wt/kept"),
		linked("w4", "/nx/app.wt/ondisk"),
		{ID: "w6", Worktree: &domain.WorktreeInfo{CheckoutPath: "/other.wt/x", RepoRoot: "/other", IsLinked: true}},
	}
	got := rules.Stale(rules.StaleParams{RepoRoot: repo, Current: []string{repo, "/nx/app.wt/kept"}, Workspaces: ws, FS: testFS(nil, "/nx/app.wt/ondisk")})
	if !slices.Equal(got, []string{"w2"}) {
		t.Fatalf("got %v", got)
	}
}

func TestBranchAt(t *testing.T) {
	wts := []domain.Worktree{{Branch: "main", Path: repo, IsParent: true}, {Branch: "feat/a", Path: "/nx/app.wt/a"}}
	for origin, want := range map[string]string{"/nx/app.wt/a/": "feat/a", repo: "", "": ""} {
		if got := rules.BranchAt(rules.BranchAtParams{Worktrees: wts, Origin: origin, FS: testFS(nil)}); got != want {
			t.Errorf("origin %q: got %q want %q", origin, got, want)
		}
	}
}

func TestIsOwnCorrelation(t *testing.T) {
	if !rules.IsOwnCorrelation("herdr-wtm:abc") || rules.IsOwnCorrelation("") || rules.IsOwnCorrelation("popup-42") {
		t.Fatal("only herdr-wtm: ids are ours")
	}
}

func TestProvisionFailure(t *testing.T) {
	no, yes, code := false, true, 1
	wt := &domain.EventWorktree{Branch: "feat/x", Path: "/nx/app.wt/x"}
	cases := []struct {
		ev   domain.Event
		want string
		ok   bool
	}{
		{domain.Event{Type: domain.EventProvisioned, Worktree: wt, OK: &no, Hook: "pnpm install", ExitCode: &code}, "feat/x: on_create failed — pnpm install (exit 1)", true},
		{domain.Event{Type: domain.EventProvisioned, Worktree: wt, OK: &no}, "feat/x: on_create failed", true},
		{domain.Event{Type: domain.EventProvisioned, Worktree: wt, OK: &yes}, "", false},
		{domain.Event{Type: domain.EventCreated, Worktree: wt, OK: &no}, "", false},
	}
	for _, c := range cases {
		got, ok := rules.ProvisionFailure(c.ev)
		if got != c.want || ok != c.ok {
			t.Errorf("%+v: got %q %v", c.ev, got, ok)
		}
	}
}

func TestSummary(t *testing.T) {
	got := rules.Summary(rules.SummaryParams{Opened: []string{"a", "b"}, Closed: 2, Failures: []string{"close w3: boom"}})
	if got != "opened a, b · closed 2 workspace(s) · failed: close w3: boom" {
		t.Fatalf("got %q", got)
	}
	if rules.Summary(rules.SummaryParams{}) != "" {
		t.Fatal("nothing to say is empty")
	}
}

func TestIsPopupCommand(t *testing.T) {
	if !rules.IsPopupCommand(domain.CmdCreate) || !rules.IsPopupCommand(domain.CmdMenu) || rules.IsPopupCommand("bogus") || rules.IsPopupCommand(domain.CmdBind) {
		t.Fatal("popup commands are the wtm commands and the menu")
	}
}
