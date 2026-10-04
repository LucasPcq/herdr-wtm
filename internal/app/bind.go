package app

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/keybind"
)

// Bind asks for the key that opens the wtm menu (Enter keeps the default),
// refuses keys herdr already uses, writes the binding into herdr's config,
// reloads it, and restores the previous config if herdr rejects it.
func (d Deps) Bind() error {
	defaults, err := d.Herdr.DefaultConfig()
	if err != nil {
		return d.fail(err)
	}
	original, err := os.ReadFile(d.HerdrConfig)
	existed := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return d.fail(err)
	}
	used, err := keybind.Used(keybind.UsedParams{Defaults: keybind.Defaults(defaults), UserConfig: string(original)})
	if err != nil {
		return d.fail(err)
	}

	key, ok := d.askKey(used)
	if !ok {
		return nil
	}

	mode := domain.FileMode
	if info, err := os.Stat(d.HerdrConfig); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.MkdirAll(filepath.Dir(d.HerdrConfig), domain.DirMode); err != nil {
		return d.fail(err)
	}
	if existed {
		if err := os.WriteFile(d.HerdrConfig+domain.BindBackupSuffix, original, mode); err != nil {
			return d.fail(err)
		}
	}
	if err := os.WriteFile(d.HerdrConfig, []byte(keybind.SetMenuKey(keybind.SetMenuKeyParams{Config: string(original), Key: key})), mode); err != nil {
		return d.fail(err)
	}
	if err := d.Herdr.ReloadConfig(); err != nil {
		if existed {
			_ = os.WriteFile(d.HerdrConfig, original, mode)
		} else {
			_ = os.Remove(d.HerdrConfig)
		}
		return d.fail(fmt.Errorf("herdr rejected the binding, config restored: %w", err))
	}
	msg := "wtm menu bound to " + key
	fmt.Fprintln(d.Out, msg)
	d.notify(msg)
	return nil
}

// askKey prompts until it gets a valid, free key; ok is false when the user
// cancels (end of input or Esc).
func (d Deps) askKey(used map[string]string) (key string, ok bool) {
	in := bufio.NewReader(d.In)
	for {
		fmt.Fprintf(d.Out, "Key for the wtm menu [%s]: ", domain.DefaultMenuKey)
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			return "", false
		}
		if strings.Contains(line, "\x1b") {
			return "", false
		}
		key = keybind.Normalize(line)
		if key == "" {
			key = domain.DefaultMenuKey
		}
		if !keybind.ValidKey(key) {
			fmt.Fprintf(d.Out, "%q is not a valid key (examples: prefix+m, ctrl+alt+w, f12)\n", strings.TrimSpace(line))
			continue
		}
		if owner, taken := used[key]; taken {
			fmt.Fprintf(d.Out, "%s is already used by %s\n", key, owner)
			continue
		}
		return key, true
	}
}
