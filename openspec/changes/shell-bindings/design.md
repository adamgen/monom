## Context

The monom architecture defines three shell files that have not yet been created:

- `src/monom` — sourced by the user's rc file; defines `monom()`, `setup_monom()`, and detects the shell
- `src/monom.bash` — bash-specific: defines `monom_completion()` and registers `complete -F monom_completion monom`
- `src/monom.zsh` — zsh-specific: defines `_monom()` and registers `compdef _monom monom`

The archive (`_archive/src/monom`, `_archive/src/completion`) contains a working but pre-`monomd` implementation. The new bindings replace all logic with calls to `monomd` subcommands. No logic lives in shell files.

This change is developed in parallel with `tab-completion-testing`, which defines the canonical Go/pty PTY harness for interactive completion assertions. That harness is not yet available; shUnit2 tests cover the non-interactive surface (sourcing, env var setup, function existence). Interactive completion tests against these bindings are a follow-up.

## Goals / Non-Goals

**Goals:**
- Implement `src/monom`, `src/monom.bash`, `src/monom.zsh` exactly as specified in `architecture.md`
- All shell files pass `shellcheck` with no suppressions (except those documented inline)
- shUnit2 e2e tests cover sourcing, `setup_monom()`, `monom()` dispatch, and `monom_cfg` wrapper
- Both bash and zsh completion hooks are registered correctly

**Non-Goals:**
- Interactive tab-completion testing (PTY harness — belongs in `tab-completion-testing`)
- Fish, dash, or other shell support
- `make_monom_alias` / alias feature (deferred — architecture.md notes this is still being determined)
- Completion behavior under edge cases (deferred to `tab-completion-testing`)

## Decisions

### Decision 1: `src/monom` targets bash; `src/monom.bash` and `src/monom.zsh` are shell-specific

monom targets macOS users running bash or zsh (see `architecture.md` — POSIX portability is explicitly out of scope). `src/monom` is written as bash (shebang `#!/usr/bin/env bash`) because that is sufficient for the target audience and avoids POSIX-compat restrictions that complicate `BASH_SOURCE`-based self-location (which has no clean POSIX equivalent).

`src/monom.bash` and `src/monom.zsh` use shell-specific syntax freely since they are only ever sourced by their respective shell.

**Alternative considered:** POSIX sh for `src/monom` — rejected because POSIX has no reliable way to get the path of the currently sourced file (no `BASH_SOURCE`, no `${(%):-%x}`), which is required to set `MONOM_LIB_ROOT`. Using `$0` as a fallback requires a shellcheck suppression and breaks in common sourcing patterns.

**Alternative considered:** A single file with runtime `if [ -n "$ZSH_VERSION" ]` guards — rejected because mixing bash arrays and zsh completion syntax in one file creates shellcheck suppression debt and is harder to reason about independently.

### Decision 2: Shell detection via `$ZSH_VERSION` / `$BASH_VERSION`

`src/monom` detects the active shell by checking `$ZSH_VERSION` first, then `$BASH_VERSION`. If neither is set, it sources no completion file (safe degradation — `monom()` still works, just without tab completion).

**Alternative considered:** `ps` inspection or `$SHELL` — rejected because `$SHELL` reflects the user's login shell, not the currently active shell; a zsh user invoking a bash subshell would get the wrong binding.

### Decision 3: `monom()` is defined in `src/monom` (shell-agnostic), completion hooks in shell-specific files

`monom()` — the function the user calls to run a command — delegates to `monomd pack` and is identical across shells. It belongs in `src/monom`. Completion hooks (`monom_completion` for bash, `_monom` for zsh) differ in their registration API and belong in the shell-specific files.

### Decision 4: `setup_monom()` — purpose and contract

`setup_monom()` is a one-time initialization function called at the start of every `monom()` invocation and every completion request. Its job is to ensure that `$MONOM_PROJECT_ROOT` and `$MONOM_USER_CONFIG` are set before any `monomd` subcommand is called.

What it does, in order:

1. **Short-circuit if already set**: If `$MONOM_PROJECT_ROOT` is already exported (e.g. the user set it explicitly, or a previous call already ran discovery), skip steps 2–3 entirely.
2. **Discover root**: Call `monomd root`, which walks up from `$PWD` looking for a directory containing an executable `monom` file. Capture stdout as the root path.
3. **Export**: Set and export `MONOM_PROJECT_ROOT` (the discovered path) and `MONOM_USER_CONFIG="$MONOM_PROJECT_ROOT/monom"`.
4. **Return non-zero on failure**: If `monomd root` exits non-zero (no project found), `setup_monom` returns non-zero immediately. Each caller handles this differently:
   - `monom()` — prints an error to stderr and returns non-zero. This is a user-facing invocation; feedback is appropriate.
   - Completion handlers (`monom_completion`, `_monom`) — return empty completions with exit 0, and redirect all stderr to `/dev/null` for every command they run. Completion is interactive; a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt (same rationale as `monomd filter`'s hard constraint).

This eliminates all shell-side root-discovery logic — that logic lives in Go (`monomd root`). `setup_monom` is purely plumbing: call Go, capture output, export env vars.

### Decision 5: `monom_cfg` as a function wrapper, not an alias

`monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }` is defined as a function in `src/monom`, consistent with the usage shown in `architecture.md`. This is more portable than an alias and avoids alias-expansion edge cases inside functions.

### Decision 6: `monom()` uses `exec` to invoke the resolved command

`monom()` resolves the command path via `monomd pack` and then runs it with `exec`. In bash, `exec <path>` replaces the current process image — the command runs directly, not as a child process, so its exit code becomes the shell function's exit code and there is no extra process layer. This is analogous to `process.execFile()` in Node.js (not `child_process.spawn`): the calling process is replaced, not forked.

Since `monom()` is a shell function (not a subshell), we need to avoid replacing the user's interactive session. The implementation runs the exec inside a subshell: `( exec "$resolved_path" )`. This way the subshell is replaced by the command, and the user's shell continues normally after the command exits.

Flow:
1. `monom_cfg run "$@"` — attempt the `run` hook; capture output. If it produces non-empty output, use that as the args for step 2. Otherwise use original `"$@"`.
2. `resolved=$(monomd pack $args)` — resolve to an absolute executable path.
3. If `monomd pack` fails (non-zero exit), print an error and return non-zero.
4. `( exec "$resolved" )` — run the command in a subshell via exec.

## Risks / Trade-offs

- [Risk] `monomd root` is called at every `setup_monom()` invocation on the completion path → Mitigation: `$MONOM_PROJECT_ROOT` is exported after first successful call; subsequent invocations short-circuit. This matches existing behavior.
- [Risk] No PTY-based completion tests at merge time → Mitigation: The `tab-completion-testing` change will add these. shUnit2 tests verify all non-interactive behavior. The binding files are minimal enough that shellcheck + functional tests cover the risk surface adequately.
- [Risk] zsh `compdef` fails if `compinit` has not been called → Accepted: monom requires `compinit` to be called before `src/monom` is sourced, which is the standard zsh setup order. This is documented in the install instructions.

### Decision 7: `MONOM_LIB_ROOT` is set via `${BASH_SOURCE[0]}` at source time

`architecture.md` specifies that `MONOM_LIB_ROOT` is set by `src/monom` at source time. Since `src/monom` targets bash, `${BASH_SOURCE[0]}` gives the path of the sourced file even when called via `. /path/to/monom` — `$0` would give the shell name instead. The implementation resolves `${BASH_SOURCE[0]}` to an absolute path using `cd + pwd` before exporting.
