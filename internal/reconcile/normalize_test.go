package reconcile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
)

func TestNormalizeCleans(t *testing.T) {
	if got := reconcile.Normalize("/nonexistent-root/a/../b/"); got != "/nonexistent-root/b" {
		t.Fatalf("got %q", got)
	}
	if got := reconcile.Normalize(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeResolvesSymlinkedParent(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(real, "wt"), 0o755); err != nil {
		t.Fatal(err)
	}
	resolvedReal, _ := filepath.EvalSymlinks(real)
	if got := reconcile.Normalize(filepath.Join(link, "wt")); got != filepath.Join(resolvedReal, "wt") {
		t.Fatalf("existing: got %q", got)
	}
	// A deleted worktree under a symlinked parent still matches its resolved spelling.
	if got := reconcile.Normalize(filepath.Join(link, "gone", "deep")); got != filepath.Join(resolvedReal, "gone", "deep") {
		t.Fatalf("deleted: got %q", got)
	}
}
