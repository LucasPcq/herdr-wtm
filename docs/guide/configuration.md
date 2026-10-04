# Configuration

## `config.toml`

Optional, in `$(herdr plugin config-dir lucaspcq.wtm)/config.toml`:

| Key | Default | Meaning |
| --- | --- | --- |
| `wtm_bin` | `"wtm"` | the wtm binary: a name on `PATH` or an absolute path (handy to try a local build) |
| `popup_width` | `"90%"` | popup width, as herdr takes it (`"90%"`, `"120"`) |
| `popup_height` | `"90%"` | popup height |

An empty value is refused with a notification; an unknown key is ignored. The file is read each time the plugin runs, so a change applies to the next popup; restart the watcher (*Sync workspaces*, or restart herdr) for it to apply there too.

## The menu key

`herdr plugin action invoke lucaspcq.wtm.bind` asks for a key and writes this to herdr's `config.toml`, then reloads it:

```toml
# herdr-wtm plugin
[[keys.command]]
key = "prefix+alt+w"
type = "plugin_action"
command = "lucaspcq.wtm.menu"
description = "wtm menu"
```

Run it again to change the key: the previous binding is replaced, the rest of the file is kept, and a backup is written next to it (`config.toml.bak-herdr-wtm`). If herdr rejects the new config, the previous one is restored.

## One key per action

Every menu entry is also an action: `lucaspcq.wtm.create`, `.open`, `.checkout`, `.clean`, `.prune`, `.ui`, `.sync`. Bind one like the menu, with its own `command`:

```toml
[[keys.command]]
key = "prefix+alt+n"
type = "plugin_action"
command = "lucaspcq.wtm.create"
description = "new worktree"
```
