# herdr-wtm

A [herdr](https://herdr.dev) plugin for [wtm](https://github.com/LucasPcq/wtm): run wtm from herdr and keep
your workspace bar in sync with your worktrees.

- **New worktree / checkout a PR / dashboard / prune / clean:** wtm runs in a herdr popup, wizard
  included. When it exits, worktrees it created open as workspaces and workspaces of worktrees it removed
  close.
- **Open a worktree:** wtm's picker, then herdr focuses the workspace (or opens it).
- **Sync:** closes workspaces whose worktree no longer exists (cleaned from another shell, by an agent…).
  Runs at herdr startup too. It never opens anything.

## Install

Requires `wtm` and Go (the plugin is built at install time).

```bash
herdr plugin install LucasPcq/herdr-wtm
```

## Actions

| Action | Does |
|---|---|
| `lucaspcq.wtm.create` | `wtm create` |
| `lucaspcq.wtm.checkout` | `wtm checkout` |
| `lucaspcq.wtm.open` | `wtm resolve` picker → focus/open the workspace |
| `lucaspcq.wtm.clean` | `wtm clean <this workspace's branch>` (picker from the main checkout) |
| `lucaspcq.wtm.prune` | `wtm prune` |
| `lucaspcq.wtm.ui` | `wtm ui` |
| `lucaspcq.wtm.sync` | close workspaces whose worktree is gone |

Actions run against the repository of the workspace you invoke them from.

## Key bindings

The plugin binds nothing. Add what you want to `~/.config/herdr/config.toml`, then `herdr server reload-config`:

```toml
[[keys.command]]
key = "prefix+alt+c"
type = "plugin_action"
command = "lucaspcq.wtm.create"
description = "wtm: new worktree"

[[keys.command]]
key = "prefix+alt+o"
type = "plugin_action"
command = "lucaspcq.wtm.open"
description = "wtm: open a worktree"

[[keys.command]]
key = "prefix+alt+x"
type = "plugin_action"
command = "lucaspcq.wtm.clean"
description = "wtm: clean this worktree"

[[keys.command]]
key = "prefix+alt+u"
type = "plugin_action"
command = "lucaspcq.wtm.ui"
description = "wtm: dashboard"
```

`alt` combinations depend on your terminal; pick other keys if they do not reach herdr.

## Configuration

Optional, in `$(herdr plugin config-dir lucaspcq.wtm)/config.toml`:

```toml
wtm_bin = "wtm"        # path or name on PATH
focus_on_open = true   # focus the last workspace opened after a command
popup_width = "90%"
popup_height = "90%"
```

## Logs

`$HERDR_PLUGIN_STATE_DIR/herdr-wtm.log`, and `herdr plugin log list --plugin lucaspcq.wtm`.

## Development

```bash
go test ./...
go build -o bin/herdr-wtm ./cmd/herdr-wtm   # `plugin link` does not run [[build]]
herdr plugin link "$PWD"
```
