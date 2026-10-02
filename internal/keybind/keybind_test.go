package keybind_test

import (
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/keybind"
)

const defaultConfig = `[keys]
# prefix = "ctrl+b"
# workspace_picker = "prefix+w"
# open_worktree = ""    # optional, unset by default
# switch_tab = "prefix+1..9"
# remote_image_paste = "ctrl+v" # only active in herdr --remote
# navigate_workspace_up = "up"
# navigate_pane_left = "h"
# [[keys.command]]
# key = "prefix+alt+g"
# type = "popup"
# [keys.indexed]
# tabs = ""
[server]
# headless_cols = 120
# [worktrees]
# directory = "~/.herdr/worktrees"
`

func TestDefaultsReadsKeysSectionOnly(t *testing.T) {
	got := keybind.Defaults(defaultConfig)
	want := map[string]string{
		"prefix":             "ctrl+b",
		"workspace_picker":   "prefix+w",
		"switch_tab":         "prefix+1..9",
		"remote_image_paste": "ctrl+v",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s: got %q want %q (all %v)", k, got[k], v, got)
		}
	}
}

func TestUsedOverlaysUserConfigAndExpandsRanges(t *testing.T) {
	user := `[keys]
workspace_picker = "prefix+a"

[[keys.command]]
key = "prefix+alt+g"
type = "popup"
command = "lazygit"

[[keys.command]]
key = "prefix+alt+w"
type = "plugin_action"
command = "lucaspcq.wtm.menu"
`
	used, err := keybind.Used(keybind.Defaults(defaultConfig), user)
	if err != nil {
		t.Fatal(err)
	}
	for key, owner := range map[string]string{"prefix+a": "workspace_picker", "prefix+3": "switch_tab", "prefix+alt+g": "lazygit", "ctrl+b": "prefix"} {
		if used[key] != owner {
			t.Fatalf("%s: got %q want %q (all %v)", key, used[key], owner, used)
		}
	}
	for _, free := range []string{"prefix+w", "prefix+alt+w"} {
		if o, ok := used[free]; ok {
			t.Fatalf("%s should be free, owned by %q", free, o)
		}
	}
}

func TestUsedRejectsInvalidUserConfig(t *testing.T) {
	if _, err := keybind.Used(nil, "[keys\n"); err == nil {
		t.Fatal("want error")
	}
}

func TestSetMenuKeyAppendsAndKeepsTheRest(t *testing.T) {
	config := "[ui]\nsidebar_width = 30" // no trailing newline
	got := keybind.SetMenuKey(config, "prefix+m")
	if !strings.HasPrefix(got, config+"\n") {
		t.Fatalf("existing content changed:\n%s", got)
	}
	if !strings.HasSuffix(got, "[[keys.command]]\nkey = \"prefix+m\"\ntype = \"plugin_action\"\ncommand = \"lucaspcq.wtm.menu\"\ndescription = \"wtm menu\"\n") {
		t.Fatalf("block missing:\n%s", got)
	}
}

func TestSetMenuKeyReplacesExistingBinding(t *testing.T) {
	config := `[keys]
prefix = "ctrl+a"

# herdr-wtm plugin
[[keys.command]]
key = "prefix+alt+w"
type = "plugin_action"
command = "lucaspcq.wtm.menu"
description = "wtm menu"

[[keys.command]]
key = "prefix+alt+g"
type = "popup"
command = "lazygit"

[ui]
sidebar_width = 30
`
	got := keybind.SetMenuKey(config, "prefix+m")
	if strings.Count(got, "lucaspcq.wtm.menu") != 1 || strings.Contains(got, "prefix+alt+w") {
		t.Fatalf("old binding not replaced:\n%s", got)
	}
	for _, keep := range []string{"prefix = \"ctrl+a\"", "command = \"lazygit\"", "[ui]\nsidebar_width = 30\n"} {
		if !strings.Contains(got, keep) {
			t.Fatalf("lost %q:\n%s", keep, got)
		}
	}
	if !strings.Contains(got, "key = \"prefix+m\"") {
		t.Fatalf("new binding missing:\n%s", got)
	}
}

func TestValidKey(t *testing.T) {
	for _, k := range []string{"prefix+alt+w", "f12", "ctrl+alt+w", "prefix+shift+tab", "prefix+?"} {
		if !keybind.ValidKey(k) {
			t.Fatalf("%q should be valid", k)
		}
	}
	for _, k := range []string{"", "prefix + w", `pre"fix`, "prefix+w\nx", "a\\b"} {
		if keybind.ValidKey(k) {
			t.Fatalf("%q should be invalid", k)
		}
	}
}
