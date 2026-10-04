# How it works

herdr-wtm has two parts: a **watcher** that keeps herdr's workspaces in step with wtm's worktrees, and a **popup** that runs wtm commands. Neither changes what wtm does.

## The watcher

When herdr starts, the plugin starts one watcher in the background (one per herdr server; starting another is a no-op). It runs `wtm events`, wtm's live stream of worktree changes, for every repository wtm has been used in, and stops when herdr stops. Opening the menu, or running *Sync workspaces*, starts it again if it is not running.

It acts only on the **repositories herdr shows**: a repository with at least one workspace, or a pane whose directory is inside one of its worktrees. A worktree an agent creates in a project you have not opened in herdr does not open a workspace.

| In wtm | In herdr |
| --- | --- |
| a worktree is created (`create`, `checkout`, `extract`, `wtm ui`, an agent, another shell) | a workspace opens for it, unless one is already open there |
| a worktree is moved (`wtm relocate`) | a workspace opens at its new place |
| its `on_create` hooks fail | a notification names the hook and its exit code |
| a worktree is removed (`clean`, `prune`) | its workspace closes, if the folder is gone |
| the watcher starts, or wtm's daemon restarts | workspaces of worktrees that disappeared meanwhile close |

Several changes at once (`wtm create a b c`) make one notification.

## What is never closed

- the main checkout's workspace;
- a workspace whose folder still exists on disk — a `clean` that stopped half-way leaves your files, and the workspace, alone;
- a workspace herdr did not open as a worktree.

## Focus

What you start from the popup carries a tag (`WTM_CORRELATION_ID=herdr-wtm:…`) that wtm copies onto the events it causes. The watcher uses it for focus only:

- a worktree you create from the popup opens **focused**;
- cleaning the worktree you are in moves you to the **main checkout** first (opening it if it has no workspace), then closes the old workspace;
- everything else — an agent, another shell, `wtm ui` — opens or closes workspaces **without moving the focus**.

## The popup

The popup runs the wtm command you chose, as you would in a terminal: wizards, pickers and errors are wtm's own. When it fails, the error stays on screen until you press Enter. *Open a worktree* is the one entry that acts on herdr directly: wtm's picker returns a path, and the popup focuses that workspace or opens it.

## Sync workspaces

*Sync workspaces* is the repair tool: it restarts the watcher if needed, reads a fresh list of the repository's worktrees, and closes what is left behind. It says so when there was nothing to do.
