# CLAUDE.md — herdr-wtm

herdr-wtm is a [herdr](https://herdr.dev) plugin for [wtm](https://github.com/LucasPcq/wtm). wtm is the product; this plugin is one of its consumers. It never asks wtm to know about herdr: it reads wtm's generic integration contract (`wtm events`, `WTM_CORRELATION_ID`, `wtm version --output json`, exit codes). A need wtm cannot meet goes to wtm as a generic feature, never as a herdr special case.

**Read on demand** — `docs/dev/` is the developer reference; open the page before working on its topic:

| Read | Before |
| -- | -- |
| [`docs/dev/architecture.md`](docs/dev/architecture.md) | adding a package or an import between layers, touching the watcher, the popup or the event handling |
| [`docs/dev/changelog.md`](docs/dev/changelog.md) | writing a `CHANGELOG.md` entry |
| [`docs/dev/releasing.md`](docs/dev/releasing.md) | cutting a release or a beta |

**Self-maintaining docs:** when a structural decision changes (new package, new layer rule, new dependency, new convention), update this file and the relevant `docs/dev/` page in the same session. Standards must reflect the actual codebase.

## Docs & README

- **`README.md` is the product page**, not a reference: pitch, why herdr + wtm, install, usage, short configuration, a short "How it works" linking the guide. Anything longer belongs in `docs/guide/`.
- **`docs/guide/`** (user guide) and **`docs/dev/`** (developer docs) are hand-written. A behaviour a user needs to understand goes in the guide, in the same change as the behaviour.
- **`CHANGELOG.md`** is in English, Keep a Changelog shape, following [`docs/dev/changelog.md`](docs/dev/changelog.md). Each release section is published as the GitHub release notes by `make release-notes VERSION=x.y.z`.
- **Markdown:** never hard-wrap prose — one paragraph is one line.
- `docs/superpowers/` is local and gitignored: design notes, never shipped.

## 1. Immutability first

Prefer `:=` for values that do not change; `var` only for zero values or package-level declarations. If a block reassigns a variable, extract it into a function.

## 2. Structs for 2+ inputs

A function taking 2 or more of **its own** inputs takes a single `<Name>Params` struct with named fields (`OpenWorktree(OpenParams{Repo, Path, Focus})`). Carriers don't count (`context.Context`, `io.Writer`, `execx.Runner`, `*testing.T`), and symmetric pairs whose order is their meaning are exempt. A review rule, not a lint rule.

## 3. Shared types — no duplication

Types, sentinel errors and constants are defined once in `internal/domain/`, which has no functions or methods. Pure decisions live in `internal/rules/`.

## 4. Validate all external input

`config.toml`, environment variables and the herdr plugin context are validated at the boundary (`internal/config`, `cmd/herdr-wtm`). A line of `wtm events` that does not decode is skipped, as wtm's contract asks; an unknown event type is ignored.

## 5. Centralized constants — no magic strings or numbers

Every env var, exit code, event type, file name, subcommand and duration is a named constant in `internal/domain/`. The argv of the wtm and herdr CLIs is the one exception: it is spelled only in its adapter (`internal/wtm`, `internal/herdr`), which is the adapter's whole job.

## 6. Early returns — no nesting

Every error or guard returns immediately; the happy path is last.

## 7. No unsafe type assertions

Always comma-ok, and type at the source (decode JSON into structs, not `any`).

## 8. Comments — the exception, not the rule

Near-zero comments: names and signatures carry the meaning. Write one only for a why (a non-obvious decision, an ordering constraint, a workaround with its reference), a one-line package comment, or godoc when the name leaves a caller guessing. Architecture belongs in `docs/dev/`. When you modify a file, delete the comments in it that restate the code.

## 9. Layers

```
cmd/herdr-wtm/   dispatch only: argv/env → an app entrypoint; detaching the watcher
internal/
  domain/        types, sentinel errors, constants — no functions
  rules/         pure functions: stdlib + domain only, the filesystem injected as domain.FS
  app/           one orchestration per entrypoint: launch, run, sync, watch, bind
  wtm/  herdr/   CLI adapters: exec + JSON decoding, no decisions
  gitx/          the one git question the plugin asks
  execx/         process runner (OS) and its test double (Fake)
  fsx/           path normalisation and existence: the plugin's only filesystem reads besides bind's config edit
  config/        config.toml, validated
  menu/          Bubble Tea model, rendering only
  keybind/       herdr key bindings as text (parses TOML, so not in rules/)
```

- `rules/` imports only the stdlib and `domain/`.
- `wtm/`, `herdr/`, `gitx/`, `config/`, `menu/`, `keybind/` never import each other or `app/`; only `app/` composes them.
- `cmd/` holds no decision: it maps argv and env onto `app.Deps` methods.
- The watcher is the only code that opens, closes or focuses workspaces after a wtm command; the popup only runs wtm (and `open`, which has no event). See `docs/dev/architecture.md`.

Checked by review; `make lint` checks everything a general-purpose linter can (gofmt, vet, deadcode, staticcheck).

## 10. Commit messages in English

Subject and body in English, prefixed as the history does (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `ci:`, `chore:`), whatever language the conversation was held in.

## 11. Validate before commit

```
make lint          # gofmt + vet + deadcode + staticcheck — all gating
make test          # go test ./... -race -count=1
make build         # bin/herdr-wtm
make release-notes VERSION=x.y.z   # print a CHANGELOG section as release notes
make snapshot      # every release archive, unpublished (needs goreleaser)
```

- `.claude/hooks/pre-commit-gates.sh` runs `make lint` and a `go mod tidy` check on every `git commit` and blocks it on failure; it does not run the tests. `HERDR_WTM_SKIP_GATES=1 git commit …` only when the gate itself is wrong.
- **Invoke the `build-validator` subagent before marking any task done** — it adds the `-race` test suite and dependency hygiene.
- Trying it in herdr: `make build && herdr plugin link "$PWD"` (`plugin link` does not run `[[build]]`), then stop the running watcher (`pkill -f "herdr-wtm watch"`) and run the `sync` action, or restart herdr: a running watcher keeps its binary and config.
