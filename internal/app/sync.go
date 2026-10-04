package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/rules"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Sync repairs the invoking workspace's repository: it makes sure the watcher
// runs and closes the workspaces a fresh snapshot no longer holds.
func (d Deps) Sync(hctx domain.HerdrContext) error {
	d.startWatcher()
	repo, err := d.resolveRepo(hctx)
	if err != nil {
		return err
	}
	return d.syncRepo(repo)
}

func (d Deps) syncFromPopup(repo string) error {
	d.startWatcher()
	if err := d.syncRepo(repo); err != nil {
		return d.fail(err)
	}
	return nil
}

func (d Deps) syncRepo(repo string) error {
	paths, err := d.snapshot(repo)
	if err != nil {
		return err
	}
	ws, err := d.Herdr.Workspaces()
	if err != nil {
		return err
	}
	stale := rules.Stale(rules.StaleParams{RepoRoot: repo, Current: paths, Workspaces: ws, FS: d.FS})
	if len(stale) == 0 {
		d.notify("workspaces already in sync")
		return nil
	}
	closed := 0
	var failures []string
	for _, id := range stale {
		if err := d.Herdr.Close(id); err != nil {
			failures = append(failures, fmt.Sprintf("close %s: %v", id, err))
			continue
		}
		closed++
	}
	d.notify(rules.Summary(rules.SummaryParams{Closed: closed, Failures: failures}))
	return nil
}

// snapshot returns repo's worktree paths from the first snapshot of its event
// stream, then ends the stream.
func (d Deps) snapshot(repo string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var paths []string
	got := false
	err := d.Wtm.Events(ctx, wtm.EventsParams{Repo: repo, OnEvent: func(ev domain.Event) {
		switch ev.Type {
		case domain.EventSnapshot:
			for _, wt := range ev.Worktrees {
				paths = append(paths, wt.Path)
			}
			got = true
		case domain.EventReady:
			cancel()
		}
	}})
	if got {
		return paths, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, errors.New("wtm events ended before its snapshot")
}

func (d Deps) startWatcher() {
	if d.StartWatcher == nil {
		return
	}
	if err := d.StartWatcher(); err != nil {
		d.Log.Printf("start watcher: %v", err)
	}
}
