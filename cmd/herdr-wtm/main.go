// Command herdr-wtm is the herdr plugin binary for wtm.
//
//	herdr-wtm launch <menu|bind|create|checkout|open|clean|prune|ui>   (herdr action)
//	herdr-wtm run                                           (popup entrypoint)
//	herdr-wtm sync [--all]                                  (action / startup)
package main

import (
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

	herdrBin := os.Getenv(domain.EnvHerdrBin)
	if herdrBin == "" {
		herdrBin = domain.DefaultHerdrBin
	}
	runner := execx.OS{}
	hc := herdr.Client{Runner: runner, Bin: herdrBin}

	cfg, err := config.Load(os.Getenv(domain.EnvPluginConfig))
	if err != nil {
		logger.Print(err)
		_ = hc.Notify(err.Error())
		os.Exit(domain.ExitCodeError)
	}

	d := newDeps(logger, runner, herdrBin, cfg)
	if err := dispatch(d, os.Args[1:], os.Getenv); err != nil {
		logger.Printf("%s: %v", strings.Join(os.Args[1:], " "), err)
		os.Exit(domain.ExitCodeError)
	}
}

// newDeps wires the real collaborators.
func newDeps(logger *log.Logger, runner execx.Runner, herdrBin string, cfg config.Config) app.Deps {
	return app.Deps{
		Wtm:    wtm.Client{Runner: runner, Bin: cfg.WtmBin},
		Herdr:  herdr.Client{Runner: runner, Bin: herdrBin},
		Git:    runner,
		Config: cfg,
		Out:    os.Stdout,
		In:     os.Stdin,
		FS:     fsx.OS(),
		Log:    logger,
		Choose: menu.Choose,
		Shield: shieldSignals,
		// HERDR_CONFIG_PATH overrides herdr's config location, as for herdr itself.
		HerdrConfig: herdrConfigPath(),
	}
}

func dispatch(d app.Deps, args []string, getenv func(string) string) error {
	if len(args) == 0 {
		return errors.New("usage: herdr-wtm launch <cmd> | run | sync [--all]")
	}
	switch args[0] {
	case "launch":
		if len(args) < 2 {
			return fmt.Errorf("usage: herdr-wtm launch <%s>", strings.Join(domain.WtmCommands, "|"))
		}
		raw := getenv(domain.EnvPluginContext)
		d.Log.Printf("launch %s context=%s", args[1], raw)
		ctx, err := herdr.ParseContext(raw)
		if err == nil {
			err = d.Launch(args[1], ctx)
		}
		if err != nil {
			_ = d.Herdr.Notify(err.Error())
		}
		return err
	case "run":
		if getenv(domain.EnvCmd) == domain.CmdBind {
			return d.Bind()
		}
		return d.Run(getenv(domain.EnvCmd), getenv(domain.EnvRepo), getenv(domain.EnvOrigin))
	case "sync":
		all := len(args) > 1 && args[1] == "--all"
		ctx, err := herdr.ParseContext(getenv(domain.EnvPluginContext))
		if err == nil {
			err = d.Sync(all, ctx)
		}
		if err != nil && !all {
			_ = d.Herdr.Notify(err.Error())
		}
		return err
	case "watch":
		if len(args) > 1 && args[1] == "--detach" {
			return detachWatch()
		}
		return runWatch(d, getenv(domain.EnvPluginState))
	}
	return fmt.Errorf("unknown subcommand %q", args[0])
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
