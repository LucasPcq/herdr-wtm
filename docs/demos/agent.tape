# An agent creates worktrees in the next pane: they open in the sidebar while
# the focus stays with you, and a failed on_create hook raises a toast.
Output docs/assets/agent.gif
Source docs/demos/_env.tape

Hide
Type "export HERDR_WTM_DEMOS=$PWD/docs/demos; source docs/demos/env.sh; cd $HERDR_WTM_DEMO_ROOT/acme; exec zsh"
Enter
Sleep 1s
Type `tmux -L $HERDR_WTM_DEMO_SESSION -f $HERDR_WTM_DEMOS/tmux.conf new-session "herdr --session $HERDR_WTM_DEMO_SESSION"`
Enter
Wait+Screen@20s /spaces/
Sleep 1s
Type "$HERDR_WTM_DEMOS/agent-layout.sh && clear"
Enter
Wait+Screen@20s /agent ─/
Sleep 1s
Type "touch $HERDR_WTM_DEMO_ROOT/go && clear"
Enter
Sleep 500ms
Show

Sleep 1.5s
Type "git log --oneline"
Sleep 500ms
Enter
Wait+Screen@30s /feat\/checkout\s*│/
Sleep 1.5s
Type "ls apps"
Sleep 500ms
Enter
Wait+Screen@30s /fix\/payments\s*│/
Sleep 5s
Type "git branch"
Sleep 500ms
Enter
Sleep 4s
