// Package config loads the plugin's optional config.toml.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// Config is the user-editable plugin configuration. Unknown keys, such as
// 0.1's focus_on_open, are ignored.
type Config struct {
	WtmBin      string `toml:"wtm_bin"`
	PopupWidth  string `toml:"popup_width"`
	PopupHeight string `toml:"popup_height"`
}

func Default() Config {
	return Config{WtmBin: domain.DefaultWtmBin, PopupWidth: "90%", PopupHeight: "90%"}
}

// Load reads dir/config.toml over the defaults. A missing file is not an error.
func Load(dir string) (Config, error) {
	cfg := Default()
	if dir == "" {
		return cfg, nil
	}
	path := filepath.Join(dir, domain.ConfigFile)
	_, err := toml.DecodeFile(path, &cfg)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), fmt.Errorf("read %s: %w", path, err)
	}
	if err := validate(cfg); err != nil {
		return Default(), fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

func validate(cfg Config) error {
	for _, field := range []struct{ key, value string }{
		{"wtm_bin", cfg.WtmBin}, {"popup_width", cfg.PopupWidth}, {"popup_height", cfg.PopupHeight},
	} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s must not be empty", field.key)
		}
	}
	return nil
}
