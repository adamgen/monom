## ADDED Requirements

### Requirement: MONOM_LIB_ROOT is set on source
When `src/monom` is sourced, it SHALL export `MONOM_LIB_ROOT` as the absolute path to its own containing directory.

#### Scenario: MONOM_LIB_ROOT is set after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `$MONOM_LIB_ROOT` is set to the absolute path of the `src/` directory

### Requirement: setup_monom uses MONOM_PROJECT_ROOT if set
`setup_monom()` SHALL use an already-exported `$MONOM_PROJECT_ROOT` without calling `monomd root`, and SHALL export `MONOM_USER_CONFIG` as `"$MONOM_PROJECT_ROOT/monom"`.

#### Scenario: Pre-set MONOM_PROJECT_ROOT skips discovery
- **WHEN** `$MONOM_PROJECT_ROOT` is already set and points to a valid project directory
- **THEN** `setup_monom` does not call `monomd root` and sets `MONOM_USER_CONFIG="$MONOM_PROJECT_ROOT/monom"`

#### Scenario: setup_monom exports MONOM_USER_CONFIG
- **WHEN** `setup_monom` completes successfully
- **THEN** `$MONOM_USER_CONFIG` is exported and equals `"$MONOM_PROJECT_ROOT/monom"`

### Requirement: setup_monom discovers root via monomd root when MONOM_PROJECT_ROOT is unset
When `$MONOM_PROJECT_ROOT` is not set, `setup_monom()` SHALL call `monomd root` to discover it. On success it SHALL export the result as `$MONOM_PROJECT_ROOT`. On failure it SHALL return non-zero without modifying `$MONOM_USER_CONFIG`.

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
`monom()` SHALL call `setup_monom`, then resolve the command path via `monomd pack "$@"`, and exec the resolved path. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `monomd pack` instead of the original args.

#### Scenario: Command execution without run hook
- **WHEN** `monom deploy` is called and `$MONOM_USER_CONFIG run deploy` produces no output or exits non-zero
- **THEN** `monomd pack deploy` is called and its output is exec'd

#### Scenario: Command execution with run hook
- **WHEN** `monom deploy` is called and `$MONOM_USER_CONFIG run deploy` outputs `infra deploy`
- **THEN** `monomd pack infra deploy` is called and its output is exec'd

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
