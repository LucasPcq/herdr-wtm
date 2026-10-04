# Troubleshooting

## Where the logs are

The plugin logs to `herdr-wtm.log` in its herdr state directory (`~/.local/state/herdr/plugins/lucaspcq.wtm/` by default), and herdr keeps its own record of plugin runs:

```bash
tail -f ~/.local/state/herdr/plugins/lucaspcq.wtm/herdr-wtm.log
herdr plugin log list --plugin lucaspcq.wtm
```

Lines starting with `watch:` are the watcher's.

## No notification shows up

The plugin reports through herdr's notifications: workspaces opened or closed, a failed `on_create` hook, a wtm too old. herdr does not show them until you choose how to deliver them, in `~/.config/herdr/config.toml`:

```toml
[ui.toast]
delivery = "herdr"   # in-app toasts; "terminal" or "system" for desktop notifications
```

Then `herdr server reload-config`. Failures are also written to the log, whatever the setting.

## "herdr-wtm needs wtm 0.29 or later"

The watcher reads `wtm events` and `wtm version --output json`, both new in wtm 0.29. Run `wtm upgrade` (or `brew upgrade wtm`), then *Sync workspaces* to start the watcher again. If you point `wtm_bin` at a specific binary, check `"$wtm_bin" version --output json` reports `"events": 1`.

## A worktree did not open as a workspace

- **Is the repository shown in herdr?** The watcher only acts on repositories with a workspace, or a pane inside one of their worktrees. Open the project in herdr first.
- **Is the watcher running?** Run *Sync workspaces*: it starts the watcher if it stopped. The log says `watch: stream ended …` when wtm's stream was interrupted; it reconnects on its own.
- **Is the repository known to wtm?** The stream follows the repositories in wtm's registry; any wtm command run in the repository adds it.

## A workspace did not close

The watcher never closes a workspace whose folder still exists: check whether `wtm clean` stopped half-way (it says so) and left the folder. Remove it, then run *Sync workspaces*.

## The menu key does nothing

`alt` combinations depend on your terminal. Run `herdr plugin action invoke lucaspcq.wtm.bind` again and pick another key, such as `prefix+m` or `f12`.

## Starting over

```bash
herdr plugin uninstall lucaspcq.wtm
herdr plugin install LucasPcq/herdr-wtm
```
