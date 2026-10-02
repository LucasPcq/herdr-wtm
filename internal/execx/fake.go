package execx

import (
	"io"
	"strings"
)

// Call is one recorded invocation on Fake.
type Call struct {
	Dir         string
	Name        string
	Args        []string
	Interactive bool
}

// Line renders the call as "name arg1 arg2".
func (c Call) Line() string {
	return strings.TrimSpace(c.Name + " " + strings.Join(c.Args, " "))
}

// Fake records calls and answers them with Handler (nil handler: empty output, no error).
type Fake struct {
	Calls   []Call
	Handler func(Call) ([]byte, error)
}

func (f *Fake) Output(dir, name string, args ...string) ([]byte, error) {
	return f.answer(Call{Dir: dir, Name: name, Args: args})
}

func (f *Fake) Interactive(dir string, stdout io.Writer, name string, args ...string) error {
	out, err := f.answer(Call{Dir: dir, Name: name, Args: args, Interactive: true})
	if stdout != nil && len(out) > 0 {
		_, _ = stdout.Write(out)
	}
	return err
}

// Lines returns every recorded call as Call.Line.
func (f *Fake) Lines() []string {
	lines := make([]string, len(f.Calls))
	for i, c := range f.Calls {
		lines[i] = c.Line()
	}
	return lines
}

func (f *Fake) answer(c Call) ([]byte, error) {
	f.Calls = append(f.Calls, c)
	if f.Handler == nil {
		return nil, nil
	}
	return f.Handler(c)
}
