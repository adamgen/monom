## Purpose

Document the shared shell entrypoint `src/monom`: the sourcing-time setup that resolves the `monomd` binary, exports the environment the completion bindings and dispatch rely on, defines the `monom` and `monom_cfg` functions, and sources the correct shell-specific completion file based on shell detection.

## Requirements

### Requirement: MONOM_LIB_ROOT is set on source
When `src/monom` is sourced, it SHALL export `MONOM_LIB_ROOT` as the absolute path to its own containing directory.

#### Scenario: MONOM_LIB_ROOT is set after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `$MONOM_LIB_ROOT` is set to the absolute path of the `src/` directory

### Requirement: monomd wrapper resolves the binary at source time
When `src/monom` is sourced, it SHALL set `$_monom_bin` to `$MONOM_LIB_ROOT/../bin/monomd` — the fixed path where the binary always ships relative to the sources — and define a `monomd()` wrapper function (`monomd() { "$_monom_bin" "$@"; }`) so that all call sites (in `src/monom` and both completion bindings) invoke `monomd <subcommand>` while always running the resolved executable.

No `PATH` lookup is performed: `monomd` is never installed as a standalone command or alias, so the sources-relative path is the single source of truth. The wrapper exists so that call sites — which run inside function bodies and completion widgets — invoke the resolved executable directly rather than depending on the bare name resolving in those contexts.

#### Scenario: monomd resolves relative to the sources
- **WHEN** `src/monom` is sourced
- **THEN** `$_monom_bin` is set to `$MONOM_LIB_ROOT/../bin/monomd` and the `monomd()` wrapper invokes that binary from completion and dispatch

### Requirement: setup_monom uses MONOM_PROJECT_ROOT if set
`setup_monom()` SHALL use an already-exported `$MONOM_PROJECT_ROOT` without calling `monomd root`, and SHALL export `MONOM_USER_CONFIG` as `"$MONOM_PROJECT_ROOT/monom"`.

#### Scenario: Pre-set MONOM_PROJECT_ROOT skips discovery
- **WHEN** `$MONOM_PROJECT_ROOT` is already set and points to a valid project directory
- **THEN** `setup_monom` does not call `monomd root` and sets `MONOM_USER_CONFIG="$MONOM_PROJECT_ROOT/monom"`

#### Scenario: setup_monom exports MONOM_USER_CONFIG
- **WHEN** `setup_monom` completes successfully
- **THEN** `$MONOM_USER_CONFIG` is exported and equals `"$MONOM_PROJECT_ROOT/monom"`

### Requirement: setup_monom discovers root via monomd root when MONOM_PROJECT_ROOT is unset
When `$MONOM_PROJECT_ROOT` is not set, `setup_monom()` SHALL call `monomd root` (via the wrapper) to discover it. On success it SHALL export the result as `$MONOM_PROJECT_ROOT`. On failure it SHALL return non-zero without modifying `$MONOM_USER_CONFIG`.

#### Scenario: Discovery succeeds
- **WHEN** `$MONOM_PROJECT_ROOT` is unset and `monomd root` returns a valid path
- **THEN** `setup_monom` exports that path as `$MONOM_PROJECT_ROOT` and sets `MONOM_USER_CONFIG`

#### Scenario: Discovery fails
- **WHEN** `$MONOM_PROJECT_ROOT` is unset and `monomd root` exits non-zero
- **THEN** `setup_monom` exits non-zero and does not set `MONOM_USER_CONFIG`

### Requirement: monom_cfg wraps MONOM_USER_CONFIG
`monom_cfg()` SHALL be defined as `monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }` and SHALL pass all arguments through to `$MONOM_USER_CONFIG`.

#### Scenario: monom_cfg forwards arguments
- **WHEN** `monom_cfg complete` is called after a successful `setup_monom`
- **THEN** it executes `"$MONOM_USER_CONFIG" complete`

### Requirement: monom function dispatches via monomd pack
`monom()` SHALL call `setup_monom`, then resolve the command path via `monomd pack "$@"` (the wrapper), and exec the resolved path. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `monomd pack` instead of the original args.

The `run` hook's exit code SHALL select the behavior:

- **exit 0 with empty stdout** — hook absent or no transform. `monom()` SHALL fall back to `"$@"`. Absent and empty are merged on purpose: a config that omits the `run` arm exits 0 with no output, and the constitution's zero-ceremony hooks principle forbids requiring a sentinel to disambiguate them.
- **exit 0 with non-empty stdout** — the hook transformed the args. `monom()` SHALL use the hook's output.
- **non-zero exit** — hook present and failed. `monom()` SHALL surface the hook's stderr, abort with its exit code, and SHALL NOT fall back or exec. A non-zero exit is an explicit failure the author raised, so surfacing it imposes no ceremony.

The hook's stderr SHALL be captured and forwarded to the user on failure rather than discarded.

The args flow through three parts. Both `monom_cfg run` and `monomd pack` **receive** the args as separate CLI arguments — that input format is identical. The asymmetry is on `run`'s **output**: a hook is a separate process, so it can only emit a flat stdout stream, not an argv array. `monom()` therefore re-splits that stream back into separate args before handing them to `pack`.

The hook may also change the *number* of args — that is its purpose (aliasing, namespace remapping). Below, the hook prepends `custom-folder`, turning 2 args into 3:

```
monom db migrate
  → "$@"  = ["db", "migrate"]                       # separate args
  → monom_cfg run db migrate                        # IN: separate args
        ↳ prints "custom-folder db migrate\n"       # OUT: one flat stream (transformed: 2 → 3 args)
  → (monom re-splits the stream on whitespace)
  → monomd pack custom-folder db migrate            # IN: separate args
        ↳ joins with "/", resolves custom-folder/db/migrate
```

Because the hook can emit a different arg count than it received, `monom()` cannot reuse `"$@"` — it must parse the hook's actual output. The re-split SHALL be done via an array, never a bare unquoted string handed to `pack`: zsh does not word-split unquoted parameters by default (`SH_WORD_SPLIT` off), so `monomd pack $string` would pass `"custom-folder db migrate"` as a single argument and fail to resolve `custom-folder/db/migrate`.

#### Scenario: Command execution without run hook
- **WHEN** `monom deploy` is called and `$MONOM_USER_CONFIG run deploy` exits 0 with no output (hook absent or declined)
- **THEN** `monomd pack deploy` is called and its output is exec'd

#### Scenario: Command execution with run hook
- **WHEN** `monom deploy` is called and `$MONOM_USER_CONFIG run deploy` outputs `infra deploy`
- **THEN** `monomd pack infra deploy` is called and its output is exec'd

#### Scenario: run hook failure aborts and surfaces the error
- **WHEN** `monom deploy` is called and `$MONOM_USER_CONFIG run deploy` exits non-zero
- **THEN** `monom` forwards the hook's stderr, exits with the hook's exit code, and does not call `monomd pack` or exec anything

#### Scenario: Multi-word command preserves separate args in both shells
- **WHEN** `monom db migrate` is called (no `run` hook) in either bash or zsh
- **THEN** `monomd pack` receives `db` and `migrate` as two separate arguments and resolves `db/migrate`, not a single `"db migrate"` argument

#### Scenario: monom exits non-zero when monomd pack fails
- **WHEN** `monomd pack` exits non-zero (command not found)
- **THEN** `monom` exits non-zero without exec'ing anything

### Requirement: Shell detection sources the correct completion file
`src/monom` SHALL detect the active shell by inspecting `$ZSH_VERSION` (checked first) and `$BASH_VERSION`, and SHALL source the matching completion file. monom supports bash and zsh only; these two checks are exhaustive for the target audience.

#### Scenario: zsh detected
- **WHEN** `$ZSH_VERSION` is set at source time
- **THEN** `"$MONOM_LIB_ROOT/monom.zsh"` is sourced

#### Scenario: bash detected
- **WHEN** `$ZSH_VERSION` is unset and `$BASH_VERSION` is set at source time
- **THEN** `"$MONOM_LIB_ROOT/monom.bash"` is sourced
