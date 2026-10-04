// Package execx runs external processes behind Runner so callers can be
// tested with Fake.
package execx

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Cmd is one process to run. Env is added to the inherited environment;
// Stdout is read by Interactive only (os.Stdout when nil).
type Cmd struct {
	Dir    string
	Name   string
	Args   []string
	Env    []string
	Stdout io.Writer
}

func (c Cmd) Line() string {
	return strings.TrimSpace(c.Name + " " + strings.Join(c.Args, " "))
}

type Runner interface {
	// Output returns stdout; a failure's error carries stderr and the exit code.
	Output(c Cmd) ([]byte, error)
	// Interactive attaches the process to the terminal.
	Interactive(c Cmd) error
	// Stream calls onLine for each stdout line until the process exits or ctx ends.
	Stream(ctx context.Context, c Cmd, onLine func([]byte)) error
}

// ExitCode returns the exit code err carries, if any.
func ExitCode(err error) (int, bool) {
	var coded interface{ ExitCode() int }
	if !errors.As(err, &coded) {
		return 0, false
	}
	return coded.ExitCode(), true
}

// ExitError is an exit status without a process, for fakes.
type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }
func (e ExitError) ExitCode() int { return e.Code }

const maxLineBytes = 4 << 20

type OS struct{}

func (OS) Output(c Cmd) ([]byte, error) {
	cmd := command(context.Background(), c)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s: %w: %s", c.Line(), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func (OS) Interactive(c Cmd) error {
	cmd := command(context.Background(), c)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if c.Stdout != nil {
		cmd.Stdout = c.Stdout
	}
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (OS) Stream(ctx context.Context, c Cmd, onLine func([]byte)) error {
	cmd := command(ctx, c)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: %w", c.Line(), err)
	}
	sc := bufio.NewScanner(out)
	sc.Buffer(make([]byte, 0, 64<<10), maxLineBytes)
	for sc.Scan() {
		onLine(sc.Bytes())
	}
	if sc.Err() != nil {
		// Nobody drains the pipe any more: without the kill, Wait never returns.
		_ = cmd.Process.Kill()
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s: %w: %s", c.Line(), err, strings.TrimSpace(stderr.String()))
	}
	return sc.Err()
}

func command(ctx context.Context, c Cmd) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	if len(c.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}
	return cmd
}
