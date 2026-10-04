package main

import (
	"io"
	"log"
	"os"
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
	err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"launch", "create"}, Getenv: env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": "{}"})})
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
	if err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"launch", "ui"}, Getenv: env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": ctx})}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(f.Lines()[0], "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestDispatchRunReadsEnv(t *testing.T) {
	f := &execx.Fake{}
	e := env(map[string]string{domain.EnvCmd: "prune", domain.EnvRepo: "/nx/app"})
	if err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"run"}, Getenv: e}); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(f.Lines(), "wtm prune") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestDispatchUsage(t *testing.T) {
	f := &execx.Fake{}
	for _, args := range [][]string{nil, {"launch"}, {"bogus"}} {
		if err := dispatch(dispatchParams{Deps: deps(f), Args: args, Getenv: env(nil)}); err == nil {
			t.Fatalf("%v: want error", args)
		}
	}
}

func TestDispatchLaunchMenu(t *testing.T) {
	f := &execx.Fake{}
	ctx := `{"worktree":{"checkout_path":"/nx/app","repo_root":"/nx/app","is_linked_worktree":false}}`
	if err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"launch", "menu"}, Getenv: env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": ctx})}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.Lines()[0], "--env HERDR_WTM_CMD=menu") {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestMainWiresMenuChooser(t *testing.T) {
	if newDeps(depsParams{Runner: execx.OS{}, HerdrBin: "herdr", Config: config.Default()}).Choose == nil {
		t.Fatal("Choose not wired")
	}
}

func TestMainWiresSignalShield(t *testing.T) {
	if newDeps(depsParams{Runner: execx.OS{}, HerdrBin: "herdr", Config: config.Default()}).Shield == nil {
		t.Fatal("Shield not wired")
	}
}

func TestDispatchRunBindAsksForKey(t *testing.T) {
	f := &execx.Fake{}
	d := deps(f)
	d.HerdrConfig = t.TempDir() + "/config.toml"
	d.In = strings.NewReader("") // cancelled at the prompt
	if err := dispatch(dispatchParams{Deps: d, Args: []string{"run"}, Getenv: env(map[string]string{domain.EnvCmd: "bind"})}); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(f.Lines(), "herdr --default-config") || slices.ContainsFunc(f.Lines(), func(l string) bool { return strings.HasPrefix(l, "wtm") }) {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestDispatchSyncNotifiesErrors(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return nil, execx.ExitError{Code: 12} }}
	ctx := `{"worktree":{"checkout_path":"/nx/app","repo_root":"/nx/app","is_linked_worktree":false}}`
	err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"sync"}, Getenv: env(map[string]string{"HERDR_PLUGIN_CONTEXT_JSON": ctx})})
	if err == nil {
		t.Fatal("want error")
	}
	if !slices.ContainsFunc(f.Lines(), func(l string) bool { return strings.HasPrefix(l, "herdr notification show wtm --body ") }) {
		t.Fatalf("no notification: %v", f.Lines())
	}
}

func TestDispatchRefusesSyncAll(t *testing.T) {
	f := &execx.Fake{}
	if err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"sync", "--all"}, Getenv: env(nil)}); err == nil {
		t.Fatal("--all is gone in 0.2.0")
	}
}

func TestMainWiresWatcherStart(t *testing.T) {
	if newDeps(depsParams{Runner: execx.OS{}, HerdrBin: "herdr", Config: config.Default()}).StartWatcher == nil {
		t.Fatal("StartWatcher not wired")
	}
}

func TestManifestHasNoPOCAction(t *testing.T) {
	data, err := os.ReadFile("../../herdr-plugin.toml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `id = "watch"`) || !strings.Contains(string(data), `version = "0.2.0"`) {
		t.Fatal("manifest must be 0.2.0 without the POC watch action")
	}
}

func TestDispatchPopupOpensTheCommandPopup(t *testing.T) {
	f := &execx.Fake{}
	e := env(map[string]string{domain.EnvCmd: domain.CmdCreate, domain.EnvRepo: "/nx/app", domain.EnvOrigin: "/nx/app.wt/a"})
	if err := dispatch(dispatchParams{Deps: deps(f), Args: []string{"popup"}, Getenv: e}); err != nil {
		t.Fatal(err)
	}
	want := "herdr plugin pane open --plugin lucaspcq.wtm --entrypoint run --placement popup --width 100 --height 30 --env HERDR_WTM_CMD=create --env HERDR_WTM_ORIGIN=/nx/app.wt/a --env HERDR_WTM_REPO=/nx/app"
	if !slices.Contains(f.Lines(), want) {
		t.Fatalf("lines %v", f.Lines())
	}
}

func TestMainWiresRelaunch(t *testing.T) {
	if newDeps(depsParams{Runner: execx.OS{}, HerdrBin: "herdr", Config: config.Default()}).Relaunch == nil {
		t.Fatal("Relaunch not wired")
	}
}
