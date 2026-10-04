#!/usr/bin/env bash
# The "agent" of agent.tape: narrates, then drives wtm like an agent would.
say() { printf '\n\033[35m●\033[0m %s\n' "$1"; sleep 1.2; }
clear
# agent.tape touches the go file once recording starts.
while [[ ! -e $HERDR_WTM_DEMO_ROOT/go ]]; do sleep 0.2; done
sleep 3
say "Starting the checkout flow in its own worktree"
wtm create feat/checkout --yes
sleep 2.5
say "Splitting the payments fix out of it"
wtm create fix/payments --yes
sleep 30
