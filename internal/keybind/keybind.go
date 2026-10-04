// Package keybind reads and edits herdr key bindings as text, so the plugin
// can bind its menu without rewriting the rest of the user's config.
package keybind

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

var (
	defaultLine = regexp.MustCompile(`^#\s*([a-z_]+)\s*=\s*"([^"]*)"`)
	menuLine    = regexp.MustCompile(`^\s*command\s*=\s*"` + regexp.QuoteMeta(domain.MenuAction) + `"`)
	validKey    = regexp.MustCompile(`^[a-z0-9+?._-]+$`)
)

// Defaults reads herdr's built-in bindings (action → key) from the
// commented [keys] section of `herdr --default-config`.
func Defaults(defaultConfig string) map[string]string {
	defaults := map[string]string{}
	inKeys := false
	for _, line := range strings.Split(defaultConfig, "\n") {
		trimmed := strings.TrimSpace(line)
		header := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		if strings.HasPrefix(header, "[") {
			inKeys = header == "[keys]"
			continue
		}
		if !inKeys {
			continue
		}
		m := defaultLine.FindStringSubmatch(trimmed)
		if m == nil || m[2] == "" || strings.HasPrefix(m[1], "navigate_") {
			continue
		}
		defaults[m[1]] = m[2]
	}
	return defaults
}

// Used maps every key in effect to what owns it: herdr's defaults overlaid
// with the user's [keys] table, plus the user's custom commands. The wtm menu
// binding itself is left out, since binding replaces it.
func Used(defaults map[string]string, userConfig string) (map[string]string, error) {
	var cfg struct {
		Keys map[string]any `toml:"keys"`
	}
	if _, err := toml.Decode(userConfig, &cfg); err != nil {
		return nil, fmt.Errorf("parse herdr config: %w", err)
	}
	actions := map[string]string{}
	for name, key := range defaults {
		actions[name] = key
	}
	used := map[string]string{}
	for name, v := range cfg.Keys {
		switch v := v.(type) {
		case string:
			actions[name] = v
		case []map[string]any:
			if name != "command" {
				continue
			}
			for _, c := range v {
				key, _ := c["key"].(string)
				command, _ := c["command"].(string)
				if key != "" && command != domain.MenuAction {
					used[normalize(key)] = command
				}
			}
		}
	}
	for name, key := range actions {
		for _, k := range expand(key) {
			used[k] = name
		}
	}
	return used, nil
}

// expand turns "prefix+1..9" into prefix+1 … prefix+9.
func expand(key string) []string {
	key = normalize(key)
	if key == "" {
		return nil
	}
	i := strings.Index(key, "..")
	if i < 1 || i+2 >= len(key) {
		return []string{key}
	}
	lo, err1 := strconv.Atoi(key[i-1 : i])
	hi, err2 := strconv.Atoi(key[i+2:])
	if err1 != nil || err2 != nil {
		return []string{key}
	}
	var keys []string
	for n := lo; n <= hi; n++ {
		keys = append(keys, key[:i-1]+strconv.Itoa(n))
	}
	return keys
}

func normalize(key string) string { return strings.ToLower(strings.TrimSpace(key)) }

// Normalize is the spelling keys are compared in.
func Normalize(key string) string { return normalize(key) }

// ValidKey accepts herdr's key syntax (lowercase names joined by '+').
func ValidKey(key string) bool { return validKey.MatchString(key) }

// SetMenuKey returns config with any existing wtm menu binding removed and a
// new one bound to key appended. Every other line is kept as is.
func SetMenuKey(config, key string) string {
	lines := strings.Split(config, "\n")
	var kept []string
	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) != "[[keys.command]]" {
			kept = append(kept, lines[i])
			i++
			continue
		}
		end := i + 1
		for end < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[end]), "[") {
			end++
		}
		block := lines[i:end]
		if !containsMenu(block) {
			kept = append(kept, block...)
		} else if n := len(kept); n > 0 && strings.TrimSpace(kept[n-1]) == domain.KeybindMarker {
			kept = kept[:n-1]
		}
		i = end
	}
	out := strings.TrimRight(strings.Join(kept, "\n"), "\n")
	if out != "" {
		out += "\n\n"
	}
	return out + fmt.Sprintf("%s\n[[keys.command]]\nkey = %q\ntype = \"plugin_action\"\ncommand = %q\ndescription = \"wtm menu\"\n", domain.KeybindMarker, key, domain.MenuAction)
}

func containsMenu(block []string) bool {
	for _, l := range block {
		if menuLine.MatchString(l) {
			return true
		}
	}
	return false
}
