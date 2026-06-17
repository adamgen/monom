## Why

Typing `monom` and pressing Tab in an active shell offers `monom`, `monomd`, and a cloud of monom-internal identifiers — `MONOM_BIN`, `MONOM_LIB_ROOT`, `MONOM_PROJECT_ROOT`, `MONOM_USER_CONFIG`, `monom_cfg`, `setup_monom`. The shell completes any function, variable, or PATH executable whose name begins with `monom`/`MONOM`, so every internal name monom exports leaks into the namespace and clutters completion for the user's actual commands. The `monomd` binary is part of this leak: `architecture.md`'s Entry Points section already classifies it as *machinery* reached through `monom`, not a command the user types — yet its `monom`-prefixed name puts it under `monom<Tab>` and contradicts that principle. This change removes every monom-defined identifier — including the binary — from the `monom` prefix, leaving exactly one command in the namespace: `monom`.

## What Changes

- Rename every monom-defined shell function and variable so it begins with `_`, and rename the engine binary out of the `monom` prefix, leaving only the one real command — `monom` — in the namespace:
  - Functions: `setup_monom` → `_setup_monom`, `monom_cfg` → `_monom_cfg`, `monom_completion` → `_monom_completion`. (`_monom` and `_monom_log` are already prefixed.)
  - Variables: `MONOM_BIN` → `_MONOM_BIN`, `MONOM_LIB_ROOT` → `_MONOM_LIB_ROOT`, `MONOM_PROJECT_ROOT` → `_MONOM_PROJECT_ROOT`, `MONOM_USER_CONFIG` → `_MONOM_USER_CONFIG`.
  - Binary: `monomd` → `mnmd`. The Go command dir (`cmd/monomd/` → `cmd/mnmd/`), build output (`bin/monomd` → `bin/mnmd`), resolution in `src/monom` (now `whence -p mnmd` / `type -P mnmd`, fallback `$_MONOM_LIB_ROOT/../bin/mnmd`), and every doc/spec/test reference. Subcommands are unchanged (`mnmd filter`, `mnmd pack`, `mnmd root`, `mnmd check`). A leading `_` is not used here because the engine is a file on disk that may be on PATH — it cannot be "hidden" like a shell identifier; taking it out of the `monom` prefix (`mnmd`) is the achievable goal.
- Reconcile the Entry Points wording: `monom` is the only command a user types; the engine binary is machinery, never invoked directly on the normal path. Fix the `architecture.md` note that currently calls `monomd` one of "two real commands."
- Reframe `MONOM_PROJECT_ROOT` and `MONOM_USER_CONFIG` as what they actually are: an internal calling convention between monom's shell layer (which resolves them) and the `monomd` binary (which reads them). The CLI author never sets or reads them; the public seam is the `monom` config file at the project root exposing `complete`. Because they were never part of the public API, renaming them is **not** a breaking change and requires **no constitution amendment** — the canonical docs are corrected to stop implying they are interface.
- Correct the canonical docs to describe the seam by the artifact, not the variable:
  - `constitution.md`: define the stability contract in terms of the executable `monom` config file and its required `complete` subcommand — name no env var. (This removes the env var name from the constitution, so future renames never touch it.)
  - `terminology.md`: define "monom config file" and "project root" as concepts; mention the env var only as an internal-mechanism footnote.
  - `architecture.md`: keep the Environment Variables table (architecture is where internal mechanism belongs) but frame it as internal shell↔Go plumbing, and call out the one genuinely public affordance — a user MAY pre-set the project root to skip auto-discovery — as a behavior rather than an API surface.
- Keep `MONOM_DEBUG_LOG` unprefixed: it is an opt-in input the *user* exports to enable debug logging, not an identifier monom defines, and it only appears in completion for users who already set it. This exception is documented in design.md.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `shell-binding-core`: Requirements that name `setup_monom`, `monom_cfg`, `MONOM_BIN`, `MONOM_LIB_ROOT`, `MONOM_PROJECT_ROOT`, `MONOM_USER_CONFIG`, and the `monomd` binary change to the underscore-prefixed names and `mnmd`.
- `shell-binding-zsh`: Requirements referencing `setup_monom`, `monom_cfg`, `MONOM_BIN`, and the `monomd` binary change to the underscore-prefixed names and `mnmd`. The `_monom` completion function name is unchanged (already prefixed, and required by zsh's `_<command>` convention).
- `shell-binding-bash`: The `monom_completion` handler is renamed to `_monom_completion`, and references to `setup_monom`, `monom_cfg`, `MONOM_BIN`, and the `monomd` binary change to the underscore-prefixed names and `mnmd`.
- `debug-log`: References to `MONOM_PROJECT_ROOT` / `MONOM_USER_CONFIG` and to `monomd filter` update to the prefixed names and `mnmd`; `MONOM_DEBUG_LOG` itself stays unprefixed. A project-local override for `MONOM_DEBUG_LOG` (a `debug` hook that lets a single project enable logging to its own path) is planned separately in the [`local-debug-log-hook`](../local-debug-log-hook/proposal.md) change, which builds on this one.
- `filter-subcommand`, `pack-subcommand`, `root-subcommand`, `check-subcommand`: the `monomd <subcommand>` invocation name updates to `mnmd <subcommand>` (behavior unchanged).
- `makefile`: build output and command dir update to `bin/mnmd` and `./cmd/mnmd`.

## Impact

- `src/monom` — rename all internal function and variable definitions and call sites; update binary resolution to `mnmd` (and `bin/mnmd` fallback).
- `src/monom.zsh`, `src/monom.bash` — update every reference to the renamed identifiers and the `mnmd` binary.
- `cmd/monomd/` → `cmd/mnmd/` (directory rename via `git mv`; package stays `main`); `bin/monomd` → `bin/mnmd`.
- `Makefile` — `build`/`clean`/`test-e2e`/`lint`/`check` targets: `bin/mnmd`, `./cmd/mnmd`, and the `tests/mnmd_*_test` globs.
- `.gitignore` — `bin/mnmd`.
- `tests/monomd_*_test` → `tests/mnmd_*_test` (rename); `tests/monom_shell_test`, `tests/helpers`, and other shUnit2 tests — update references to the renamed identifiers and binary.
- `internal/install/install.go`, `cmd/monomd/nudge_test.go` — update any embedded `monomd` name string.
- `CLAUDE.md`, `README.md` — update `monomd` references and build instructions.
- `constitution.md` — reword the Pluggability and Required-Interface principles to describe the seam as the `monom` config file exposing `complete`, naming no env var. No amendment needed; the required *interface* (the `complete` subcommand) is unchanged.
- `terminology.md` — define "monom config file" and "project root" by concept; demote the env var to an internal-mechanism note.
- `architecture.md` — keep the Environment Variables table, reframed as internal shell↔Go plumbing; surface "user MAY pre-set the project root" as the one public affordance.
- Go env-var read sites: the binary already reads env vars by name; the names it reads change in lockstep with the shell that sets them. Any `os.Getenv("MONOM_...")` call sites in `internal/` and `cmd/` must be renamed too. The Go module path (`github.com/adamgen/monom`) is unaffected — only the command subdirectory and built binary name change.
- No change to any subcommand's behavior, flags, or I/O contract — `mnmd <sub>` behaves exactly as `monomd <sub>` did. No change to the user config *interface contract* (the required `complete` subcommand): only internal names change.
- **Coordination:** other unapplied changes reference `monomd` (`monom-install`, `monomd-args-subcommand`, `args-shell-helpers`, `managed-projects-scaffolding`, `tab-completion-testing`). They are separate artifacts, not live contract; each will need the same rename when it lands. This change updates only the live specs and code.
