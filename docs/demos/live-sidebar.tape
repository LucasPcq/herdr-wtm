# Worktrees created and removed from a shell outside herdr show up in its
# sidebar as they happen, without moving the focus.
Output docs/assets/live-sidebar.gif
Source docs/demos/_env.tape

Hide
Type "export HERDR_WTM_DEMOS=$PWD/docs/demos; source docs/demos/env.sh; cd $HERDR_WTM_DEMO_ROOT/acme; exec zsh"
Enter
Sleep 1s
Type `tmux -L $HERDR_WTM_DEMO_SESSION -f $HERDR_WTM_DEMOS/tmux.conf new-session "herdr --session $HERDR_WTM_DEMO_SESSION" \; split-window -h -l 38% "$HERDR_WTM_DEMOS/wait-ready.sh; clear; exec zsh"`
Enter
Wait+Screen@20s /spaces/
Sleep 8s
Show

Sleep 1.5s
Type "wtm create feat/login feat/search --yes"
Sleep 600ms
Enter
Wait+Screen@30s /feat\/search\s*│/
Sleep 4.5s

Type "wtm clean feat/search --yes"
Sleep 600ms
Enter
Sleep 6s
