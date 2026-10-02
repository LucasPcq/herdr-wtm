package gitx_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/gitx"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.email=t@t", "-c", "user.name=t"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestRepoRootFromLinkedWorktree(t *testing.T) {
	main := filepath.Join(t.TempDir(), "app")
	git(t, filepath.Dir(main), "init", "-q", "-b", "main", main)
	git(t, main, "commit", "-q", "--allow-empty", "-m", "init")
	wt := filepath.Join(filepath.Dir(main), "app.wt", "feat a")
	git(t, main, "worktree", "add", "-q", "-b", "feat-a", wt)

	want, _ := filepath.EvalSymlinks(main)
	for _, dir := range []string{main, wt} {
		got, err := gitx.RepoRoot(execx.OS{}, dir)
		if err != nil {
			t.Fatal(err)
		}
		if resolved, _ := filepath.EvalSymlinks(got); resolved != want {
			t.Fatalf("from %s: got %s, want %s", dir, got, want)
		}
	}
}

func TestRepoRootOutsideGit(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return nil, errors.New("not a git repository") }}
	if _, err := gitx.RepoRoot(f, "/tmp/x"); err == nil {
		t.Fatal("want error")
	}
}

func TestRepoRootBare(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte("/srv/app.git\n"), nil }}
	got, err := gitx.RepoRoot(f, "/srv/app.git")
	if err != nil || got != "/srv/app.git" {
		t.Fatalf("got %q, %v", got, err)
	}
	if f.Calls[0].Line() != "git rev-parse --path-format=absolute --git-common-dir" || f.Calls[0].Dir != "/srv/app.git" {
		t.Fatalf("call %+v", f.Calls[0])
	}
}
