package domain

import (
	"io/fs"
	"time"
)

const PluginID = "lucaspcq.wtm"

// MenuAction is the plugin action the menu key binding launches.
const MenuAction = PluginID + ".menu"

const (
	EnvCmd           = "HERDR_WTM_CMD"
	EnvRepo          = "HERDR_WTM_REPO"
	EnvOrigin        = "HERDR_WTM_ORIGIN"
	EnvPluginContext = "HERDR_PLUGIN_CONTEXT_JSON"
	EnvPluginState   = "HERDR_PLUGIN_STATE_DIR"
	EnvPluginConfig  = "HERDR_PLUGIN_CONFIG_DIR"
	EnvHerdrBin      = "HERDR_BIN_PATH"
	EnvHerdrConfig   = "HERDR_CONFIG_PATH"
	EnvCorrelationID = "WTM_CORRELATION_ID"
	EnvNoUpdateCheck = "WTM_NO_UPDATE_CHECK"
)

// Subcommands and flags of the herdr-wtm binary.
const (
	SubLaunch  = "launch"
	SubRun     = "run"
	SubSync    = "sync"
	SubWatch   = "watch"
	FlagDetach = "--detach"
)

// Popup commands: wtm commands run as is, and the plugin's own.
const (
	CmdCreate   = "create"
	CmdCheckout = "checkout"
	CmdOpen     = "open"
	CmdClean    = "clean"
	CmdPrune    = "prune"
	CmdUI       = "ui"
	CmdMenu     = "menu"
	CmdSync     = "sync"
	CmdBind     = "bind"
)

// WtmCommands are the wtm commands the plugin exposes, in manifest order.
var WtmCommands = []string{CmdCreate, CmdCheckout, CmdOpen, CmdClean, CmdPrune, CmdUI}

// Event types of `wtm events` the plugin acts on.
const (
	EventSnapshot    = "snapshot"
	EventReady       = "ready"
	EventCreated     = "worktree.created"
	EventProvisioned = "worktree.provisioned"
	EventRelocated   = "worktree.relocated"
	EventRemoved     = "worktree.removed"
)

// wtm exit codes the plugin tells apart (wtm docs/guide/integrations.md).
const (
	WtmExitUsage        = 2
	WtmExitSchemaTooNew = 20
)

// MinEventsVersion is the `events` contract version the plugin reads.
const MinEventsVersion = 1

// CorrelationPrefix marks the wtm commands the popup starts.
const CorrelationPrefix = "herdr-wtm:"

// GlobalStreamDir is outside any repository, where `wtm events` follows every
// repository wtm knows (LUC-248 tracks an explicit flag).
const GlobalStreamDir = "/"

const (
	DefaultWtmBin    = "wtm"
	DefaultHerdrBin  = "herdr"
	DefaultMenuKey   = "prefix+alt+w"
	ConfigFile       = "config.toml"
	LogFile          = "herdr-wtm.log"
	WatchLockFile    = "watch.lock"
	BindBackupSuffix = ".bak-herdr-wtm"
	KeybindMarker    = "# herdr-wtm plugin"
	NotifyTitle      = "wtm"
	LogPrefix        = "herdr-wtm "
)

const (
	FileMode fs.FileMode = 0o644
	DirMode  fs.FileMode = 0o755
)

const ExitCodeError = 1

const (
	WatchLivenessTick = 15 * time.Second
	StreamBackoffMin  = time.Second
	StreamBackoffMax  = 30 * time.Second
	NotifyQuietWindow = time.Second
	// SnapshotTimeout bounds Sync: `wtm events` waits forever for a daemon that cannot start.
	SnapshotTimeout       = 10 * time.Second
	WatchLivenessFailures = 3
)

// Sizes of the popups `launch` opens.
const (
	BindPopupWidth  = "72"
	BindPopupHeight = "14"
	PopupEntrypoint = "run"
)
