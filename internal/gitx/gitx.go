// Package gitx answers the few git questions the plugin needs.
package gitx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// RepoRoot returns the main checkout of the repository containing cwd, even
// when cwd is inside a linked worktree.
func RepoRoot(r execx.Runner, cwd string) (string, error) {
	out, err := r.Output(cwd, "git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a git repository: %w", cwd, err)
	}
	common := strings.TrimSpace(string(out))
	if filepath.Base(common) == ".git" {
		return filepath.Dir(common), nil
	}
	return common, nil
}
