# Changelog

All notable changes to herdr-wtm are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and herdr-wtm adheres to [Semantic Versioning](https://semver.org); how to write an entry is in [docs/dev/changelog.md](docs/dev/changelog.md).

## [Unreleased]

## [0.2.0] - 2026-10-04

Workspaces follow your worktrees wherever they change, on top of wtm 0.29's event stream.

### Highlights

- **Workspaces follow every worktree change** — from the popup, another shell, an agent or `wtm ui` — as it happens, in the repositories herdr shows. → [How it works](docs/guide/how-it-works.md)
- **Focus follows you, not your agents**: what you start from the popup is focused, nothing else moves it. → [Focus](docs/guide/how-it-works.md#focus)

### Breaking

- **wtm 0.29 or later** is required: run `wtm upgrade`. → [Migrating to 0.2](docs/guide/migrating-to-0.2.md)
- **`focus_on_open`** is removed: focus follows your own actions. → [Migrating to 0.2](docs/guide/migrating-to-0.2.md#focus_on_open-is-gone)

### Added

- **A notification** when a new worktree's `on_create` hooks fail, naming the hook and its exit code.

### Changed

- **Sync workspaces** reads wtm's current state and starts the watcher if it stopped.
- **Startup** starts the watcher instead of syncing every repository once. → [Migrating to 0.2](docs/guide/migrating-to-0.2.md#sync-at-startup-is-gone)

## [0.1.0] - 2026-10-02

First release: wtm's commands in a herdr popup, and workspaces kept in sync with what they change.

### Added

- **A wtm menu** in a herdr popup: new worktree, open, checkout a PR, clean, prune, dashboard, sync.
- **A bind action** to choose the menu key, `prefix+alt+w` by default, refusing keys herdr already uses.
- **Workspaces follow the popup's commands**: a created worktree opens, a removed one closes, cleaning the current worktree lands on the main checkout.
- **Sync workspaces** closes workspaces whose worktree was removed elsewhere, also run when herdr starts.
- **Prebuilt binaries** for macOS and Linux, checksum-verified at install, with a source build as fallback.

[Unreleased]: https://github.com/LucasPcq/herdr-wtm/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/LucasPcq/herdr-wtm/releases/tag/v0.2.0
[0.1.0]: https://github.com/LucasPcq/herdr-wtm/releases/tag/v0.1.0
