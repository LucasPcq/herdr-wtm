#!/usr/bin/env bash
# Records every tape (or the one named) against a fresh demo world: `make demos`.
set -euo pipefail
cd "$(dirname "$0")/../.."
command -v vhs >/dev/null || { echo "vhs is required: brew install vhs" >&2; exit 1; }
make build >/dev/null
trap 'docs/demos/teardown.sh' EXIT
for tape in docs/demos/${1:-*}.tape; do
  [[ $(basename "$tape") == _* ]] && continue
  docs/demos/setup.sh
  vhs "$tape"
  docs/demos/teardown.sh
done
