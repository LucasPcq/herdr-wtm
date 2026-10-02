package wtm_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

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
	want := []wtm.Worktree{
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
	if err := (wtm.Client{Runner: f, Bin: "/opt/wtm"}).Run("/repo", "clean", "feat/a"); err != nil {
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
