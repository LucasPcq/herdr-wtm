// Package app orchestrates wtm and herdr for each plugin entrypoint.
package app

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

// Chooser shows the action menu and returns the chosen command, or "" when cancelled.
type Chooser func(p menu.ChooseParams) (string, error)

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
	// Shield, when set, keeps SIGINT/SIGHUP off the popup while wtm runs, so
	// closing the popup cannot cut a command half-way; it returns the lift.
	Shield func() func()
	// HerdrConfig is the path of herdr's config.toml, edited by Bind.
	HerdrConfig string
	// StartWatcher makes sure a watcher runs; nil does nothing.
	StartWatcher func() error
}

// fail shows err in the popup and waits for Enter so the user can read it.
func (d Deps) fail(err error) error {
	fmt.Fprintf(d.Out, "\nherdr-wtm: %v\n\nPress Enter to close.", err)
	_, _ = bufio.NewReader(d.In).ReadString('\n')
	return err
}

// notify shows body as a herdr notification; a failure is only logged.
func (d Deps) notify(body string) {
	if err := d.Herdr.Notify(body); err != nil {
		d.Log.Printf("notify: %v", err)
	}
}
