# Migrating to 0.2

0.2 replaces the popup's own bookkeeping with wtm's event stream: workspaces now follow every worktree change, not only the ones made from the popup.

## Upgrade wtm first

herdr-wtm 0.2 needs **wtm 0.29 or later**. With an older wtm the plugin tells you so in a notification and does nothing else:

```bash
wtm upgrade
wtm version --output json   # "events": 1
```

## `focus_on_open` is gone

0.1 focused the last workspace it opened after any popup command, unless `focus_on_open = false`. 0.2 focuses what *you* start from the popup and never what an agent or another shell does, so the setting has no job left. Remove it from `config.toml` if you like; it is ignored either way.

## Sync at startup is gone

0.1 ran *Sync workspaces* for every repository when herdr started. 0.2 starts the watcher instead, which does the same on its first snapshot and keeps doing it. `herdr-wtm sync --all` no longer exists; the *Sync workspaces* action stays, for one repository, and also starts the watcher if it stopped.

## Nothing else to do

The menu key, the actions and their ids are unchanged.
