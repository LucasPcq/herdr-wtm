# herdr-wtm

A [herdr](https://herdr.dev) plugin for [wtm](https://github.com/LucasPcq/wtm): run wtm from herdr and keep
your workspace bar in sync with your worktrees.

One key opens a popup menu of wtm actions — keyboard or mouse:

- **New worktree / checkout a PR / dashboard / prune / clean:** the chosen wtm command runs in the popup,
  wizard included. When it exits, worktrees it created open as workspaces and workspaces of worktrees it
  removed close.
- **Open a worktree:** wtm's picker, then herdr focuses the workspace (or opens it).
- **Sync workspaces:** closes workspaces whose worktree no longer exists (cleaned from another shell, by an
  agent…). Runs at herdr startup too. It never opens anything.

## Install

```bash
herdr plugin install LucasPcq/herdr-wtm
```

Requires `wtm`. Installing downloads the prebuilt binary of the manifest's version from the GitHub release
(`curl` or `wget`, checksum-verified) for macOS and Linux on amd64/arm64. If no release binary fits, it builds
from source when Go is installed. Pin a version with `--ref v0.1.0`.

## Actions

| Action | Does |
|---|---|
| `lucaspcq.wtm.menu` | the action menu (bind this one) |
| `lucaspcq.wtm.create` | `wtm create` |
| `lucaspcq.wtm.checkout` | `wtm checkout` |
| `lucaspcq.wtm.open` | `wtm resolve` picker → focus/open the workspace |
| `lucaspcq.wtm.clean` | `wtm clean <this workspace's branch>` (picker from the main checkout) |
| `lucaspcq.wtm.prune` | `wtm prune` |
| `lucaspcq.wtm.ui` | `wtm ui` |
| `lucaspcq.wtm.sync` | close workspaces whose worktree is gone |

Actions run against the repository of the workspace you invoke them from.

## Key bindings

Bind the menu in `~/.config/herdr/config.toml`, then `herdr server reload-config`:

```toml
[[keys.command]]
key = "prefix+alt+w"
type = "plugin_action"
command = "lucaspcq.wtm.menu"
description = "wtm menu"
```

In the menu: ↑↓ or j/k, Enter, 1-7 to run directly, Esc to close; or click an entry. The individual actions
can be bound the same way if you want a direct key for one of them. `alt` combinations depend on your
terminal; pick another key if it does not reach herdr.

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
HERDR_WTM_BUILD_FROM_SOURCE=1 sh scripts/install.sh   # `plugin link` does not run [[build]]
herdr plugin link "$PWD"
```

## Releasing

1. Set `version` in `herdr-plugin.toml` (e.g. `0.2.0`) and commit it on `main`.
2. Tag and push: `git tag v0.2.0 && git push origin v0.2.0`.

The Release workflow runs the tests, checks the tag matches the manifest version, and publishes the binaries
with GoReleaser. A tag like `v0.2.0-beta.1` becomes a pre-release.
