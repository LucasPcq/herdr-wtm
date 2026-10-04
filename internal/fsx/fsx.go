// Package fsx is the plugin's only reader of the filesystem's shape: path spelling and existence.
package fsx

import (
	"os"
	"path/filepath"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// Normalize returns a canonical spelling of p so wtm's and herdr's paths compare
// equal: cleaned, with symlinks resolved on the deepest ancestor that still
// exists (a removed worktree keeps matching its old resolved spelling).
func Normalize(p string) string {
	if p == "" {
		return ""
	}
	clean := filepath.Clean(p)
	dir, rest := clean, ""
	for {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(resolved, rest)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return clean
		}
		rest = filepath.Join(filepath.Base(dir), rest)
		dir = parent
	}
}

// Exists reports whether anything is at p.
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// OS is the real filesystem, for the rules that compare paths.
func OS() domain.FS { return domain.FS{Normalize: Normalize, Exists: Exists} }
