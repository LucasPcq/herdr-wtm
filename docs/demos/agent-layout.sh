#!/usr/bin/env bash
# Run hidden inside herdr's pane by agent.tape: once the watcher is up, opens
# the "agent" pane on the right without taking the focus, and starts agent.sh.
set -euo pipefail
"$HERDR_WTM_DEMOS/wait-ready.sh"
agent=$(herdr pane split --current --direction right --ratio 0.5 --no-focus |
  python3 -c 'import json, sys; print(json.load(sys.stdin)["result"]["pane"]["pane_id"])')
herdr pane rename "$HERDR_PANE_ID" you >/dev/null
herdr pane rename "$agent" agent >/dev/null
herdr pane run "$agent" "$HERDR_WTM_DEMOS/agent.sh" >/dev/null
