package main

import (
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

func deps(f *execx.Fake) app.Deps {
	return app.Deps{
		Wtm: wtm.Client{Runner: f, Bin: "wtm"}, Herdr: herdr.Client{Runner: f, Bin: "herdr"}, Git: f,
		Config: config.Default(), Out: io.Discard, In: strings.NewReader("\n"),
		FS: domain.FS{Normalize: filepath.Clean, Exists: func(string) bool { return false }}, Log: log.New(io.Discard, "", 0),
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
	e := env(map[string]string{domain.EnvCmd: "prune", domain.EnvRepo: "/nx/app"})
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

func TestDispatchLaunchMenu(t *testing.T) {
	f := &execx.Fake{}
	ctx := `{"worktree":{"checkout_path":"/nx/app","repo_root":"/nx/app","is_linked_worktree":false}}`
	if err := dispatch(deps(f), []string{"launch", "menu"}, env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": ctx})); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.Lines()[0], "--env HERDR_WTM_CMD=menu") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestMainWiresMenuChooser(t *testing.T) {
	if newDeps(nil, execx.OS{}, "herdr", config.Default()).Choose == nil {
		t.Fatal("Choose not wired")
	}
}

func TestMainWiresSignalShield(t *testing.T) {
	if newDeps(nil, execx.OS{}, "herdr", config.Default()).Shield == nil {
		t.Fatal("Shield not wired")
	}
}

func TestDispatchRunBindAsksForKey(t *testing.T) {
	f := &execx.Fake{}
	d := deps(f)
	d.HerdrConfig = t.TempDir() + "/config.toml"
	d.In = strings.NewReader("") // cancelled at the prompt
	if err := dispatch(d, []string{"run"}, env(map[string]string{domain.EnvCmd: "bind"})); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(f.Lines(), "herdr --default-config") || slices.ContainsFunc(f.Lines(), func(l string) bool { return strings.HasPrefix(l, "wtm") }) {
		t.Fatalf("lines %v", f.Lines())
	}
}
