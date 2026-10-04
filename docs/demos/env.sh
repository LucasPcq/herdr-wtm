# Sourced by every demo script and tape: the isolated world's environment.
# Resolve the real binaries before PATH and HOME change.
: "${HERDR_WTM_REAL_HERDR:=$(command -v herdr)}"
: "${HERDR_WTM_REAL_WTM:=$(command -v wtm)}"
: "${HERDR_WTM_REAL_TMUX:=$(command -v tmux)}"
export HERDR_WTM_REAL_HERDR HERDR_WTM_REAL_WTM HERDR_WTM_REAL_TMUX
export HERDR_WTM_DEMO_ROOT=${HERDR_WTM_DEMO_ROOT:-/tmp/herdr-wtm-demo}
export HERDR_WTM_DEMO_SESSION=herdr-wtm-demo
export HOME=$HERDR_WTM_DEMO_ROOT/home
export PATH=$HERDR_WTM_DEMO_ROOT/bin:/usr/bin:/bin:/usr/sbin:/sbin
export TERM=xterm-256color LANG=en_US.UTF-8 WTM_NO_UPDATE_CHECK=1
# Inherited from a herdr pane, these would point the demo at the real session.
unset HERDR_BIN_PATH HERDR_ENV HERDR_PANE_ID HERDR_SOCKET_PATH HERDR_TAB_ID HERDR_WORKSPACE_ID TMUX
