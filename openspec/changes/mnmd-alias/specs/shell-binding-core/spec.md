## MODIFIED Requirements

### Requirement: mnmd wrapper function is defined
When `src/monom` is sourced, it SHALL define `mnmd()` as a shell-function wrapper. For every subcommand except `alias`, the wrapper SHALL invoke the `mnmd` binary that ships next to the sources at `$_MONOM_LIB_ROOT/../bin/mnmd`, passing all arguments through. The wrapper is user-facing: after sourcing, the user can invoke `mnmd <subcommand>` by name without `bin/` being on `$PATH`. All binary call sites in `src/monom` and both completion bindings SHALL call `mnmd <subcommand>` via this wrapper. No path variable is exported.

The wrapper SHALL intercept `mnmd alias <name> <path>` and handle it in the shell rather than passing it to the binary, because binding an alias must define a function and register completion in the parent shell — work a subprocess cannot do. On interception it SHALL define the alias in the current shell via `_monom_bind_alias <name> <path>` and persist it by invoking the binary subcommand `mnmd alias-save <name> <path>`.

#### Scenario: mnmd wrapper is callable after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `mnmd <subcommand>` resolves and invokes the binary at `bin/mnmd` relative to the install root

#### Scenario: mnmd remains defined in the user's shell
- **WHEN** a user sources `src/monom`
- **THEN** `mnmd` is defined as a shell function in the interactive session (it is part of the user-facing surface, not an internal name)

#### Scenario: mnmd alias is intercepted by the wrapper
- **WHEN** the user runs `mnmd alias myapp /path/to/project`
- **THEN** the wrapper binds `myapp` in the current shell via `_monom_bind_alias` and persists it via `mnmd alias-save`, rather than passing `alias` to the binary as an ordinary subcommand

#### Scenario: non-alias subcommands still pass through to the binary
- **WHEN** the user runs any `mnmd` subcommand other than `alias` (e.g. `mnmd root`)
- **THEN** the wrapper invokes the binary with the given arguments unchanged

## ADDED Requirements

### Requirement: _monom_run is the shared execution core
`src/monom` SHALL define `_monom_run <name> [args...]` as the shared execution core carrying the full `monom` execution behavior (`run` hook handling, `mnmd pack` resolution and exec, command-group listing, and the no-argument case). `_monom_run` SHALL read the project root from `_MONOM_PROJECT_ROOT` in scope rather than requiring it to be exported globally, and SHALL prefix all user-facing messages with `<name>` instead of a hard-coded `monom`. `monom()` SHALL be defined as `_monom_run monom "$@"`, and each alias function SHALL delegate to `_monom_run <name> "$@"` with the root pinned as a `local`.

#### Scenario: monom delegates to the shared core
- **WHEN** `monom deploy` is invoked
- **THEN** it runs `_monom_run monom deploy`, producing the same behavior as before the core was extracted

#### Scenario: Core parameterizes the message prefix by name
- **WHEN** `_monom_run myapp` is invoked with no arguments against a pinned root
- **THEN** the command-group listing is prefixed with `myapp:` rather than `monom:`

#### Scenario: Core reads the pinned root from scope
- **WHEN** `_monom_run` runs with `_MONOM_PROJECT_ROOT` set as a caller-local variable and not exported
- **THEN** it resolves and dispatches against that root without requiring a global export
