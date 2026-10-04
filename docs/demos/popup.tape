# The menu key opens the compact menu; New worktree reopens the popup at the
# wizard's size, and the new worktree opens as the focused workspace. Back in
# the menu from there, Clean names the worktree you are in.
Output docs/assets/popup.gif
Source docs/demos/_env.tape

Hide
Type "export HERDR_WTM_DEMOS=$PWD/docs/demos; source docs/demos/env.sh; cd $HERDR_WTM_DEMO_ROOT/acme; exec zsh"
Enter
Sleep 1s
Type `tmux -L $HERDR_WTM_DEMO_SESSION -f $HERDR_WTM_DEMOS/tmux.conf new-session "herdr --session $HERDR_WTM_DEMO_SESSION"`
Enter
Wait+Screen@20s /spaces/
Type "$HERDR_WTM_DEMOS/wait-ready.sh && clear"
Enter
Sleep 4s
Show

Sleep 1.5s
Ctrl+B
Sleep 150ms
Type "m"
Wait+Screen@15s /WORKTREES/
Sleep 1.5s
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms
Up
Sleep 1s
Enter
Wait+Screen@15s /Step 1\/6/
Sleep 1s
Type "feat/login"
Sleep 800ms
Enter
Wait+Screen@15s /Step 2\/6/
Sleep 1.5s
Enter
Wait+Screen@15s /Step 3\/6/
Sleep 1.5s
Enter
Wait+Screen@15s /Review/
Sleep 2.5s
Enter
Wait+Screen@30s /feat-login ❯/
Sleep 3s

Ctrl+B
Sleep 150ms
Type "m"
Wait+Screen@15s /CLEAN UP/
Sleep 3.5s
Escape
Sleep 2s
