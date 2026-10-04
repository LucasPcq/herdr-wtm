// Package app orchestrates wtm and herdr for each plugin entrypoint.
package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/reconcile"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Chooser shows the action menu and returns the chosen command, or "" when cancelled.
type Chooser func(title string, items []menu.Item) (string, error)

// Deps carries every collaborator; main wires real ones, tests wire fakes.
type Deps struct {
	Wtm    wtm.Client
	Herdr  herdr.Client
	Git    execx.Runner
	Config config.Config
	Out    io.Writer
	In     io.Reader
	FS     domain.FS
	Log    *log.Logger
	Choose Chooser
	// Shield, when set, keeps SIGINT/SIGHUP from killing the process while a
	// wtm command runs, so its changes still get reconciled; it returns the
	// function that lifts the shield.
	Shield func() func()
	// HerdrConfig is the path of herdr's config.toml, edited by Bind.
	HerdrConfig string
}

// fail shows err in the popup and waits for Enter so the user can read it.
func (d Deps) fail(err error) error {
	fmt.Fprintf(d.Out, "\nherdr-wtm: %v\n\nPress Enter to close.", err)
	_, _ = bufio.NewReader(d.In).ReadString('\n')
	return err
}

// apply closes then opens per plan, focusing the last opened workspace, and
// notifies a summary. An empty plan with no failures notifies nothing. It
// returns the ids it closed and whether it opened anything.
func (d Deps) apply(repo string, plan reconcile.Plan) (closedIDs []string, openedAny bool) {
	var opened, failures []string
	for _, id := range plan.Close {
		if err := d.Herdr.Close(id); err != nil {
			failures = append(failures, fmt.Sprintf("close %s: %v", id, err))
			continue
		}
		closedIDs = append(closedIDs, id)
	}
	for i, path := range plan.Open {
		focus := d.Config.FocusOnOpen && i == len(plan.Open)-1
		if _, err := d.Herdr.OpenWorktree(herdr.OpenParams{Repo: repo, Path: path, Focus: focus}); err != nil {
			failures = append(failures, fmt.Sprintf("open %s: %v", filepath.Base(path), err))
			continue
		}
		opened = append(opened, filepath.Base(path))
	}
	for _, f := range failures {
		d.Log.Print(f)
	}
	body := summary(opened, len(closedIDs), failures)
	if body != "" {
		d.notify(body)
	}
	return closedIDs, len(opened) > 0
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

// notify shows body as a herdr notification; a failure is only logged.
func (d Deps) notify(body string) {
	if err := d.Herdr.Notify(body); err != nil {
		d.Log.Printf("notify: %v", err)
	}
}
