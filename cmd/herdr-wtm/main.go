// Command herdr-wtm is the herdr plugin binary for wtm.
//
//	herdr-wtm launch <menu|bind|create|checkout|open|clean|prune|ui>   herdr action
//	herdr-wtm run                                                     popup entrypoint
//	herdr-wtm sync                                                    herdr action
//	herdr-wtm watch [--detach]                                        startup hook
package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/app"
	"github.com/LucasPcq/herdr-wtm/internal/config"
	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/fsx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/menu"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

func main() {
	_ = os.Setenv(domain.EnvNoUpdateCheck, "1")
	logger := newLogger(os.Getenv(domain.EnvPluginState))
	herdrBin := cmp.Or(os.Getenv(domain.EnvHerdrBin), domain.DefaultHerdrBin)
	runner := execx.OS{}

	cfg, err := config.Load(os.Getenv(domain.EnvPluginConfig))
	if err != nil {
		logger.Print(err)
		_ = herdr.Client{Runner: runner, Bin: herdrBin}.Notify(err.Error())
		os.Exit(domain.ExitCodeError)
	}

	d := newDeps(depsParams{Logger: logger, Runner: runner, HerdrBin: herdrBin, Config: cfg})
	if err := dispatch(dispatchParams{Deps: d, Args: os.Args[1:], Getenv: os.Getenv}); err != nil {
		logger.Printf("%s: %v", strings.Join(os.Args[1:], " "), err)
		os.Exit(domain.ExitCodeError)
	}
}

type depsParams struct {
	Logger   *log.Logger
	Runner   execx.Runner
	HerdrBin string
	Config   config.Config
}

func newDeps(p depsParams) app.Deps {
	return app.Deps{
		Wtm:          wtm.Client{Runner: p.Runner, Bin: p.Config.WtmBin},
		Herdr:        herdr.Client{Runner: p.Runner, Bin: p.HerdrBin},
		Git:          p.Runner,
		Config:       p.Config,
		FS:           fsx.OS(),
		Out:          os.Stdout,
		In:           os.Stdin,
		Log:          p.Logger,
		Choose:       menu.Choose,
		Shield:       shieldSignals,
		StartWatcher: detachWatch,
		HerdrConfig:  herdrConfigPath(),
	}
}

type dispatchParams struct {
	Deps   app.Deps
	Args   []string
	Getenv func(string) string
}

const usage = "usage: herdr-wtm launch <cmd> | run | sync | watch [--detach]"

func dispatch(p dispatchParams) error {
	if len(p.Args) == 0 {
		return errors.New(usage)
	}
	switch p.Args[0] {
	case domain.SubLaunch:
		return launch(p)
	case domain.SubRun:
		if p.Getenv(domain.EnvCmd) == domain.CmdBind {
			return p.Deps.Bind()
		}
		return p.Deps.Run(app.RunParams{Cmd: p.Getenv(domain.EnvCmd), Repo: p.Getenv(domain.EnvRepo), Origin: p.Getenv(domain.EnvOrigin)})
	case domain.SubSync:
		return syncWorkspaces(p)
	case domain.SubWatch:
		if len(p.Args) > 1 && p.Args[1] == domain.FlagDetach {
			return detachWatch()
		}
		return runWatch(p.Deps, p.Getenv(domain.EnvPluginState))
	}
	return fmt.Errorf("unknown subcommand %q (%s)", p.Args[0], usage)
}

func launch(p dispatchParams) error {
	if len(p.Args) < 2 {
		return fmt.Errorf("usage: herdr-wtm launch <%s|%s|%s>", domain.CmdMenu, domain.CmdBind, strings.Join(domain.WtmCommands, "|"))
	}
	raw := p.Getenv(domain.EnvPluginContext)
	p.Deps.Log.Printf("launch %s context=%s", p.Args[1], raw)
	hctx, err := herdr.ParseContext(raw)
	if err == nil {
		err = p.Deps.Launch(app.LaunchParams{Cmd: p.Args[1], Context: hctx})
	}
	if err != nil {
		_ = p.Deps.Herdr.Notify(err.Error())
	}
	return err
}

func syncWorkspaces(p dispatchParams) error {
	if len(p.Args) > 1 {
		return fmt.Errorf("unexpected argument %q (%s)", p.Args[1], usage)
	}
	hctx, err := herdr.ParseContext(p.Getenv(domain.EnvPluginContext))
	if err == nil {
		err = p.Deps.Sync(hctx)
	}
	if err != nil {
		_ = p.Deps.Herdr.Notify(err.Error())
	}
	return err
}

func newLogger(stateDir string) *log.Logger {
	var w io.Writer = os.Stderr
	if stateDir != "" {
		if f, err := os.OpenFile(filepath.Join(stateDir, domain.LogFile), os.O_CREATE|os.O_APPEND|os.O_WRONLY, domain.FileMode); err == nil {
			w = io.MultiWriter(f, os.Stderr)
		}
	}
	return log.New(w, domain.LogPrefix, log.LstdFlags)
}

func herdrConfigPath() string {
	if p := os.Getenv(domain.EnvHerdrConfig); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", domain.ConfigFile)
}
