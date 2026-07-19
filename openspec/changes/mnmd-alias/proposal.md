## Why

Developers routinely work across several monom projects, but `monom` auto-discovers a single project root by walking up from `$PWD`, so only one project is drivable at a time and only from inside its tree. There is no way to keep multiple monom CLIs active under distinct command names in one shell session. `mnmd alias` binds a chosen command name to a fixed project root, letting several monom-based CLIs coexist and be invoked from anywhere.

## What Changes

- New `mnmd alias <name> <path>` command: binds `<name>` as a top-level command to the monom project at `<path>`. Invoking `<name> …` behaves **identically** to `monom …` run from inside that project — same discovery, hooks, completion, command packing, and group dispatch — except the project root is pinned to `<path>` instead of discovered from `$PWD`.
- The `mnmd()` shell wrapper intercepts `alias`: it defines the alias in the **current shell immediately** (usable without re-sourcing or a new terminal), then persists it **once** to the rc file via the binary. Every other `mnmd` invocation still falls through to the binary unchanged.
- New internal binary subcommand `mnmd alias-save <name> <path>` (machinery the user never types) that idempotently appends a single define-only bind line to the rc file, reusing the `mnmd install` rc-writing machinery.
- New source-time shell helper `_monom_bind_alias <name> <root>` that binds the alias function plus its completion, pinning the root via a `local` so it never clobbers the global `_MONOM_PROJECT_ROOT` — the mechanism that lets two aliased projects coexist in one session.
- `monom()` and the completion handlers are factored into name-parameterized cores (`_monom_run`, `_monom_complete` / `_monom_complete_zsh`) shared by both `monom` and every alias.
- Aliases persist as one `_monom_bind_alias …` line per alias in the rc file; shell startup binds them with **zero subprocess spawns**.

## Capabilities

### New Capabilities
- `mnmd-alias`: the alias feature end-to-end — the `mnmd alias` command, in-shell binding with immediate availability, the `alias-save` persistence subcommand and its idempotency, root-pinning that avoids clobbering the discovery globals, and behavioral parity of an aliased command with `monom` (execution, `run` hook, command-group dispatch, and tab completion).

### Modified Capabilities
- `shell-binding-core`: the `mnmd()` wrapper now intercepts `alias` (previously a pure pass-through to the binary); `src/monom` additionally defines `_monom_bind_alias` and factors `monom()`'s body into a shared `_monom_run <name>` core that reads the project root from scope.
- `shell-binding-bash`: bash completion is factored into a `_monom_complete` core that reads the project root from scope, so alias completion bindings can reuse it.
- `shell-binding-zsh`: zsh completion is factored into a `_monom_complete_zsh` core, for the same reuse.

## Impact

- **Shell:** `src/monom` (new `_monom_bind_alias`, `_monom_run` core, `mnmd()` interception, `alias` added to the `mnmd` subcommand completion list), `src/monom.bash` (`_monom_complete` core), `src/monom.zsh` (`_monom_complete_zsh` core).
- **Go:** new `internal/alias` package (`alias-save` logic); new `alias-save` case in `cmd/mnmd/main.go`; extraction of a shared `internal/rc` package from `internal/install` (rc-file selection, idempotency check, append) so `install` and `alias-save` share one owner. Coded errors via `internal/cli`.
- **Docs:** `architecture.md` — replace the `make_monom_alias` TBD placeholder with the real design and document `mnmd alias`. `terminology.md` — the `alias` term (added during design).
- **Tests:** Go unit tests for `internal/alias` and the extracted `internal/rc` helpers; shUnit2 e2e `tests/mnmd_alias_test`; a completion e2e for an aliased command.
- **No change** to the constitution-protected required user-config interface (`complete`).
