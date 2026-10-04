package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

const herdrDefaults = "[keys]\n# prefix = \"ctrl+b\"\n# workspace_picker = \"prefix+w\"\n[server]\n"

// bindWorld fakes herdr for Bind: default config, and a reload answering reload.
func bindWorld(reload string) func(execx.Call) ([]byte, error) {
	return func(c execx.Call) ([]byte, error) {
		switch c.Line() {
		case "herdr --default-config":
			return []byte(herdrDefaults), nil
		case "herdr server reload-config":
			return []byte(reload), nil
		}
		return nil, nil
	}
}

const applied = `{"id":"cli:server:reload-config","result":{"diagnostics":[],"status":"applied","type":"config_reload"}}`

func configFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "herdr", "config.toml")
	if content != "" {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestBindDefaultKeyOnEnter(t *testing.T) {
	d, f, _ := newDeps(bindWorld(applied))
	d.HerdrConfig = configFile(t, "[ui]\nsidebar_width = 30\n")
	d.In = strings.NewReader("\n")
	if err := d.Bind(); err != nil {
		t.Fatal(err)
	}
	if got := read(t, d.HerdrConfig); !strings.Contains(got, `key = "prefix+alt+w"`) || !strings.HasPrefix(got, "[ui]\nsidebar_width = 30\n") {
		t.Fatalf("config:\n%s", got)
	}
	if read(t, d.HerdrConfig+".bak-herdr-wtm") != "[ui]\nsidebar_width = 30\n" {
		t.Fatal("backup missing or wrong")
	}
	assertHas(t, f, "herdr server reload-config")
	assertHas(t, f, "herdr notification show wtm --body wtm menu bound to prefix+alt+w")
}

func TestBindRejectsTakenKeyThenAccepts(t *testing.T) {
	d, _, out := newDeps(bindWorld(applied))
	d.HerdrConfig = configFile(t, "")
	d.In = strings.NewReader("prefix+w\nprefix+m\n")
	if err := d.Bind(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "prefix+w is already used by workspace_picker") {
		t.Fatalf("output %q", out.String())
	}
	if got := read(t, d.HerdrConfig); !strings.Contains(got, `key = "prefix+m"`) || strings.Contains(got, `key = "prefix+w"`) {
		t.Fatalf("config:\n%s", got)
	}
}

func TestBindRejectsInvalidKeyThenAccepts(t *testing.T) {
	d, _, out := newDeps(bindWorld(applied))
	d.HerdrConfig = configFile(t, "")
	d.In = strings.NewReader("prefix + m\nf12\n")
	if err := d.Bind(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "not a valid key") {
		t.Fatalf("output %q", out.String())
	}
	if !strings.Contains(read(t, d.HerdrConfig), `key = "f12"`) {
		t.Fatal("f12 not bound")
	}
}

func TestBindRestoresConfigWhenHerdrRejects(t *testing.T) {
	rejected := `{"result":{"diagnostics":[{"message":"unknown key"}],"status":"rejected","type":"config_reload"}}`
	d, _, out := newDeps(bindWorld(rejected))
	original := "[ui]\nsidebar_width = 30\n"
	d.HerdrConfig = configFile(t, original)
	d.In = strings.NewReader("prefix+m\n\n")
	if err := d.Bind(); err == nil {
		t.Fatal("want error")
	}
	if read(t, d.HerdrConfig) != original {
		t.Fatalf("config not restored:\n%s", read(t, d.HerdrConfig))
	}
	if !strings.Contains(out.String(), "unknown key") {
		t.Fatalf("output %q", out.String())
	}
}

func TestBindRemovesCreatedConfigWhenHerdrRejects(t *testing.T) {
	d, _, _ := newDeps(bindWorld(`{"result":{"diagnostics":[{"message":"bad"}],"status":"rejected"}}`))
	d.HerdrConfig = configFile(t, "")
	d.In = strings.NewReader("prefix+m\n\n")
	if err := d.Bind(); err == nil {
		t.Fatal("want error")
	}
	if _, err := os.Stat(d.HerdrConfig); !os.IsNotExist(err) {
		t.Fatalf("config created by a rejected bind was left behind: %v", err)
	}
}

func TestBindCancelledWritesNothing(t *testing.T) {
	for name, input := range map[string]string{"eof": "", "esc": "\x1b\n"} {
		d, f, _ := newDeps(bindWorld(applied))
		d.HerdrConfig = configFile(t, "[ui]\n")
		d.In = strings.NewReader(input)
		if err := d.Bind(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if read(t, d.HerdrConfig) != "[ui]\n" {
			t.Fatalf("%s: config changed", name)
		}
		assertNoPrefix(t, f, "herdr server reload-config")
	}
}

func TestLaunchBindNeedsNoRepository(t *testing.T) {
	d, f, _ := newDeps(nil)
	if err := d.Launch(app.LaunchParams{Cmd: domain.CmdBind}); err != nil {
		t.Fatal(err)
	}
	if len(f.Calls) != 1 || !strings.Contains(f.Calls[0].Line(), "--env HERDR_WTM_CMD=bind") || strings.Contains(f.Calls[0].Line(), "HERDR_WTM_REPO") {
		t.Fatalf("lines %v", f.Lines())
	}
}
