## Context

monom currently has no root-level `Makefile`. Contributors must read `CLAUDE.md` to discover commands (`go test ./...`, `shellcheck`, etc.), and there is no single canonical entry point for CI or new contributors. The `CLAUDE.md` explicitly lists `make build`, `make test`, and `make check` as the intended interface, but those targets do not yet exist.

The repo structure relevant to this change:
- `cmd/monomd/` — Go binary source
- `bin/monomd` — compiled output (gitignored)
- `tests/` — shUnit2 e2e suites, one file per subcommand
- `src/` — shell binding files
- `dependencies/shunit2` — vendored shUnit2

## Goals / Non-Goals

**Goals:**
- Single discoverable entry point for all common developer tasks (`build`, `test`, `check`, `lint`, `clean`, `help`)
- `make help` self-documents all targets via `##` comments
- Makefile evolves incrementally — targets are added as the project adds capabilities, not speculatively
- Wraps existing commands exactly as documented in `CLAUDE.md`; no divergence from what contributors already know

**Non-Goals:**
- Not a replacement for running commands directly (`go test ./...`, `bash tests/...` remain valid)
- Not a build system for user config files or managed projects — the Makefile is for the monom repo itself
- No target for releasing or publishing (that is a future change)
- No target for generating shell completions for the Makefile itself

## Decisions

### Make over shell script wrappers

**Decision:** Use `make` rather than a standalone shell script (e.g. `scripts/run.sh`).

**Rationale:** `make` is universally available on macOS and Linux, requires no install step, supports dependency ordering between targets, and the `make <target>` invocation pattern is a deeply established convention for Go projects. A shell script wrapper would need its own discoverability mechanism; `make help` is standard.

**Alternative considered:** `Taskfile` (go-task) — rejected because it requires an additional install and the project has no Go toolchain dependency on it. `make` needs no install.

### `##` comment convention for `make help`

**Decision:** Every public target carries a `## <description>` comment on the same line. The `help` target extracts these with `awk`.

**Rationale:** This is the most common self-documenting Makefile pattern. It requires no external tool and produces output consistent with the project's CLI philosophy (help text is plain, column-aligned, and scannable).

### `check` = `test` + shUnit2 e2e + shellcheck

**Decision:** `make check` runs Go tests, all shUnit2 suites under `tests/`, and `shellcheck` on all shell files. This matches the definition in `CLAUDE.md` exactly.

**Rationale:** `CLAUDE.md` already defines `check` this way; the Makefile formalises it. The shUnit2 runner invokes each test file with `bash tests/monomd_*_test` — the glob is safe because `tests/helpers` is not executable.

### `build` produces `bin/monomd`

**Decision:** `make build` runs `go build -o bin/monomd ./cmd/monomd`.

**Rationale:** `bin/monomd` is the canonical output path already referenced in `CLAUDE.md` and the shUnit2 e2e test template (`MONOMD="$REPO_ROOT/bin/monomd"`).

### Makefile is not recursive

**Decision:** A single flat `Makefile` at the repo root. No sub-makes.

**Rationale:** The repo is a single Go module with a small, flat directory structure. Recursive make adds complexity with no benefit at this scale.

## Risks / Trade-offs

- **Risk:** shUnit2 e2e glob `tests/monomd_*_test` picks up unexpected files as new tests are added → Mitigation: the naming convention (`monomd_<subcommand>_test`) is enforced by `CLAUDE.md`; the glob matches only that pattern.
- **Risk:** `make check` depends on `bin/monomd` existing; running it before `make build` will fail → Mitigation: `check` declares `build` as a prerequisite target so `make check` always rebuilds first.
- **Risk:** shellcheck target needs to enumerate all shell files — new files could be missed → Mitigation: use a glob (`src/monom* tests/*`) rather than an explicit list; the Makefile comment notes that new shell files under `src/` are automatically covered.
