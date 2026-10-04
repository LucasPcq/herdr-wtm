package app

import (
	"context"
	"fmt"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Watch keeps herdr's workspaces in step with the worktrees of every
// repository wtm knows, until herdr stops answering or ctx ends.
func (d Deps) Watch(ctx context.Context) error {
	if err := d.CheckWtm(); err != nil {
		d.notify(err.Error())
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go d.cancelWhenHerdrGone(ctx, cancel)
	w := NewWatcher(WatcherParams{Deps: d, Quiet: domain.NotifyQuietWindow})
	defer w.Flush()
	return d.stream(ctx, w)
}

// CheckWtm refuses a wtm that predates the event stream.
func (d Deps) CheckWtm() error {
	contracts, err := d.Wtm.Contracts()
	if code, ok := execx.ExitCode(err); ok && code == domain.WtmExitUsage {
		return domain.ErrWtmTooOld
	}
	if err != nil {
		return err
	}
	if contracts.Events < domain.MinEventsVersion {
		return domain.ErrWtmTooOld
	}
	return nil
}

// stream reads the global stream, reopening it with a backoff: a fresh stream
// opens on snapshots, so a restart loses nothing but latency.
func (d Deps) stream(ctx context.Context, w *Watcher) error {
	backoff := domain.StreamBackoffMin
	for {
		ready := false
		err := d.Wtm.Events(ctx, wtm.EventsParams{OnEvent: func(ev domain.Event) {
			ready = ready || ev.Type == domain.EventReady
			w.Handle(ev)
		}})
		if ctx.Err() != nil {
			return nil
		}
		if stop := streamStop(err); stop != nil {
			d.Log.Printf("watch: %v", err)
			d.notify(stop.Error())
			return stop
		}
		if ready {
			backoff = domain.StreamBackoffMin
		}
		d.Log.Printf("watch: stream ended (%v), retrying in %s", err, backoff)
		if !sleep(ctx, backoff) {
			return nil
		}
		backoff = min(backoff*2, domain.StreamBackoffMax)
	}
}

// streamStop returns why a stream must not be retried: no retry changes its exit code.
func streamStop(err error) error {
	code, ok := execx.ExitCode(err)
	if !ok {
		return nil
	}
	switch code {
	case domain.WtmExitSchemaTooNew:
		return domain.ErrWtmSchemaTooNew
	case domain.WtmExitUsage:
		return fmt.Errorf("wtm events refused its arguments: %w", err)
	}
	return nil
}

func (d Deps) cancelWhenHerdrGone(ctx context.Context, cancel context.CancelFunc) {
	tick := time.NewTicker(domain.WatchLivenessTick)
	defer tick.Stop()
	var alive liveness
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			_, err := d.Herdr.Workspaces()
			if alive.gone(err) {
				d.Log.Printf("watch: herdr is gone, stopping: %v", err)
				cancel()
				return
			}
		}
	}
}

// liveness tells herdr gone from herdr busy: only consecutive failures count.
type liveness struct{ failures int }

func (l *liveness) gone(err error) bool {
	if err == nil {
		l.failures = 0
		return false
	}
	l.failures++
	return l.failures >= domain.WatchLivenessFailures
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
