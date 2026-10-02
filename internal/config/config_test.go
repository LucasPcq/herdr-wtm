package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/config"
)

func TestLoadMissingFileGivesDefaults(t *testing.T) {
	cfg, err := config.Load(t.TempDir())
	if err != nil || cfg != config.Default() {
		t.Fatalf("got %+v, %v", cfg, err)
	}
	cfg, err = config.Load("")
	if err != nil || cfg != config.Default() {
		t.Fatalf("empty dir: got %+v, %v", cfg, err)
	}
}

func TestLoadPartialFileKeepsDefaults(t *testing.T) {
	dir := t.TempDir()
	content := "wtm_bin = \"/opt/bin/wtm\"\nfocus_on_open = false\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{WtmBin: "/opt/bin/wtm", FocusOnOpen: false, PopupWidth: "90%", PopupHeight: "90%"}
	if cfg != want {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadInvalidFileFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("wtm_bin = "), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(dir); err == nil {
		t.Fatal("want error")
	}
}
