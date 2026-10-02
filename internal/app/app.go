// Package app orchestrates wtm and herdr for each plugin entrypoint.
package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Environment passed from `launch` to the popup's `run`.
const (
	EnvCmd    = "HERDR_WTM_CMD"
	EnvRepo   = "HERDR_WTM_REPO"
	EnvOrigin = "HERDR_WTM_ORIGIN"
)

// Commands are the wtm commands the plugin exposes, in manifest order.
var Commands = []string{"create", "checkout", "open", "clean", "prune", "ui"}

func IsCommand(cmd string) bool { return slices.Contains(Commands, cmd) }

// Deps carries every collaborator; main wires real ones, tests wire fakes.
type Deps struct {
	Wtm      wtm.Client
	Herdr    herdr.Client
	Git      execx.Runner
	Config   config.Config
	PluginID string
	Out      io.Writer
	In       io.Reader
	Exists   func(string) bool
	Log      *log.Logger
}

// fail shows err in the popup and waits for Enter so the user can read it.
func (d Deps) fail(err error) error {
	fmt.Fprintf(d.Out, "\nherdr-wtm: %v\n\nPress Enter to close.", err)
	_, _ = bufio.NewReader(d.In).ReadString('\n')
	return err
}

// apply closes then opens per plan, focusing the last opened workspace, and
// notifies a summary. An empty plan with no failures notifies nothing.
func (d Deps) apply(repo string, plan reconcile.Plan) {
	var opened, failures []string
	closed := 0
	for _, id := range plan.Close {
		if err := d.Herdr.Close(id); err != nil {
			failures = append(failures, fmt.Sprintf("close %s: %v", id, err))
			continue
		}
		closed++
	}
	for i, path := range plan.Open {
		focus := d.Config.FocusOnOpen && i == len(plan.Open)-1
		if _, err := d.Herdr.OpenWorktree(repo, path, focus); err != nil {
			failures = append(failures, fmt.Sprintf("open %s: %v", filepath.Base(path), err))
			continue
		}
		opened = append(opened, filepath.Base(path))
	}
	for _, f := range failures {
		d.Log.Print(f)
	}
	body := summary(opened, closed, failures)
	if body == "" {
		return
	}
	if err := d.Herdr.Notify("wtm", body); err != nil {
		d.Log.Printf("notify: %v", err)
	}
}

func summary(opened []string, closed int, failures []string) string {
	var parts []string
	if len(opened) > 0 {
		parts = append(parts, "opened "+strings.Join(opened, ", "))
	}
	if closed > 0 {
		parts = append(parts, fmt.Sprintf("closed %d workspace(s)", closed))
	}
	if len(failures) > 0 {
		parts = append(parts, "failed: "+strings.Join(failures, "; "))
	}
	return strings.Join(parts, " · ")
}
