## RENAMED Requirements

- FROM: `### Requirement: MONOM_LIB_ROOT is set on source`
- TO: `### Requirement: _MONOM_LIB_ROOT is set on source`

- FROM: `### Requirement: MONOM_BIN resolves the monomd binary at source time`
- TO: `### Requirement: _MONOM_BIN resolves the mnmd binary at source time`

- FROM: `### Requirement: setup_monom uses MONOM_PROJECT_ROOT if set`
- TO: `### Requirement: _setup_monom uses _MONOM_PROJECT_ROOT if set`

- FROM: `### Requirement: setup_monom discovers root via monomd root when MONOM_PROJECT_ROOT is unset`
- TO: `### Requirement: _setup_monom discovers root via mnmd root when _MONOM_PROJECT_ROOT is unset`

- FROM: `### Requirement: monom_cfg wraps MONOM_USER_CONFIG`
- TO: `### Requirement: _monom_cfg wraps _MONOM_USER_CONFIG`

- FROM: `### Requirement: monom function dispatches via monomd pack`
- TO: `### Requirement: monom function dispatches via mnmd pack`

## MODIFIED Requirements

### Requirement: _MONOM_LIB_ROOT is set on source
When `src/monom` is sourced, it SHALL export `_MONOM_LIB_ROOT` as the absolute path to its own containing directory.

#### Scenario: _MONOM_LIB_ROOT is set after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `$_MONOM_LIB_ROOT` is set to the absolute path of the `src/` directory

### Requirement: _MONOM_BIN resolves the mnmd binary at source time
When `src/monom` is sourced, it SHALL resolve the `mnmd` binary to an explicit executable path, export it as `$_MONOM_BIN`, and all call sites (in `src/monom` and both completion bindings) SHALL invoke `"$_MONOM_BIN"` rather than the bare name `mnmd`. Resolution SHALL prefer a real `mnmd` executable found on `PATH` ignoring aliases and functions (`whence -p` in zsh, `type -P` in bash), and SHALL fall back to `$_MONOM_LIB_ROOT/../bin/mnmd` when no PATH binary is found.

This is required because a user may have `mnmd` defined only as a shell *alias*. Aliases are not expanded inside function bodies or completion widgets, so a bare `mnmd` call there resolves to nothing and fails with "command not found" (exit 127) — silently, since call sites discard stderr. A `mnmd()` wrapper function is deliberately NOT used: zsh refuses to define a function whose name collides with an existing alias (parse error).

#### Scenario: mnmd available only as an alias, not on PATH
- **WHEN** `src/monom` is sourced in a shell where `mnmd` exists only as an alias and no `mnmd` binary is on `PATH`
- **THEN** `$_MONOM_BIN` is set to `$_MONOM_LIB_ROOT/../bin/mnmd` and completion and dispatch invoke that binary successfully

#### Scenario: mnmd binary on PATH
- **WHEN** `src/monom` is sourced in a shell where a real `mnmd` executable is on `PATH`
- **THEN** `$_MONOM_BIN` is set to that executable's path

### Requirement: _setup_monom uses _MONOM_PROJECT_ROOT if set
`_setup_monom()` SHALL use an already-exported `$_MONOM_PROJECT_ROOT` without calling `mnmd root`, and SHALL export `_MONOM_USER_CONFIG` as `"$_MONOM_PROJECT_ROOT/monom"`.

#### Scenario: Pre-set _MONOM_PROJECT_ROOT skips discovery
- **WHEN** `$_MONOM_PROJECT_ROOT` is already set and points to a valid project directory
- **THEN** `_setup_monom` does not call `mnmd root` and sets `_MONOM_USER_CONFIG="$_MONOM_PROJECT_ROOT/monom"`

#### Scenario: _setup_monom exports _MONOM_USER_CONFIG
- **WHEN** `_setup_monom` completes successfully
- **THEN** `$_MONOM_USER_CONFIG` is exported and equals `"$_MONOM_PROJECT_ROOT/monom"`

### Requirement: _setup_monom discovers root via mnmd root when _MONOM_PROJECT_ROOT is unset
When `$_MONOM_PROJECT_ROOT` is not set, `_setup_monom()` SHALL call `"$_MONOM_BIN" root` to discover it. On success it SHALL export the result as `$_MONOM_PROJECT_ROOT`. On failure it SHALL return non-zero without modifying `$_MONOM_USER_CONFIG`.

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
`monom()` SHALL call `_setup_monom`, then resolve the command path via `"$_MONOM_BIN" pack "$@"`, and exec the resolved path. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `"$_MONOM_BIN" pack` instead of the original args.

#### Scenario: Command execution without run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` produces no output or exits non-zero
- **THEN** `mnmd pack deploy` is called and its output is exec'd

#### Scenario: Command execution with run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` outputs `infra deploy`
- **THEN** `mnmd pack infra deploy` is called and its output is exec'd

#### Scenario: monom exits non-zero when mnmd pack fails
- **WHEN** `mnmd pack` exits non-zero (command not found)
- **THEN** `monom` exits non-zero without exec'ing anything
