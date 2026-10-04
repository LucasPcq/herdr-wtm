package app_test

import (
	"errors"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func otherWS(id, path string) domain.Workspace {
	return domain.Workspace{ID: id, Worktree: &domain.WorktreeInfo{CheckoutPath: path, RepoRoot: "/nx/other", IsLinked: true}}
}

func syncWorld(otherListErr error) func(execx.Call) ([]byte, error) {
	return func(c execx.Call) ([]byte, error) {
		switch {
		case c.Line() == "herdr workspace list":
			return workspacesJSON(primaryWS(), linkedWS("w2", "/nx/app.wt/gone"), linkedWS("w3", "/nx/app.wt/alive"),
				otherWS("w4", "/nx/other.wt/gone"), domain.Workspace{ID: "w5"}), nil
		case c.Line() == "wtm list --output json" && c.Dir == repo:
			return listJSON(mainWT(), wt("alive", "/nx/app.wt/alive")), nil
		case c.Line() == "wtm list --output json" && c.Dir == "/nx/other":
			if otherListErr != nil {
				return nil, otherListErr
			}
			return listJSON(wtm0("/nx/other")), nil
		}
		return nil, nil
	}
}

func TestSyncCurrentRepoOnly(t *testing.T) {
	d, f, _ := newDeps(syncWorld(nil))
	ctx := domain.HerdrContext{Worktree: &domain.WorktreeInfo{CheckoutPath: repo, RepoRoot: repo}}
	if err := d.Sync(false, ctx); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w3")
	assertNoPrefix(t, f, "herdr workspace close w4")
	assertNoPrefix(t, f, "herdr workspace close w1")
}

func TestSyncAllCoversEveryRepo(t *testing.T) {
	d, f, _ := newDeps(syncWorld(nil))
	if err := d.Sync(true, domain.HerdrContext{}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertHas(t, f, "herdr workspace close w4")
	assertNoPrefix(t, f, "herdr workspace close w3")
}

func TestSyncAllSkipsRepoWhenListFails(t *testing.T) {
	d, f, _ := newDeps(syncWorld(errors.New("wtm: not initialized")))
	if err := d.Sync(true, domain.HerdrContext{}); err != nil {
		t.Fatal(err)
	}
	assertHas(t, f, "herdr workspace close w2")
	assertNoPrefix(t, f, "herdr workspace close w4")
}

func TestSyncKeepsWorktreesStillOnDisk(t *testing.T) {
	d, f, _ := newDeps(syncWorld(nil))
	d.Exists = func(p string) bool { return p == "/nx/app.wt/gone" }
	ctx := domain.HerdrContext{Worktree: &domain.WorktreeInfo{RepoRoot: repo}}
	if err := d.Sync(false, ctx); err != nil {
		t.Fatal(err)
	}
	assertNoPrefix(t, f, "herdr workspace close")
}

func TestSyncFailsWhenHerdrUnavailable(t *testing.T) {
	d, f, _ := newDeps(func(c execx.Call) ([]byte, error) {
		if c.Line() == "herdr workspace list" {
			return nil, errors.New("socket closed")
		}
		return nil, nil
	})
	if err := d.Sync(true, domain.HerdrContext{}); err == nil {
		t.Fatal("want error")
	}
	assertNoPrefix(t, f, "wtm")
}
