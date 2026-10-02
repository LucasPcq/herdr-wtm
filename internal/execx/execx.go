// Package execx runs external processes behind an interface so callers can be
// tested with Fake.
package execx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Runner runs external commands.
type Runner interface {
	// Output runs name in dir and returns its stdout. A failure's error carries stderr.
	Output(dir, name string, args ...string) ([]byte, error)
	// Interactive runs name in dir attached to the terminal (stdin, stderr) and
	// writes its stdout to stdout, or os.Stdout when nil.
	Interactive(dir string, stdout io.Writer, name string, args ...string) error
}

// OS runs real processes.
type OS struct{}

func (OS) Output(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func (OS) Interactive(dir string, stdout io.Writer, name string, args ...string) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
