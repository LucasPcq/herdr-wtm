package wtm_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

func TestListParsesFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/list.json")
	if err != nil {
		t.Fatal(err)
	}
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return data, nil }}
	got, err := wtm.Client{Runner: f, Bin: "wtm"}.List("/Users/me/dev/app")
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Worktree{
		{Branch: "main", Path: "/Users/me/dev/app", IsParent: true},
		{Branch: "feat/login", Path: "/Users/me/dev/app.worktrees/feat login"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if f.Calls[0].Line() != "wtm list --output json" || f.Calls[0].Dir != "/Users/me/dev/app" || f.Calls[0].Interactive {
		t.Fatalf("call %+v", f.Calls[0])
	}
}

func TestListPropagatesFailure(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return nil, errors.New("not a wtm project") }}
	_, err := wtm.Client{Runner: f, Bin: "wtm"}.List("/repo")
	if err == nil || !strings.Contains(err.Error(), "not a wtm project") {
		t.Fatalf("got %v", err)
	}
}

func TestListRejectsGarbage(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte("Update available!\n[]"), nil }}
	if _, err := (wtm.Client{Runner: f, Bin: "wtm"}).List("/repo"); err == nil {
		t.Fatal("want parse error")
	}
}

func TestRunIsInteractive(t *testing.T) {
	f := &execx.Fake{}
	if err := (wtm.Client{Runner: f, Bin: "/opt/wtm"}).Run(wtm.RunParams{Repo: "/repo", Args: []string{"clean", "feat/a"}}); err != nil {
		t.Fatal(err)
	}
	if f.Calls[0].Line() != "/opt/wtm clean feat/a" || !f.Calls[0].Interactive || f.Calls[0].Dir != "/repo" {
		t.Fatalf("call %+v", f.Calls[0])
	}
}

func TestResolveTrimsStdout(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte("/wt/a b\n"), nil }}
	got, err := wtm.Client{Runner: f, Bin: "wtm"}.Resolve("/repo")
	if err != nil || got != "/wt/a b" {
		t.Fatalf("got %q, %v", got, err)
	}
	if f.Calls[0].Line() != "wtm resolve" || !f.Calls[0].Interactive {
		t.Fatalf("call %+v", f.Calls[0])
	}
}

func TestResolveAbortedIsEmpty(t *testing.T) {
	f := &execx.Fake{}
	got, err := wtm.Client{Runner: f, Bin: "wtm"}.Resolve("/repo")
	if err != nil || got != "" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestContracts(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte(`{"version":"0.29.0","events":1}`), nil }}
	got, err := wtm.Client{Runner: f, Bin: "wtm"}.Contracts()
	if err != nil || got != (domain.Contracts{Version: "0.29.0", Events: 1}) {
		t.Fatalf("got %+v err %v", got, err)
	}
	if f.Lines()[0] != "wtm version --output json" {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestContractsKeepsExitCode(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return nil, execx.ExitError{Code: 2} }}
	_, err := wtm.Client{Runner: f, Bin: "wtm"}.Contracts()
	if code, ok := execx.ExitCode(err); !ok || code != 2 {
		t.Fatalf("err %v", err)
	}
}

func TestRunTagsCorrelation(t *testing.T) {
	f := &execx.Fake{}
	err := wtm.Client{Runner: f, Bin: "wtm"}.Run(wtm.RunParams{Repo: "/nx/app", Args: []string{"create"}, CorrelationID: "herdr-wtm:x"})
	if err != nil {
		t.Fatal(err)
	}
	c := f.Calls[0]
	if !c.Interactive || c.Dir != "/nx/app" || c.Line() != "wtm create" || !slices.Equal(c.Env, []string{"WTM_CORRELATION_ID=herdr-wtm:x"}) {
		t.Fatalf("call %+v", c)
	}
}

func TestEventsGlobalRunsOutsideAnyRepo(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) {
		return []byte("{\"v\":1,\"type\":\"ready\"}\nnot json\n{\"v\":1,\"type\":\"worktree.created\",\"correlation_id\":\"herdr-wtm:1\",\"worktree\":{\"branch\":\"a\",\"path\":\"/x/a\"}}\n"), nil
	}}
	var got []domain.Event
	err := wtm.Client{Runner: f, Bin: "wtm"}.Events(context.Background(), wtm.EventsParams{OnEvent: func(ev domain.Event) { got = append(got, ev) }})
	if err != nil {
		t.Fatal(err)
	}
	if c := f.Calls[0]; c.Dir != "/" || c.Line() != "wtm events --output json" || !c.Streamed {
		t.Fatalf("call %+v", c)
	}
	if len(got) != 2 || got[1].CorrelationID != "herdr-wtm:1" || got[1].Worktree.Path != "/x/a" {
		t.Fatalf("events %+v", got)
	}
}

func TestEventsOfOneRepo(t *testing.T) {
	f := &execx.Fake{}
	_ = wtm.Client{Runner: f, Bin: "wtm"}.Events(context.Background(), wtm.EventsParams{Repo: "/nx/app", OnEvent: func(domain.Event) {}})
	if c := f.Calls[0]; c.Dir != "/nx/app" || c.Line() != "wtm events --output json --repo /nx/app" {
		t.Fatalf("call %+v", c)
	}
}
