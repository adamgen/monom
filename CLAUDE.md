# CLAUDE.md — AI Working Guide

Read these documents in order before making any changes:

1. `constitution.md` — governing principles. Every task must be validated against it.
2. `architecture.md` — current intended architecture: binary interface, shell files, data flow.
3. `terminology.md` — canonical term definitions. Use these exactly when naming anything.

Before changing the behavior of a package, also read the `TRADEOFFS.md` beside it, if one exists.

---

## Spelling & Casing

The project name is **monom** — always lowercase, even at the start of a sentence.

---

## How This Repo Documents Itself

monom is **not** spec-driven. There are no specification documents, no change proposals, and no task files. Do not create them, and do not ask for them before starting work.

The governing rule is the constitution's *Documentation Is Load-Bearing or Absent*. In practice:

| Where | Holds | Example |
| --- | --- | --- |
| The code and its tests | All behavior, structure, and contract | What `filter.Filter` returns for a given input |
| `constitution.md` | Invariants that resolve conflicts before they're argued | "Go owns logic, shell owns surface" |
| `architecture.md` | The intended shape, and the interfaces between parts | The `run` hook's exit-code contract |
| `terminology.md` | Names | What "command packing" means |
| `<dir>/TRADEOFFS.md` | Contested decisions, next to the code implementing them | Why `filter` never exits non-zero |
| `BACKLOG.md` | Intended, unstarted work — one entry, no ceremony | `mnmd args` |
| `INSTALL.md` | The procedure an agent follows to install monom on a user's machine | Verifying the rc line loads in an interactive shell |

### When to write a `TRADEOFFS.md` entry

Only when a decision was **genuinely contested**: a real alternative existed, and the reason it lost is not obvious from reading the outcome. The bar is "a competent contributor would otherwise 'fix' this, or would re-litigate it from scratch."

Each entry states four things:

1. **Chosen** — what the code does.
2. **Rejected** — the specific alternative, concretely enough to recognize.
3. **Why** — the reasoning, including the failure mode the rejected option produces.
4. **What it costs** — the price of the choice. An entry with no cost is usually not a tradeoff.

Do **not** write an entry for a decision with no real alternative, for restating what the code does, or for a preference. If you find yourself explaining *what* rather than *why*, delete it and improve the code instead.

### When to touch a canonical document

Add to `architecture.md` when a new part, interface, or data-flow path exists that a reader could not infer from any single package. Add to `constitution.md` only for an invariant — and note that changing the required user config interface needs an explicit amendment. Add to `terminology.md` when a new domain term enters use.

Keep contracts in `architecture.md` and rationale in `TRADEOFFS.md`. Do not duplicate one into the other; a second copy will drift.

### When to touch `BACKLOG.md`

Add an entry when work is identified but not being started now — a few sentences of intent and why it matters. Remove the entry when the work lands or is abandoned. Backlog entries are not commitments and carry no design.

---

## Temporary Files

The `tmp/` directory at the repo root is for scratch and temporary files (scratch scripts, intermediate output, log dumps, throwaway fixtures). Its contents are git-ignored (only `tmp/.gitkeep` is tracked, so the folder always exists).

- Write any temporary or throwaway files here instead of cluttering the repo root or source directories.
- Never rely on anything in `tmp/` persisting or being committed — treat it as disposable.
- Do not add real project artifacts (source, tests, config) here.

---

## Testing

Run everything with `make check`. See `make help` for the full target list. `tests/README.md` is the guide to writing tests, including the declarative case format.

### Conventions

- Go unit tests: `*_test.go`, colocated with the file they test.
- Declarative CLI cases: `tests/cases/<area>.yaml`, one file per surface, run in bash and zsh by the Go test `TestCases` (`tests/harness/cases_test.go`, run by `make test-cases`). `tests/harness` is a separate Go module holding the test-only dependencies (yaml.v3, creack/pty, vt10x); the root `go.mod` stays dependency-free, so never add them there. Each case is an `input` line, an `action` (`enter`, `tab`, or `keys` for real Tab presses in an interactive shell in a pseudo-terminal), and the expected result (`expect`, or `line`/`candidates` for keys); the project root comes from `root:` at the suite or group level. The harness validates the YAML strictly.
- shUnit2 e2e tests: one file per surface under `tests/`, named `mnmd_<subcommand>_test` or `monom_<area>_test`.
- shUnit2 shared helpers: `tests/helpers` — sourced by every test file, never executed directly.
- shUnit2 test functions: `test_descriptive_name()`.

### What to test where

| What | Tool |
|---|---|
| Logic edge cases not observable from the binary surface | Go unit test |
| Pure function correctness (return values, error messages) | Go unit test |
| What a user sees after typing a `monom`/`mnmd` line and pressing Enter or Tab | Declarative case (`tests/cases/*.yaml`) |
| What the line editor does on Tab (inserted text, listing, zsh menu) | Declarative `keys` case (`tests/cases/monom_keys.yaml`) |
| Full CLI binary surface (stdin, args, stdout, exit codes) | shUnit2 e2e test |
| Env var integration (`_MONOM_PROJECT_ROOT`, `_MONOM_USER_CONFIG`, `MONOM_CHECK_ROOT_CONTENTS`) | Declarative case with `env:` when it fits; otherwise shUnit2 e2e test |
| Cross-package integration (e.g. `pack` calling `root`) | shUnit2 e2e test |
| Completion candidates in a real shell environment | Declarative case with `action: tab` |
| Completion registration and function tables | shUnit2 completion e2e test |
| Shell binding files | shUnit2 test |

**Avoid testing the same scenario in both layers.** Go tests own logic correctness and unreachable-from-outside edge cases (panic recovery, walk stops at filesystem root, non-executable file skipped during walk, empty input). e2e tests own the binary's external contract. If a scenario is equally expressible in both, put it in e2e only.

Tests are documentation. When a behavior is subtle, the test name is where you say so — not a comment restating the assertion.

### shUnit2 e2e test structure

Shared fixtures and assertion helpers live in `tests/helpers`, which is sourced by every test file and never executed directly. `fixtures/demo-project/` is a complete, runnable example project; point tests at it rather than building an ad-hoc tree inline. `fixtures/hooks-project/` adds a `run` hook (an alias and a deliberate failure) and an empty command group. `fixtures/zero-config-project/` has no monom config file and one executable per discovery-gate outcome, and `fixtures/zero-config-variants/*` are small projects that each have a different monom file (empty, declarations, strict severity, `discover.hide`, hook scripts). Declarative cases select a fixture with `root:`. A fixture without a monom file resolves to this repository's git root unless the case pins it with `env: {_MONOM_PROJECT_ROOT: "{root}"}`. shUnit2 tests that must write a `monom` file copy the fixture with `make_zero_config_project`; add a new fixture under `fixtures/` rather than generating one inside a case.

Each test file follows this pattern:

```sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MONOMD="$REPO_ROOT/bin/mnmd"

. "$SCRIPT_DIR/helpers"

test_something() { ... }

. "$SHUNIT2"
```

---

## Common Tasks

### Build and check

```
make help     # list all targets
make build    # compiles bin/mnmd
make check    # build + gofmt + go vet + go test + shUnit2 e2e + declarative cases + shellcheck
```

### Add a mnmd subcommand

Current subcommands: `filter`, `root`, `pack`, `discover`, `check`, `install`.

1. Add or update the logic package under `internal/<subcommand>/` with a `*_test.go` covering edge cases not testable from outside the binary
2. Wire the dispatch in `cmd/mnmd/main.go` and add the subcommand to `usage()`
3. Return errors that carry their exit code — use `cli.WrapError` or a type embedding the registry's codes; never call `os.Exit` with a literal
4. Add or update `tests/mnmd_<subcommand>_test` covering stdin, args, stdout, exit codes, and env var integration
5. Document the contract in `architecture.md`; if a design alternative was rejected, add a `TRADEOFFS.md` entry in the package
6. Run `make check`

### Change shell behavior

Shell files exist only where Go is technically impossible. Before adding a line to `src/`, answer the constitution's question — *is there a technical reason this cannot be in Go?* — and if the answer is no, put it in Go.

`src/TRADEOFFS.md` records the shell-layer decisions that are load-bearing (zsh word-splitting, the `exec` subshell, the writability probe). Read it before editing `src/monom`; several lines there look redundant and are not.

### Fix spurious shellcheck errors in the IDE

`.vscode/settings.json` uses `"*": "shellscript"` as a catch-all. Any new file type without a more specific entry will be treated as a shell script, causing the IDE's shellcheck extension to report false errors (e.g. SC2148 "shebang missing") on non-shell files.

When a file gets a spurious shellcheck error in the editor, add an explicit language association to `.vscode/settings.json`:

```json
"Makefile": "makefile"
```

Do not add a shebang or `# shellcheck shell=` directive to non-shell files — fix the association instead.

---

## Validation Checklist

Before completing any task:

- [ ] No logic added to shell files
- [ ] No new subprocess roundtrips on the completion or run path
- [ ] Go unit tests added for any new Go logic
- [ ] shUnit2 e2e test added or updated if CLI behavior changed
- [ ] `make check` passes
- [ ] No new required subcommands added to the user config interface (amendment required — see `constitution.md`)
- [ ] Terminology from `terminology.md` used consistently (do not redefine terms inline)
- [ ] No prose written that restates what the code says
- [ ] Any rejected design alternative recorded in the relevant `TRADEOFFS.md`, with its cost
- [ ] `architecture.md` updated if a contract or interface changed
- [ ] `INSTALL.md` updated if building, `mnmd install`, the checkout layout, or shell loading changed
