#!/usr/bin/env bash
# Run hidden at the start of a tape: returns once the plugin's watcher has
# taken its first snapshot of acme, so the recording starts on a live sidebar.
log="$HOME/.local/state/herdr/plugins/lucaspcq.wtm/herdr-wtm.log"
for _ in $(seq 100); do
  grep -q "watch: started" "$log" 2>/dev/null && sleep 1.5 && exit 0
  sleep 0.2
done
echo "watcher did not start: see $log" >&2
exit 1
