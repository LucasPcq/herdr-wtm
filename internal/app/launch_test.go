package app_test

import (
	"errors"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func TestLaunchFromLinkedWorktreePassesOrigin(t *testing.T) {
	d, f, _ := newDeps(nil)
	ctx := domain.HerdrContext{
		WorkspaceID:    "w2",
		FocusedPaneCWD: "/nx/app.wt/a/src",
		Worktree:       &domain.WorktreeInfo{CheckoutPath: "/nx/app.wt/a", RepoRoot: repo, IsLinked: true},
	}
	if err := d.Launch("clean", ctx); err != nil {
		t.Fatal(err)
	}
	want := "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 90% --height 90% --env HERDR_WTM_CMD=clean --env HERDR_WTM_ORIGIN=/nx/app.wt/a --env HERDR_WTM_REPO=/nx/app"
	if len(f.Calls) != 1 || f.Calls[0].Line() != want {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestLaunchFromMainCheckoutHasNoOrigin(t *testing.T) {
	d, f, _ := newDeps(nil)
	ctx := domain.HerdrContext{Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
	if err := d.Launch("create", ctx); err != nil {
		t.Fatal(err)
	}
	want := "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 90% --height 90% --env HERDR_WTM_CMD=create --env HERDR_WTM_REPO=/nx/app"
	if f.Calls[0].Line() != want {
		t.Fatalf("got %q", f.Calls[0].Line())
	}
}

func TestLaunchFallsBackToGit(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Name == "git" {
			return []byte("/nx/app/.git\n"), nil
		}
		return nil, nil
	})
	if err := d.Launch("ui", domain.HerdrContext{WorkspaceCWD: "/nx/app/sub"}); err != nil {
		t.Fatal(err)
	}
	if f.Calls[0].Dir != "/nx/app/sub" || f.Calls[0].Name != "git" {
		t.Fatalf("call %+v", f.Calls[0])
	}
	assertHas(t, f, "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 90% --height 90% --env HERDR_WTM_CMD=ui --env HERDR_WTM_REPO=/nx/app")
}

func TestLaunchFailsOutsideGit(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Name == "git" {
			return nil, errors.New("not a git repository")
		}
		return nil, nil
	})
	if err := d.Launch("create", domain.HerdrContext{FocusedPaneCWD: "/nx/scratch"}); err == nil {
		t.Fatal("want error")
	}
	assertNoPrefix(t, f, "herdr plugin pane open")
}

func TestLaunchRejectsEmptyContextAndUnknownCommand(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Launch("create", domain.HerdrContext{}); err == nil {
		t.Fatal("empty context: want error")
	}
	if err := d.Launch("nope", domain.HerdrContext{Worktree: &domain.WorktreeInfo{RepoRoot: repo}}); err == nil {
		t.Fatal("unknown command: want error")
	}
	if len(f.Calls) != 0 {
		t.Fatalf("calls %v", f.Lines())
	}
}

func TestLaunchMenu(t *testing.T) {
	d, f, _ := newDeps(nil)
	ctx := domain.HerdrContext{Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
	if err := d.Launch("menu", ctx); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 90% --height 90% --env HERDR_WTM_CMD=menu --env HERDR_WTM_REPO=/nx/app")
}
