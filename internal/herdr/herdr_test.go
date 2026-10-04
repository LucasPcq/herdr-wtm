package herdr_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestWorkspacesParsesFixture(t *testing.T) {
	data := fixture(t, "workspace-list.json")
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return data, nil }}
	got, err := herdr.Client{Runner: f, Bin: "herdr"}.Workspaces()
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Workspace{
		{ID: "w3", Label: "app", Worktree: &domain.WorktreeInfo{CheckoutPath: "/Users/me/dev/app", RepoRoot: "/Users/me/dev/app"}},
		{ID: "wE", Label: "feat-login", Worktree: &domain.WorktreeInfo{CheckoutPath: "/Users/me/dev/app.worktrees/feat login", RepoRoot: "/Users/me/dev/app", IsLinked: true}},
		{ID: "wF", Label: "scratch", Focused: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if f.Calls[0].Line() != "herdr workspace list" {
		t.Fatalf("call %q", f.Calls[0].Line())
	}
}

func TestOpenWorktreeReturnsWorkspaceID(t *testing.T) {
	data := fixture(t, "worktree-open.json")
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return data, nil }}
	c := herdr.Client{Runner: f, Bin: "herdr"}
	id, err := c.OpenWorktree(herdr.OpenParams{Repo: "/Users/me/dev/app", Path: "/Users/me/dev/app.worktrees/feat-new", Focus: true})
	if err != nil || id != "wG" {
		t.Fatalf("got %q, %v", id, err)
	}
	if _, err := c.OpenWorktree(herdr.OpenParams{Repo: "/r", Path: "/p"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"herdr worktree open --cwd /Users/me/dev/app --path /Users/me/dev/app.worktrees/feat-new --focus",
		"herdr worktree open --cwd /r --path /p --no-focus",
	}
	if !reflect.DeepEqual(f.Lines(), want) {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestSimpleCommands(t *testing.T) {
	f := &execx.Fake{}
	c := herdr.Client{Runner: f, Bin: "/bin/herdr"}
	if err := c.Focus("w2"); err != nil {
		t.Fatal(err)
	}
	if err := c.Close("w3"); err != nil {
		t.Fatal(err)
	}
	if err := c.Notify("opened feat-new"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/bin/herdr workspace focus w2",
		"/bin/herdr workspace close w3",
		"/bin/herdr notification show wtm --body opened feat-new",
	}
	if !reflect.DeepEqual(f.Lines(), want) {
		t.Fatalf("lines %v", f.Lines())
	}
	if !reflect.DeepEqual(f.Calls[2].Args, []string{"notification", "show", "wtm", "--body", "opened feat-new"}) {
		t.Fatalf("body must be a single arg: %v", f.Calls[2].Args)
	}
}

func TestOpenPopupSortsEnvAndSkipsEmpty(t *testing.T) {
	f := &execx.Fake{}
	err := herdr.Client{Runner: f, Bin: "herdr"}.OpenPopup(herdr.PopupParams{
		Plugin: "lucaspcq.wtm", Entrypoint: "run", Width: "90%", Height: "90%",
		Env: map[string]string{"B": "2", "A": "1", "EMPTY": ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 90% --height 90% --env A=1 --env B=2"
	if f.Calls[0].Line() != want {
		t.Fatalf("got %q", f.Calls[0].Line())
	}
}

func TestErrorsPropagate(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return nil, errors.New("socket closed") }}
	c := herdr.Client{Runner: f, Bin: "herdr"}
	if _, err := c.Workspaces(); err == nil {
		t.Fatal("Workspaces: want error")
	}
	if _, err := c.OpenWorktree(herdr.OpenParams{Repo: "/r", Path: "/p", Focus: true}); err == nil {
		t.Fatal("OpenWorktree: want error")
	}
	if err := c.Close("w1"); err == nil {
		t.Fatal("Close: want error")
	}
}

func TestParseContext(t *testing.T) {
	ctx, err := herdr.ParseContext(string(fixture(t, "context.json")))
	if err != nil {
		t.Fatal(err)
	}
	want := domain.HerdrContext{
		WorkspaceID:    "wE",
		WorkspaceCWD:   "/Users/me/dev/app.worktrees/feat login",
		FocusedPaneCWD: "/Users/me/dev/app.worktrees/feat login/src",
		Worktree:       &domain.WorktreeInfo{CheckoutPath: "/Users/me/dev/app.worktrees/feat login", RepoRoot: "/Users/me/dev/app", IsLinked: true},
	}
	if !reflect.DeepEqual(ctx, want) {
		t.Fatalf("got %+v", ctx)
	}
}

func TestParseContextEmpty(t *testing.T) {
	ctx, err := herdr.ParseContext("")
	if err != nil || !reflect.DeepEqual(ctx, domain.HerdrContext{}) {
		t.Fatalf("got %+v, %v", ctx, err)
	}
	if _, err := herdr.ParseContext("{nope"); err == nil {
		t.Fatal("want parse error")
	}
}

func TestOpenWorktreeFocusFlag(t *testing.T) {
	for focus, flag := range map[bool]string{true: "--focus", false: "--no-focus"} {
		f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte(`{"result":{"workspace":{"workspace_id":"w7"}}}`), nil }}
		id, err := herdr.Client{Runner: f, Bin: "herdr"}.OpenWorktree(herdr.OpenParams{Repo: "/nx/app", Path: "/nx/app.wt/a", Focus: focus})
		if err != nil || id != "w7" || f.Lines()[0] != "herdr worktree open --cwd /nx/app --path /nx/app.wt/a "+flag {
			t.Fatalf("id %q err %v lines %v", id, err, f.Lines())
		}
	}
}

func TestNotifyUsesWtmTitle(t *testing.T) {
	f := &execx.Fake{}
	_ = herdr.Client{Runner: f, Bin: "herdr"}.Notify("hello")
	if f.Lines()[0] != "herdr notification show wtm --body hello" {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestWorkspacesReadFocus(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) {
		return []byte(`{"result":{"workspaces":[{"workspace_id":"w1","focused":true}]}}`), nil
	}}
	ws, err := herdr.Client{Runner: f, Bin: "herdr"}.Workspaces()
	if err != nil || len(ws) != 1 || !ws[0].Focused {
		t.Fatalf("ws %+v err %v", ws, err)
	}
}
