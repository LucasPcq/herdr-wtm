#!/usr/bin/env bash
# Builds the throwaway world every demo tape records against: an isolated HOME,
# a repository "acme" initialised with wtm, and this checkout linked as a plugin
# of an isolated herdr session. The real herdr, wtm daemon and repository
# registry are never touched.
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
plugin=$(cd "$here/../.." && pwd)
. "$here/env.sh"

"$here/teardown.sh"
rm -rf "$HERDR_WTM_DEMO_ROOT"
mkdir -p "$HOME/.config/herdr" "$HERDR_WTM_DEMO_ROOT/bin" "$HERDR_WTM_DEMO_ROOT/acme"

# Real herdr and wtm, plus a stand-in pnpm that fails on the branch the agent demo breaks.
ln -s "$HERDR_WTM_REAL_HERDR" "$HERDR_WTM_DEMO_ROOT/bin/herdr"
ln -s "$HERDR_WTM_REAL_WTM" "$HERDR_WTM_DEMO_ROOT/bin/wtm"
ln -s "$HERDR_WTM_REAL_TMUX" "$HERDR_WTM_DEMO_ROOT/bin/tmux"
cat > "$HERDR_WTM_DEMO_ROOT/bin/pnpm" <<'SH'
#!/bin/sh
case "$(git branch --show-current)" in
  *payments*) echo " ERR_PNPM_FETCH_404  GET https://registry.npmjs.org/@acme%2Fstripe-mock: Not Found" >&2; exit 1 ;;
esac
echo "Lockfile is up to date, resolution step is skipped"
echo "Done in 0.4s"
SH
chmod +x "$HERDR_WTM_DEMO_ROOT/bin/pnpm"

cp "$here/zshrc" "$HOME/.zshrc"
cat > "$HOME/.config/herdr/config.toml" <<'TOML'
onboarding = false

[ui.toast]
delivery = "herdr"

[ui.toast.herdr]
position = "bottom-right"
TOML

cd "$HERDR_WTM_DEMO_ROOT/acme"
git init -q -b main
git config user.email demo@acme.dev
git config user.name "Acme Dev"
mkdir -p apps/web apps/api
echo '{ "name": "web" }' > apps/web/package.json
echo '{ "name": "api" }' > apps/api/package.json
echo "# acme" > README.md
git add -A && git commit -qm "chore: bootstrap acme"

wtm init --yes --base-path ../acme.trees </dev/null >/dev/null 2>&1
perl -0pi -e 's/on_create = \[\n\]/on_create = [\n  "pnpm install",\n]/' .git/wtm/config.toml

herdr --session "$HERDR_WTM_DEMO_SESSION" plugin link "$plugin" >/dev/null
