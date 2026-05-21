## Why

The monom prototype has been archived (moved to `_archive/`), leaving the repository root with only governance documents, OpenSpec artifacts, and `go.mod`. This change delivers the first correct implementation of the `monomd` Go binary — written from scratch into a clean repo, following the architecture document exactly. Companion changes will follow for the shell files, build scripts, and broader test infrastructure.

## What Changes

- Implement `monomd filter [word...]` from scratch: reads slash-delimited command paths from stdin, accepts space-separated typed words as args, returns next-level completions; silently skips paths with spaces in segments; SHALL never exit non-zero
- Implement `monomd pack <word...>` from scratch: takes space-separated args, joins with `/`, discovers project root internally, validates executable, prints absolute path
- Implement `monomd root` from scratch: honors `$MONOM_PROJECT_ROOT` when valid, otherwise walks up from `$PWD` looking for a directory containing an executable `monom` file
- Implement `monomd check` from scratch: runs `$MONOM_USER_CONFIG complete`, validates each path, reports problems; intended for development and CI use
- Constitution amendment (companion to this change): replace the "User Config Interface" principle with two principles — "Pluggability via Hooks" and "The Required User Config Interface Requires a Constitution Amendment to Change". Reduces the required interface to `complete` only; `run` becomes a documented optional hook.
- Go unit tests for each subcommand's logic; shUnit2 e2e tests for the binary's invocation surface

## Capabilities

### New Capabilities

- `filter-subcommand`: `monomd filter [word...]` — reads slash-delimited command paths from stdin, accepts space-separated typed words as args, returns next-level completions; always exits 0
- `pack-subcommand`: `monomd pack <word...>` — takes space-separated args, discovers project root, joins with `/`, resolves to absolute executable path
- `root-subcommand`: `monomd root` — env-aware project root discovery: honors `$MONOM_PROJECT_ROOT` when valid, otherwise walks up from `$PWD`
- `check-subcommand`: `monomd check` — validates the project's `complete` output for problems and reports them; intended for development and CI use

### Modified Capabilities

*(none — no existing specs)*

## Impact

- New `src/main.go`: `filter`, `pack`, `root`, `check` subcommand dispatch
- New `src/go_utils/filter.go`, `pack.go`, `root.go`, `check.go`: Go logic for each subcommand, each with a `*_test.go`
- A shared `findProjectRoot()` helper used by both `monomd root` and `monomd pack`
- Minimal shUnit2 e2e tests for `monomd` invocations (colocated with binary)
- `go.mod` updated to match the new `src/` package structure
- **Constitution amendment** to `constitution.md`: replace "User Config Interface" principle with "Pluggability via Hooks" + "Required User Config Interface" principles
- **Architecture documentation** in `architecture.md`: new "Hooks" section documenting `run` as the first optional hook; updated `monomd pack` and `monomd root` sections
- **Out of scope (separate changes):** shell files (`src/monom`, `src/monom.bash`), build scripts (`build`, `check`, `sh_test_runner`, `go_e2e_test`, `shellcheck`), broader `test_projects/` fixtures
