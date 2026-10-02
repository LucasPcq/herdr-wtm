// Command herdr-wtm is the herdr plugin binary for wtm.
//
//	herdr-wtm launch <create|checkout|open|clean|prune|ui>   (herdr action)
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
	"github.com/LucasPcq/herdr-wtm/internal/execx"
	"github.com/LucasPcq/herdr-wtm/internal/herdr"
	"github.com/LucasPcq/herdr-wtm/internal/wtm"
)

const pluginID = "lucaspcq.wtm"

func main() {
	_ = os.Setenv("WTM_NO_UPDATE_CHECK", "1")
	logger := newLogger(os.Getenv("HERDR_PLUGIN_STATE_DIR"))

	herdrBin := os.Getenv("HERDR_BIN_PATH")
	if herdrBin == "" {
		herdrBin = "herdr"
	}
	runner := execx.OS{}
	hc := herdr.Client{Runner: runner, Bin: herdrBin}

	cfg, err := config.Load(os.Getenv("HERDR_PLUGIN_CONFIG_DIR"))
	if err != nil {
		logger.Print(err)
		_ = hc.Notify("wtm", err.Error())
		os.Exit(1)
	}

	d := app.Deps{
		Wtm:      wtm.Client{Runner: runner, Bin: cfg.WtmBin},
		Herdr:    hc,
		Git:      runner,
		Config:   cfg,
		PluginID: pluginID,
		Out:      os.Stdout,
		In:       os.Stdin,
		Exists:   exists,
		Log:      logger,
	}
	if err := dispatch(d, os.Args[1:], os.Getenv); err != nil {
		logger.Printf("%s: %v", strings.Join(os.Args[1:], " "), err)
		os.Exit(1)
	}
}

func dispatch(d app.Deps, args []string, getenv func(string) string) error {
	if len(args) == 0 {
		return errors.New("usage: herdr-wtm launch <cmd> | run | sync [--all]")
	}
	switch args[0] {
	case "launch":
		if len(args) < 2 {
			return fmt.Errorf("usage: herdr-wtm launch <%s>", strings.Join(app.Commands, "|"))
		}
		raw := getenv("HERDR_PLUGIN_CONTEXT_JSON")
		d.Log.Printf("launch %s context=%s", args[1], raw)
		ctx, err := herdr.ParseContext(raw)
		if err == nil {
			err = d.Launch(args[1], ctx)
		}
		if err != nil {
			_ = d.Herdr.Notify("wtm", err.Error())
		}
		return err
	case "run":
		defer shieldSignals()()
		return d.Run(getenv(app.EnvCmd), getenv(app.EnvRepo), getenv(app.EnvOrigin))
	case "sync":
		all := len(args) > 1 && args[1] == "--all"
		ctx, err := herdr.ParseContext(getenv("HERDR_PLUGIN_CONTEXT_JSON"))
		if err == nil {
			err = d.Sync(all, ctx)
		}
		if err != nil && !all {
			_ = d.Herdr.Notify("wtm", err.Error())
		}
		return err
	}
	return fmt.Errorf("unknown subcommand %q", args[0])
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func newLogger(stateDir string) *log.Logger {
	var w io.Writer = os.Stderr
	if stateDir != "" {
		if f, err := os.OpenFile(filepath.Join(stateDir, "herdr-wtm.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			w = io.MultiWriter(f, os.Stderr)
		}
	}
	return log.New(w, "herdr-wtm ", log.LstdFlags)
}
