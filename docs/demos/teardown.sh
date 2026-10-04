#!/usr/bin/env bash
# Stops what a recording started: the demo herdr session, its watcher, the demo
# wtm daemon and the demo tmux server. Leaves the files for inspection.
set -uo pipefail
here=$(cd "$(dirname "$0")" && pwd)
. "$here/env.sh"
[[ -d $HOME ]] || exit 0

lock="$HOME/.local/state/herdr/plugins/lucaspcq.wtm/watch.lock"
[[ -s $lock ]] && kill "$(head -1 "$lock")" 2>/dev/null
[[ -x $HERDR_WTM_DEMO_ROOT/bin/herdr ]] && herdr --session "$HERDR_WTM_DEMO_SESSION" server stop >/dev/null 2>&1
[[ -d $HERDR_WTM_DEMO_ROOT/acme ]] && (cd "$HERDR_WTM_DEMO_ROOT/acme" && wtm run daemon stop --yes >/dev/null 2>&1)
tmux -L "$HERDR_WTM_DEMO_SESSION" kill-server 2>/dev/null
exit 0
