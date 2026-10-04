# Writing the changelog

`CHANGELOG.md` is written in **English**, in the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) shape, and follows [semver](https://semver.org). Each release section is published verbatim as the GitHub release notes (`make release-notes`, run by the release workflow): write it for someone deciding whether to update the plugin, not for the reviewer of the PR.

## Template

Copy this under `## [Unreleased]` and keep only the sections that have entries, in this order.

```markdown
## [0.3.0] - 2026-11-02

One sentence on what this release is about.

### Highlights

- **Workspaces follow your worktrees** wherever they are created: a shell, an agent, `wtm ui`. → [How it works](docs/guide/how-it-works.md)

### Breaking

- **wtm 0.29 or later** is required: run `wtm upgrade`. → [Migrating to 0.2](docs/guide/migrating-to-0.2.md)

### Added

- **A notification** when a new worktree's `on_create` hooks fail.

### Changed

- **Sync workspaces** also restarts the watcher.

### Fixed

- **Cleaning the current worktree** lands on the main checkout even when it had no workspace.

### Removed

- **`focus_on_open`**: focus now follows your own actions only.
```

## Rules

- **Curate, don't inventory.** List what a user would notice: a new action or setting, a behaviour that changed under them, a bug they may have hit. A release with more than ~10 bullets is an inventory: cut.
- **Short.** About 20 words a bullet; the guide link carries the rest.
- **One bullet, one line, one change.** Bold the action, setting or behaviour, then say what the user gets, in the present tense. No "now", no "we", no internal names (packages, tickets, PR numbers).
- **Effect, not mechanism.** "A workspace closes when its worktree is removed", not how the event is read.
- **Breaking is its own section**, and every entry says what to do instead. When that takes more than one line, write `docs/guide/migrating-to-<version>.md` and link it.
- **Highlights** is optional: one to three bullets. A bullet there is not repeated under Added.
- **Internal-only changes are left out** (refactors, lint, tests) unless a user can see them.
- **Section titles are fixed**: `Highlights`, `Breaking`, `Added`, `Changed`, `Fixed`, `Removed`. The heading is `## [x.y.z] - YYYY-MM-DD`; the summary sentence carries the theme.
- Links into the docs are relative (`docs/guide/…`): `make release-notes` pins them to the release tag.
- Add a link reference per release at the bottom of the file (`[0.2.0]: https://github.com/LucasPcq/herdr-wtm/releases/tag/v0.2.0`).
