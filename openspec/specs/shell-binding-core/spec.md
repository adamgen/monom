## Purpose

Document the shared shell entrypoint `src/monom`: the sourcing-time setup that locates the `mnmd` binary, exports the environment the completion bindings and dispatch rely on, defines the `monom` and `_monom_cfg` functions, and sources the correct shell-specific completion file based on shell detection.

## Requirements

### Requirement: _MONOM_LIB_ROOT is set on source
When `src/monom` is sourced, it SHALL export `_MONOM_LIB_ROOT` as the absolute path to its own containing directory.

#### Scenario: _MONOM_LIB_ROOT is set after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `$_MONOM_LIB_ROOT` is set to the absolute path of the `src/` directory

### Requirement: mnmd wrapper function is defined
When `src/monom` is sourced, it SHALL define `mnmd()` as a thin wrapper that invokes the `mnmd` binary that ships next to the sources at `$_MONOM_LIB_ROOT/../bin/mnmd`. All call sites in `src/monom` and both completion bindings SHALL call `mnmd <subcommand>` via this wrapper. No path variable is exported.

#### Scenario: mnmd wrapper is callable after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `mnmd <subcommand>` resolves and invokes the binary at `bin/mnmd` relative to the install root

### Requirement: _setup_monom uses _MONOM_PROJECT_ROOT if set
`_setup_monom()` SHALL use an already-exported `$_MONOM_PROJECT_ROOT` without calling `mnmd root`, and SHALL export `_MONOM_USER_CONFIG` as `"$_MONOM_PROJECT_ROOT/monom"`.

#### Scenario: Pre-set _MONOM_PROJECT_ROOT skips discovery
- **WHEN** `$_MONOM_PROJECT_ROOT` is already set and points to a valid project directory
- **THEN** `_setup_monom` does not call `mnmd root` and sets `_MONOM_USER_CONFIG="$_MONOM_PROJECT_ROOT/monom"`

#### Scenario: _setup_monom exports _MONOM_USER_CONFIG
- **WHEN** `_setup_monom` completes successfully
- **THEN** `$_MONOM_USER_CONFIG` is exported and equals `"$_MONOM_PROJECT_ROOT/monom"`

### Requirement: _setup_monom discovers root via mnmd root when _MONOM_PROJECT_ROOT is unset
When `$_MONOM_PROJECT_ROOT` is not set, `_setup_monom()` SHALL call `mnmd root` to discover it. On success it SHALL export the result as `$_MONOM_PROJECT_ROOT`. On failure it SHALL return non-zero without modifying `$_MONOM_USER_CONFIG`.

#### Scenario: Discovery succeeds
- **WHEN** `$_MONOM_PROJECT_ROOT` is unset and `mnmd root` returns a valid path
- **THEN** `_setup_monom` exports that path as `$_MONOM_PROJECT_ROOT` and sets `_MONOM_USER_CONFIG`

#### Scenario: Discovery fails
- **WHEN** `$_MONOM_PROJECT_ROOT` is unset and `mnmd root` exits non-zero
- **THEN** `_setup_monom` exits non-zero and does not set `_MONOM_USER_CONFIG`

### Requirement: _monom_cfg wraps _MONOM_USER_CONFIG
`_monom_cfg()` SHALL be defined as `_monom_cfg() { "$_MONOM_USER_CONFIG" "$@"; }` and SHALL pass all arguments through to `$_MONOM_USER_CONFIG`.

#### Scenario: _monom_cfg forwards arguments
- **WHEN** `_monom_cfg complete` is called after a successful `_setup_monom`
- **THEN** it executes `"$_MONOM_USER_CONFIG" complete`

### Requirement: monom function dispatches via mnmd pack
`monom()` SHALL call `_setup_monom`, then resolve the command path via `mnmd pack "$@"`, and exec the resolved path. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `mnmd pack` instead of the original args.

The `run` hook's exit code SHALL select the behavior:

- **exit 0 with empty stdout** — hook absent or no transform. `monom()` SHALL fall back to `"$@"`. Absent and empty are merged on purpose: a config that omits the `run` arm exits 0 with no output, and the constitution's zero-ceremony hooks principle forbids requiring a sentinel to disambiguate them.
- **exit 0 with non-empty stdout** — the hook transformed the args. `monom()` SHALL use the hook's output.
- **non-zero exit** — hook present and failed. `monom()` SHALL surface the hook's stderr, abort with its exit code, and SHALL NOT fall back or exec. A non-zero exit is an explicit failure the author raised, so surfacing it imposes no ceremony.

The hook's stderr SHALL be captured and forwarded to the user on failure rather than discarded.

The args flow through three parts. Both `_monom_cfg run` and `mnmd pack` **receive** the args as separate CLI arguments — that input format is identical. The asymmetry is on `run`'s **output**: a hook is a separate process, so it can only emit a flat stdout stream, not an argv array. `monom()` therefore re-splits that stream back into separate args before handing them to `pack`.

The hook may also change the *number* of args — that is its purpose (aliasing, namespace remapping). Below, the hook prepends `custom-folder`, turning 2 args into 3:

```
monom db migrate
  → "$@"  = ["db", "migrate"]                       # separate args
  → _monom_cfg run db migrate                       # IN: separate args
        ↳ prints "custom-folder db migrate\n"       # OUT: one flat stream (transformed: 2 → 3 args)
  → (monom re-splits the stream on whitespace)
  → mnmd pack custom-folder db migrate              # IN: separate args
        ↳ joins with "/", resolves custom-folder/db/migrate
```

Because the hook can emit a different arg count than it received, `monom()` cannot reuse `"$@"` — it must parse the hook's actual output. The re-split SHALL be done via an array, never a bare unquoted string handed to `pack`: zsh does not word-split unquoted parameters by default (`SH_WORD_SPLIT` off), so `mnmd pack $string` would pass `"custom-folder db migrate"` as a single argument and fail to resolve `custom-folder/db/migrate`.

#### Scenario: Command execution without run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` exits 0 with no output (hook absent or declined)
- **THEN** `mnmd pack deploy` is called and its output is exec'd

#### Scenario: Command execution with run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` outputs `infra deploy`
- **THEN** `mnmd pack infra deploy` is called and its output is exec'd

#### Scenario: run hook failure aborts and surfaces the error
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` exits non-zero
- **THEN** `monom` forwards the hook's stderr, exits with the hook's exit code, and does not call `mnmd pack` or exec anything

#### Scenario: Multi-word command preserves separate args in both shells
- **WHEN** `monom db migrate` is called (no `run` hook) in either bash or zsh
- **THEN** `mnmd pack` receives `db` and `migrate` as two separate arguments and resolves `db/migrate`, not a single `"db migrate"` argument

#### Scenario: monom exits non-zero when mnmd pack fails
- **WHEN** `mnmd pack` exits non-zero (command not found)
- **THEN** `monom` exits non-zero without exec'ing anything

### Requirement: Shell detection sources the correct completion file
`src/monom` SHALL detect the active shell by inspecting `$ZSH_VERSION` (checked first) and `$BASH_VERSION`, and SHALL source the matching completion file. monom supports bash and zsh only; these two checks are exhaustive for the target audience.

#### Scenario: zsh detected
- **WHEN** `$ZSH_VERSION` is set at source time
- **THEN** `"$_MONOM_LIB_ROOT/monom.zsh"` is sourced

#### Scenario: bash detected
- **WHEN** `$ZSH_VERSION` is unset and `$BASH_VERSION` is set at source time
- **THEN** `"$_MONOM_LIB_ROOT/monom.bash"` is sourced
