## Context

The project has a clearly specified target architecture (`architecture.md`, `constitution.md`) for a new `mnmd` binary with a clean shell/Go separation. However, the current working tree contains a substantial body of prototype-era code that predates and conflicts with this design:

- `src/` — prototype shell scripts (`monom`, `monom_tools`, `monom_usage`, `run`, `completion`, `usage_completion`, `make_monom_alias`, `monom_test`), sh_utils (`get_monom_project_root`, `log`), and an early Go implementation (`main.go`, `go_utils/finder.go`, `go_utils/remvoe_prefix.go`) that uses a `complete` subcommand instead of the intended `filter`/`pack`/`root` model.
- `test_projects/` — two prototype test fixtures (`file_commands`, `monorepo1`) using the old command model.
- `dependencies/shunit2` — the shUnit2 test runner vendored for prototype tests.
- Root-level scripts: `build`, `check`, `go_e2e_test`, `sh_test_runner`, `shellcheck`, `install.sh` — runner and CI scripts tied to the prototype setup.

These files are not being deleted — they may be useful as reference material during new implementation. They need to be moved out of the working tree without being lost.

## Goals / Non-Goals

**Goals:**
- Move all non-markdown, non-OpenSpec code files into `_archive/` at the repo root.
- Preserve full directory structure within `_archive/` so the prototype remains navigable as reference.
- Leave the repo root containing only governance documents, OpenSpec artifacts, and `go.mod`.
- The move is a single atomic git operation (no file content changes).

**Non-Goals:**
- Deleting the prototype code entirely.
- Updating `go.mod` or any import paths (that belongs to the `mnmd-binary` change).
- Creating any new source files or test files.
- Modifying any markdown or OpenSpec documents.

## Decisions

### Decision: Use `_archive/` as the destination folder name

Prefix with `_` so it sorts to the top of directory listings and is visually distinct as "not active code." The name `_archive` is self-documenting. Alternative `legacy/` or `old/` were considered but are less explicit about the folder's role.

### Decision: Preserve the full directory structure inside `_archive/`

Move `src/` → `_archive/src/`, `test_projects/` → `_archive/test_projects/`, etc. This makes the archive immediately navigable and preserves the logical grouping. Alternative of flattening all files was rejected — it would destroy the reference value.

### Decision: Move root-level scripts as a flat group under `_archive/`

The scripts `build`, `check`, `go_e2e_test`, `sh_test_runner`, `shellcheck`, `install.sh` are moved directly to `_archive/` (no subdirectory). They are standalone executables that logically belong at the same level as the prototype code they orchestrate.

### Decision: Keep `go.mod` at the repo root, do not move it

`go.mod` defines the Go module root. It must remain at the repo root for `go` tooling to work. The module path and dependencies will be updated when the new `mnmd` implementation is built. Moving it would break the Go toolchain.

### Decision: Keep `dependencies/` structure intact in the archive

`dependencies/shunit2` is moved to `_archive/dependencies/shunit2`. This preserves the relative path references within the prototype test scripts, keeping the archive self-consistent.

## Risks / Trade-offs

- **`./check` and other root scripts become unavailable** → Mitigation: This is intentional and expected. `CLAUDE.md` references `./check`; a follow-up update to `CLAUDE.md` will note this is pending the `mnmd-binary` implementation. No code is lost.
- **`go.mod` module path `github.com/adamgen/monom/src/go_utils` still references `src/`** → Mitigation: The import path will break if `go build` is run, but that's acceptable. The module is currently not buildable to a clean `mnmd` anyway. This will be corrected in the `mnmd-binary` change.
- **Prototype tests are no longer runnable** → Expected. The prototype test harness is being archived along with the code it tests. This is not a regression — the prototype tests tested prototype behavior.
