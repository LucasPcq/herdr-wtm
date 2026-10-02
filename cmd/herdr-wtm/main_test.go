package main

import (
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

func deps(f *execx.Fake) app.Deps {
	return app.Deps{
		Wtm: wtm.Client{Runner: f, Bin: "wtm"}, Herdr: herdr.Client{Runner: f, Bin: "herdr"}, Git: f,
		Config: config.Default(), PluginID: "lucaspcq.wtm", Out: io.Discard, In: strings.NewReader("\n"),
		Exists: func(string) bool { return false }, Log: log.New(io.Discard, "", 0),
	}
}

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestDispatchLaunchNotifiesOnError(t *testing.T) {
	f := &execx.Fake{}
	err := dispatch(deps(f), []string{"launch", "create"}, env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": "{}"}))
	if err == nil {
		t.Fatal("want error")
	}
	if !slices.ContainsFunc(f.Lines(), func(l string) bool { return strings.HasPrefix(l, "herdr notification show wtm --body ") }) {
		t.Fatalf("no notification: %v", f.Lines())
	}
}

func TestDispatchLaunchOpensPopup(t *testing.T) {
	f := &execx.Fake{}
	ctx := `{"worktree":{"checkout_path":"/nx/app","repo_root":"/nx/app","is_linked_worktree":false}}`
	if err := dispatch(deps(f), []string{"launch", "ui"}, env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": ctx})); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(f.Lines()[0], "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestDispatchRunReadsEnv(t *testing.T) {
	f := &execx.Fake{Handler: func(c execx.Call) ([]byte, error) {
		if c.Line() == "herdr workspace list" {
			return []byte(`{"result":{"workspaces":[]}}`), nil
		}
		return []byte("[]"), nil
	}}
	e := env(map[string]string{app.EnvCmd: "prune", app.EnvRepo: "/nx/app"})
	if err := dispatch(deps(f), []string{"run"}, e); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(f.Lines(), "wtm prune") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestDispatchSyncAllDoesNotNotifyErrors(t *testing.T) {
	f := &execx.Fake{Handler: func(c execx.Call) ([]byte, error) { return nil, io.ErrUnexpectedEOF }}
	if err := dispatch(deps(f), []string{"sync", "--all"}, env(nil)); err == nil {
		t.Fatal("want error")
	}
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, "herdr notification") {
			t.Fatalf("startup sync must not notify errors: %v", f.Lines())
		}
	}
}

func TestDispatchUsage(t *testing.T) {
	f := &execx.Fake{}
	for _, args := range [][]string{nil, {"launch"}, {"bogus"}} {
		if err := dispatch(deps(f), args, env(nil)); err == nil {
			t.Fatalf("%v: want error", args)
		}
	}
}
