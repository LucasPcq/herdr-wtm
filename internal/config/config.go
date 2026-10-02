// Package config loads the plugin's optional config.toml.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the user-editable plugin configuration.
type Config struct {
	WtmBin      string `toml:"wtm_bin"`
	FocusOnOpen bool   `toml:"focus_on_open"`
	PopupWidth  string `toml:"popup_width"`
	PopupHeight string `toml:"popup_height"`
}

func Default() Config {
	return Config{WtmBin: "wtm", FocusOnOpen: true, PopupWidth: "90%", PopupHeight: "90%"}
}

// Load reads dir/config.toml over the defaults. A missing file is not an error.
func Load(dir string) (Config, error) {
	cfg := Default()
	if dir == "" {
		return cfg, nil
	}
	path := filepath.Join(dir, "config.toml")
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Default(), nil
		}
		return Default(), fmt.Errorf("read %s: %w", path, err)
	}
	return cfg, nil
}
