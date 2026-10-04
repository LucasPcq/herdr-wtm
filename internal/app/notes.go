package app

import (
	"sync"
	"time"

	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

// notes gathers what the watcher did into one notification per burst of
// events, so `wtm create a b c` notifies once.
type notes struct {
	mu       sync.Mutex
	quiet    time.Duration
	timer    *time.Timer
	send     func(string)
	opened   []string
	closed   int
	failures []string
}

type notesParams struct {
	Quiet time.Duration
	Send  func(string)
}

func newNotes(p notesParams) *notes { return &notes{quiet: p.Quiet, send: p.Send} }

func (n *notes) Opened(name string) { n.add(func() { n.opened = append(n.opened, name) }) }
func (n *notes) Closed()            { n.add(func() { n.closed++ }) }
func (n *notes) Failed(msg string)  { n.add(func() { n.failures = append(n.failures, msg) }) }

func (n *notes) add(record func()) {
	n.mu.Lock()
	defer n.mu.Unlock()
	record()
	if n.quiet <= 0 {
		return
	}
	if n.timer != nil {
		n.timer.Stop()
	}
	n.timer = time.AfterFunc(n.quiet, n.Flush)
}

// Flush sends what was gathered, if anything.
func (n *notes) Flush() {
	n.mu.Lock()
	body := rules.Summary(rules.SummaryParams{Opened: n.opened, Closed: n.closed, Failures: n.failures})
	n.opened, n.closed, n.failures = nil, 0, nil
	n.mu.Unlock()
	if body != "" {
		n.send(body)
	}
}
