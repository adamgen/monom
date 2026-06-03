# CLAUDE.md — AI Working Guide

Read these documents in order before making any changes:

1. `constitution.md` — governing principles. Every task must be validated against it.
2. `architecture.md` — current intended architecture: binary interface, shell files, data flow.
3. `terminology.md` — canonical term definitions. Use these exactly when naming anything.

---

## Spelling & Casing

The project name is **monom** — always lowercase, even at the start of a sentence.

---

## Temporary Files

The `tmp/` directory at the repo root is for scratch and temporary files (scratch scripts, intermediate output, log dumps, throwaway fixtures). Its contents are git-ignored (only `tmp/.gitkeep` is tracked, so the folder always exists).

- Write any temporary or throwaway files here instead of cluttering the repo root or source directories.
- Never rely on anything in `tmp/` persisting or being committed — treat it as disposable.
- Do not add real project artifacts (source, tests, specs, config) here.

---

## Testing

> **Note:** The test runner (`make check`) is pending build infrastructure — a separate change after `monomd-binary`. To run Go tests directly in the interim, use `go test ./...` from the repo root. To run a shUnit2 suite directly, use `bash tests/monomd_<subcommand>_test`. See `openspec/changes/monomd-binary/`.

### Conventions

- Go unit tests: `*_test.go`, colocated with the file they test.
- shUnit2 e2e tests: one file per subcommand under `tests/`, named `monomd_<subcommand>_test`.
- shUnit2 shared helpers: `tests/helpers` — sourced by every test file, never executed directly.
- shUnit2 test functions: `test_descriptive_name()`.

### What to test where

| What | Tool |
|---|---|
| Logic edge cases not observable from the binary surface | Go unit test |
| Pure function correctness (return values, error messages) | Go unit test |
| Full CLI binary surface (stdin, args, stdout, exit codes) | shUnit2 e2e test |
| Env var integration (`MONOM_PROJECT_ROOT`, `MONOM_USER_CONFIG`) | shUnit2 e2e test |
| Cross-package integration (e.g. `pack` calling `root`) | shUnit2 e2e test |
| Completion behavior in a real shell environment | shUnit2 completion e2e test |
| Shell binding files | shUnit2 test |

**Avoid testing the same scenario in both layers.** Go tests own logic correctness and unreachable-from-outside edge cases (panic recovery, walk stops at filesystem root, non-executable file skipped during walk, empty input). e2e tests own the binary's external contract. If a scenario is equally expressible in both, put it in e2e only.

### shUnit2 e2e test structure

One test file per subcommand under `tests/`, named `monomd_<subcommand>_test`. Shared fixtures and assertion helpers live in `tests/helpers`, which is sourced by every test file and never executed directly.

Each test file follows this pattern:

```sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MONOMD="$REPO_ROOT/bin/monomd"

. "$SCRIPT_DIR/helpers"

test_something() { ... }

. "$SHUNIT2"
```

---

---

## Common Tasks

### Build

```
make build    # compiles bin/monomd
make test     # go test ./...
make check    # test + shUnit2 e2e + shellcheck (pending build infrastructure change)
```

### Add a monomd subcommand

The four subcommands in scope for the initial binary are `filter`, `root`, `pack`, and `check`. `monomd args` is documented in `architecture.md` as TBD and is out of scope until a separate change defines it.

To add or extend a subcommand:

1. Add or update the logic package under `internal/<subcommand>/` with a `*_test.go` covering edge cases not testable from outside the binary
2. Wire the dispatch in `cmd/monomd/main.go`
3. Add or update `tests/monomd_<subcommand>_test` covering stdin, args, stdout, exit codes, and env var integration
4. Run `make check`

### Add a new shell binding

Shell files are a separate change after the binary. When that work begins:

1. Create `src/monom.<shell>` with only the completion registration
2. Update `src/monom` to source the right file based on shell detection
3. Verify shellcheck passes

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
- [ ] `go vet` passes with no errors
- [ ] All shell files pass `shellcheck` with no suppressions (except those documented inline with an explanation)
- [ ] shUnit2 e2e test added or updated if CLI behavior changed
- [ ] `make check` passes (once build infrastructure exists)
- [ ] No new required subcommands added to the user config interface (amendment required — see `constitution.md`)
- [ ] Terminology from `terminology.md` used consistently (do not redefine terms inline)
