# Architecture

herdr-wtm keeps herdr's workspaces in step with wtm's worktrees and runs wtm from a herdr popup. Two processes do the work: a long-lived **watcher** and a short-lived **popup**.

```
herdr startup ──► herdr-wtm watch --detach ──► herdr-wtm watch (own session, flock)
                                                   │
                                     wtm events --output json   (cwd=/, every repository)
                                                   │ snapshot · ready · worktree.* · repo.*
                                                   ▼
                                     Watcher.Handle ──► herdr worktree open / workspace close / focus
                                                   ▲
key ──► herdr-wtm launch <cmd> ──► popup: herdr-wtm run ──► WTM_CORRELATION_ID=herdr-wtm:<rand> wtm <cmd>
```

## Packages

| Package | Role |
| --- | --- |
| `cmd/herdr-wtm` | `main`: env and config → `app.Deps`; dispatch of `launch`, `run`, `sync`, `watch [--detach]`; detaching the watcher; the popup's signal shield |
| `internal/domain` | `Worktree`, `Workspace`, `HerdrContext`, `Event`, `Contracts`, `FS`; constants (env, event types, wtm exit codes, timings, file names); `ErrWtmTooOld`, `ErrWtmSchemaTooNew` |
| `internal/rules` | `Stale`, `WorkspaceAt`, `MainWorkspace`, `RepoShown`, `BranchAt`, `IsOwnCorrelation`, `ProvisionFailure`, `Summary`, `IsPopupCommand` |
| `internal/app` | `Launch`, `Run` (popup), `Sync`, `Watch` + `Watcher` + `notes`, `Bind` |
| `internal/wtm` | `List`, `Run` (correlated), `Resolve`, `Contracts`, `Events` |
| `internal/herdr` | `Workspaces`, `PaneCWDs`, `OpenWorktree`, `Focus`, `Close`, `Notify`, `OpenPopup`, `DefaultConfig`, `ReloadConfig`, `ParseContext` |
| `internal/execx` | `Runner` (`Output`, `Interactive`, `Stream`), `ExitCode`, `Fake` |
| `internal/fsx` | `Normalize` (symlinks resolved on the deepest existing ancestor), `Exists`, `OS()` |
| `internal/gitx` | `RepoRoot` |
| `internal/config` | `config.toml` |
| `internal/menu` | the popup menu (Bubble Tea) |
| `internal/keybind` | reading herdr's bindings and writing the menu's, as text |

## The watcher

- **One per herdr server.** `[[startup]]` and every `launch`/`sync` run `watch --detach`: herdr-wtm re-executes itself as `watch` with `setsid` and returns. The new process takes an exclusive `flock` on `$HERDR_PLUGIN_STATE_DIR/watch.lock`; a second one finds it held and exits. It stops when `herdr workspace list` fails (herdr is gone), checked every `WatchLivenessTick`.
- **Compatibility first.** `wtm version --output json` must report `events >= MinEventsVersion`; otherwise it notifies `ErrWtmTooOld` and exits.
- **One global stream.** `wtm events --output json` from `/` follows every repository in wtm's registry. Exit `20` (an event newer than this wtm) and `2` stop the watcher; anything else restarts the stream with a backoff from 1 s to 30 s, reset once a stream reaches `ready`. A new stream opens on snapshots, which close what disappeared meanwhile; a worktree created during the gap is not opened (snapshots never open).
- **Only repositories herdr shows.** An event about repository R is acted on only if a workspace's `worktree.repo_root` is R, or a pane's cwd lies inside one of R's worktrees (herdr reports no `worktree` for a workspace it did not open as one). The watcher remembers each repository's worktree paths from its snapshot and later events for that test.
- **Handling.** `snapshot` closes stale workspaces (`rules.Stale`: linked, of this repository, gone from the snapshot and from disk). `worktree.created` / `relocated` open a workspace unless one is open at that path. `worktree.provisioned` with `ok: false` notifies. `worktree.removed` closes the linked workspace at that path if the folder is gone. Everything else is ignored.
- **Focus follows the user.** An event whose `correlation_id` starts with `herdr-wtm:` comes from the popup: a created worktree opens focused, and removing the focused worktree first focuses the main checkout (opening it if needed). Any other event (an agent, another shell, `wtm ui`) never moves the focus.
- **Notifications** are gathered by `notes` and sent once per burst (`NotifyQuietWindow` of silence).

## The popup

`launch` resolves the repository from herdr's plugin context, starts the watcher, and opens the `run` entrypoint as a popup with `HERDR_WTM_CMD`, `HERDR_WTM_REPO`, `HERDR_WTM_ORIGIN`. `run` runs the command — or the menu, then the chosen command — under a signal shield, with a fresh `WTM_CORRELATION_ID`. It reads `wtm list` only to name the origin worktree's branch (menu label, `clean`). `open` is the exception that acts on herdr itself: `wtm resolve` emits no event, so the popup focuses or opens the chosen worktree.

`sync` (menu entry and action) starts the watcher and closes the stale workspaces of one repository from a one-off snapshot (`wtm events --repo R`, stopped at `ready`).

## Testing

`execx.Fake` answers every process by its command line, and `Stream` replays the answer line by line, so the watcher, the popup and sync are tested end to end without wtm or herdr. `domain.FS` is injected, so path aliasing (symlinks) and on-disk checks are part of the fixtures.
