package execx

import (
	"bytes"
	"context"
	"sync"
)

// Call is one recorded invocation on Fake.
type Call struct {
	Cmd
	Interactive bool
	Streamed    bool
}

// Fake records calls and answers them with Handler (nil: empty output, no error).
// Stream feeds the answer to onLine line by line, then returns its error.
type Fake struct {
	mu      sync.Mutex
	Calls   []Call
	Handler func(Call) ([]byte, error)
}

func (f *Fake) Output(c Cmd) ([]byte, error) { return f.answer(Call{Cmd: c}) }

func (f *Fake) Interactive(c Cmd) error {
	out, err := f.answer(Call{Cmd: c, Interactive: true})
	if c.Stdout != nil && len(out) > 0 {
		_, _ = c.Stdout.Write(out)
	}
	return err
}

func (f *Fake) Stream(_ context.Context, c Cmd, onLine func([]byte)) error {
	out, err := f.answer(Call{Cmd: c, Streamed: true})
	for line := range bytes.Lines(out) {
		if trimmed := bytes.TrimRight(line, "\n"); len(trimmed) > 0 {
			onLine(trimmed)
		}
	}
	return err
}

func (f *Fake) Lines() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	lines := make([]string, len(f.Calls))
	for i, c := range f.Calls {
		lines[i] = c.Line()
	}
	return lines
}

func (f *Fake) answer(c Call) ([]byte, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, c)
	f.mu.Unlock()
	if f.Handler == nil {
		return nil, nil
	}
	return f.Handler(c)
}
