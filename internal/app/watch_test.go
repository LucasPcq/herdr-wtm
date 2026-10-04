package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// withVersion answers `wtm version --output json` with body and err, and
// every other call with next (nil: empty output).
func withVersion(body string, err error, next func(execx.Call) ([]byte, error)) func(execx.Call) ([]byte, error) {
	return func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm version --output json" {
			return []byte(body), err
		}
		if next == nil {
			return nil, nil
		}
		return next(c)
	}
}

const wtm029 = `{"version":"0.29.0","events":1}`

func TestWatchRefusesWtmWithoutVersionCommand(t *testing.T) {
	d, f, _ := newDeps(withVersion("", execx.ExitError{Code: 2}, nil))
	if err := d.Watch(context.Background()); !errors.Is(err, domain.ErrWtmTooOld) {
		t.Fatalf("err %v", err)
	}
	assertHas(t, f, "herdr notification show wtm --body "+domain.ErrWtmTooOld.Error())
	assertNoPrefix(t, f, "wtm events")
}

func TestWatchRefusesWtmWithoutEventsContract(t *testing.T) {
	d, _, _ := newDeps(withVersion(`{"version":"0.28.0"}`, nil, nil))
	if err := d.Watch(context.Background()); !errors.Is(err, domain.ErrWtmTooOld) {
		t.Fatalf("err %v", err)
	}
}

func TestWatchStopsOnNewerSchema(t *testing.T) {
	d, f, _ := newDeps(withVersion(wtm029, nil, func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm events --output json" {
			return eventLines(snapshotEv(), domain.Event{V: 1, Type: domain.EventReady}), execx.ExitError{Code: 20}
		}
		return herdrState([]domain.Workspace{primaryWS()})(c)
	}))
	if err := d.Watch(context.Background()); !errors.Is(err, domain.ErrWtmSchemaTooNew) {
		t.Fatalf("err %v", err)
	}
	assertHas(t, f, "herdr notification show wtm --body "+domain.ErrWtmSchemaTooNew.Error())
}

func TestWatchRestartsAfterStreamEnds(t *testing.T) {
	streams := 0
	d, f, _ := newDeps(withVersion(wtm029, nil, func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm events --output json" {
			streams++
			if streams == 1 {
				return eventLines(snapshotEv(), domain.Event{V: 1, Type: domain.EventReady}), execx.ExitError{Code: 1}
			}
			return nil, execx.ExitError{Code: 20}
		}
		return herdrState([]domain.Workspace{primaryWS()})(c)
	}))
	start := time.Now()
	_ = d.Watch(context.Background())
	if streams != 2 {
		t.Fatalf("streams %d, lines %v", streams, f.Lines())
	}
	if time.Since(start) < domain.StreamBackoffMin {
		t.Fatal("restarted without backing off")
	}
}

func TestWatchReturnsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d, _, _ := newDeps(withVersion(wtm029, nil, func(c execx.Call) ([]byte, error) {
		if strings.HasPrefix(c.Line(), "wtm events") {
			cancel()
			return nil, execx.ExitError{Code: -1}
		}
		return nil, nil
	}))
	if err := d.Watch(ctx); err != nil {
		t.Fatalf("err %v", err)
	}
}
