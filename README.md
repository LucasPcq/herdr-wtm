<h1 align="center">herdr-wtm</h1>

<p align="center">
  <strong>Your wtm worktrees, one keypress away in herdr.</strong><br>
  A <a href="https://herdr.dev">herdr</a> plugin for <a href="https://github.com/LucasPcq/wtm">wtm</a>: run worktree commands from a popup and keep your workspace bar in sync with them.
</p>

<p align="center">
  <a href="https://github.com/LucasPcq/herdr-wtm/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/LucasPcq/herdr-wtm?sort=semver"></a>
  <a href="https://github.com/LucasPcq/herdr-wtm/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/LucasPcq/herdr-wtm/actions/workflows/ci.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue"></a>
</p>

> [!WARNING]
> **Early preview.** This first version is a bridge to test how wtm and herdr work together: it drives wtm's
> own CLI from a herdr popup and mirrors the result in herdr's workspaces. A deeper, cleaner integration
> between the two is being designed. Expect changes, and please share what you would like to see in the
> [issues](https://github.com/LucasPcq/herdr-wtm/issues).

## What is wtm?

[**wtm**](https://github.com/LucasPcq/wtm) is a worktree manager built on one idea: **one branch, one worktree,
one isolated dev stack**. It is made for people who work on several branches at once, and for the agents that
do too.

`git worktree` gives each branch its own directory. wtm does everything around it:

- **A worktree is ready when it is created:** `.env` files copied, your `on_create` hooks run (`pnpm install`, …).
- **Each worktree runs its own stack:** dev servers and `docker compose` on their own ports and project names,
  side by side, without fighting over ports, containers or databases.
- **Stacked branches stay in order:** every worktree knows its parent, and `wtm sync` rebases the whole chain.
- **A dashboard for all of it:** `wtm ui` shows every worktree, the branch tree, PR status and running services.
- **Built for scripts and agents:** JSON output, `--yes` everywhere, and a skill that teaches Claude Code or
  Cursor to drive it.

New to wtm? Start with its [README](https://github.com/LucasPcq/wtm#readme) and the
[getting started guide](https://github.com/LucasPcq/wtm/blob/main/docs/guide/getting-started.md).

## Why herdr + wtm

[herdr](https://herdr.dev) organizes terminals and coding agents into workspaces, and it already knows about
git worktrees. wtm decides what a worktree *is*: provisioned, isolated, part of a stack. This plugin connects
the two, so the worktrees wtm manages show up in herdr as workspaces, without leaving the keyboard (or the
mouse):

- **One menu for wtm.** A single key opens a popup with the wtm actions. Pick one with the keyboard or a click;
  wtm runs in the popup, wizard and pickers included.
- **Your workspace bar follows.** A worktree wtm creates opens as a herdr workspace and gets the focus. A
  worktree wtm removes closes its workspace. Clean the worktree you are in, and you land back on the main
  checkout.
- **Nothing left behind.** `Sync workspaces` (also run when herdr starts) closes workspaces whose worktree was
  removed elsewhere — from another shell, by an agent. It never closes the main checkout, nor a workspace whose
  folder still exists.

## Install

You need [herdr](https://herdr.dev) 0.9 or later and [wtm](https://github.com/LucasPcq/wtm):

```bash
brew install LucasPcq/tap/wtm     # or see wtm's README for other ways
cd your-repo && wtm init          # once per repository
```

Then install the plugin:

```bash
herdr plugin install LucasPcq/herdr-wtm
```

Installation downloads the prebuilt binary for your platform (macOS and Linux, amd64 and arm64) from the
matching GitHub release and checks it against the release's checksums. No Go toolchain is needed; if no
prebuilt binary fits, it builds from source when Go is installed. Pin a version with
`herdr plugin install LucasPcq/herdr-wtm --ref v0.1.0`.

## Usage

Bind the menu in `~/.config/herdr/config.toml`, then run `herdr server reload-config`:

```toml
[[keys.command]]
key = "prefix+alt+w"
type = "plugin_action"
command = "lucaspcq.wtm.menu"
description = "wtm menu"
```

Press it from any workspace of a wtm repository:

| # | Entry | What happens |
|---|---|---|
| 1 | New worktree | `wtm create` wizard; the new worktree opens as a workspace |
| 2 | Open a worktree | wtm's picker; herdr focuses that workspace, or opens it |
| 3 | Checkout a pull request | `wtm checkout`; the PR's worktree opens as a workspace |
| 4 | Clean this worktree | `wtm clean` on the current worktree (a picker from the main checkout); its workspace closes |
| 5 | Prune finished worktrees | `wtm prune`; workspaces of removed worktrees close |
| 6 | Dashboard | `wtm ui`; what you create or delete there is mirrored when you quit |
| 7 | Sync workspaces | close workspaces whose worktree no longer exists |

In the menu: <kbd>↑</kbd>/<kbd>↓</kbd> or <kbd>j</kbd>/<kbd>k</kbd> and <kbd>Enter</kbd>, a digit to run an
entry directly, <kbd>Esc</kbd> to close — or click an entry.

Each entry also exists as its own action (`lucaspcq.wtm.create`, `.open`, `.checkout`, `.clean`, `.prune`,
`.ui`, `.sync`), to bind to a key or run with `herdr plugin action invoke`. Actions run against the repository
of the workspace you invoke them from.

`alt` combinations depend on your terminal; pick another key if <kbd>prefix</kbd>+<kbd>alt</kbd>+<kbd>w</kbd>
does not reach herdr.

## Configuration

Optional, in `$(herdr plugin config-dir lucaspcq.wtm)/config.toml`:

```toml
wtm_bin = "wtm"        # path or name on PATH
focus_on_open = true   # focus the last workspace opened after a command
popup_width = "90%"
popup_height = "90%"
```

## How it works

The plugin never changes what wtm does: it runs wtm's own commands. Around each one it reads
`wtm list --output json` before and after, then opens and closes herdr workspaces through the herdr CLI to match
what changed. A workspace is only closed when its worktree is gone from wtm and from disk.

Logs go to the plugin's state directory (`herdr-wtm.log`) and to `herdr plugin log list --plugin lucaspcq.wtm`.

## Roadmap

This preview is a starting point. Ideas being explored for a tighter integration:

- Create worktrees from herdr's own *New worktree* menu and have wtm provision them (`.env`, hooks, isolated
  ports).
- A persistent side panel listing your worktrees and their state.
- Ctrl+click a pull request URL to check it out as a worktree.

Have a use case in mind? [Open an issue](https://github.com/LucasPcq/herdr-wtm/issues).

## Contributing

```bash
go test ./...
HERDR_WTM_BUILD_FROM_SOURCE=1 sh scripts/install.sh   # `plugin link` does not run [[build]]
herdr plugin link "$PWD"
```

### Releasing

1. Set `version` in `herdr-plugin.toml` (e.g. `0.2.0`) and commit it on `main`.
2. Tag and push: `git tag v0.2.0 && git push origin v0.2.0`.

The Release workflow runs the tests, checks the tag matches the manifest version, and publishes the binaries with
GoReleaser. A tag like `v0.2.0-beta.1` becomes a pre-release.

## License

[MIT](LICENSE)
