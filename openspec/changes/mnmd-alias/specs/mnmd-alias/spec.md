## ADDED Requirements

### Requirement: mnmd alias binds a command name to a project root
`mnmd alias <name> <path>` SHALL bind `<name>` as a top-level command bound to the monom project at `<path>`. It SHALL be handled by the `mnmd()` shell wrapper (not the binary directly), which SHALL both define the alias in the current shell (via `_monom_bind_alias`) and persist it once (via `mnmd alias-save`). Before binding or persisting, `mnmd alias` SHALL verify that `<path>` is a directory containing an executable `monom` config file — the same definition of a valid project root used by `mnmd root`. If `<path>` is not a valid project root, or `<name>` is not a shell-identifier-safe token, `mnmd alias` SHALL print an error, exit non-zero, and neither bind nor persist anything.

#### Scenario: Valid name and root binds and persists
- **WHEN** `mnmd alias myapp /path/to/project` is run and `/path/to/project` contains an executable `monom` file
- **THEN** `myapp` is defined in the current shell and a bind line for it is persisted to the rc file

#### Scenario: Path is not a valid project root
- **WHEN** `mnmd alias myapp /tmp/empty` is run and `/tmp/empty` contains no executable `monom` file
- **THEN** `mnmd alias` prints an error, exits non-zero, defines no `myapp` function, and writes nothing to the rc file

#### Scenario: Invalid alias name is rejected
- **WHEN** `mnmd alias 'bad name' /path/to/project` is run with a name containing characters outside the shell-identifier-safe set
- **THEN** `mnmd alias` prints an error, exits non-zero, and neither binds nor persists anything

### Requirement: An aliased command is usable immediately in the current shell
After `mnmd alias <name> <path>` succeeds, `<name>` SHALL be invocable in the same shell session without re-sourcing any file or opening a new shell, because the `mnmd()` wrapper defines it in-process.

#### Scenario: Alias works right after creation
- **WHEN** `mnmd alias myapp /path/to/project` succeeds and the user then types `myapp` in the same session
- **THEN** `myapp` resolves to the alias function and runs against `/path/to/project`

### Requirement: _monom_bind_alias binds the alias function and completion with a pinned root
`src/monom` SHALL define `_monom_bind_alias <name> <root>`. When called, it SHALL define a shell function `<name>()` that sets `local _MONOM_PROJECT_ROOT="<root>"` and delegates to `_monom_run <name> "$@"`, and it SHALL define a per-name completion function that likewise sets `local _MONOM_PROJECT_ROOT="<root>"` and delegates to the shared completion core, then register that completion function for `<name>` (via `complete -F` in bash, `compdef` in zsh). `<name>` and `<root>` SHALL be substituted into the definitions such that the root is a fixed constant for that alias.

#### Scenario: Bind defines a callable function and completion
- **WHEN** `_monom_bind_alias myapp /path/to/project` is called in bash
- **THEN** `myapp` is defined as a function, a `myapp`-specific completion function is defined, and completion for `myapp` is registered

#### Scenario: Pinned root reaches the execution core
- **WHEN** `myapp deploy` is invoked after binding to `/path/to/project`
- **THEN** `_monom_run` runs with `_MONOM_PROJECT_ROOT` equal to `/path/to/project`

### Requirement: Alias root pinning does not clobber discovery globals
Because `_monom_bind_alias` pins the root with `local`, invoking an alias SHALL NOT leave `_MONOM_PROJECT_ROOT` exported to the surrounding session. A subsequent bare `monom` invocation SHALL still auto-discover its root from `$PWD`, and two aliases bound to different roots SHALL each resolve to their own root regardless of invocation order.

#### Scenario: Bare monom still auto-discovers after an alias call
- **WHEN** the user invokes an alias bound to project A, then runs bare `monom` from inside project B's tree
- **THEN** `monom` auto-discovers project B's root and does not use project A's root

#### Scenario: Two aliases coexist in one session
- **WHEN** aliases `appa` (root A) and `appb` (root B) are both bound and invoked in the same session
- **THEN** `appa` resolves against root A and `appb` resolves against root B

### Requirement: An aliased command behaves identically to monom
Invoking `<name> …` SHALL produce the same behavior as invoking `monom …` from inside the pinned project: the `run` hook is applied when present, commands resolve via `mnmd pack` and are exec'd, command groups and the no-argument case list children via the discovery pipeline, and tab completion works — all rooted at the pinned path. User-facing messages SHALL identify the command by `<name>` rather than `monom`.

#### Scenario: Alias executes a leaf command
- **WHEN** `myapp deploy` is invoked and `deploy` resolves to a leaf in the pinned project
- **THEN** the resolved executable is exec'd, exactly as `monom deploy` would from inside that project

#### Scenario: Alias lists a command group
- **WHEN** `myapp infra` is invoked and `infra` is a command group in the pinned project
- **THEN** the alias prints `myapp: 'infra' is a command group` and the available children to stderr and exits non-zero without exec'ing anything

#### Scenario: Alias with no arguments lists the top level
- **WHEN** `myapp` is invoked with no arguments
- **THEN** it lists the top-level command group of the pinned project, prefixed with `myapp:`

### Requirement: mnmd alias-save persists a define-only bind line idempotently
`mnmd alias-save <name> <path>` (the binary subcommand the `mnmd()` wrapper calls) SHALL append a single line `_monom_bind_alias <name> "<path>"` to the user's rc file if no identical line is already present, using an absolute path. If the line is already present it SHALL make no change and report that the alias is already saved. It SHALL select the rc file by shell and prepend a newline when the file does not end with one, reusing the shared rc machinery. Errors (e.g. unsupported shell) SHALL be reported via coded errors.

#### Scenario: First save appends the bind line
- **WHEN** `mnmd alias-save myapp /path/to/project` runs and the rc file has no matching bind line
- **THEN** `_monom_bind_alias myapp "/path/to/project"` is appended to the rc file and the outcome is reported

#### Scenario: Repeated save is idempotent
- **WHEN** `mnmd alias-save myapp /path/to/project` runs and an identical bind line already exists
- **THEN** the rc file is unchanged and `mnmd alias-save` reports the alias is already saved

#### Scenario: Unsupported shell errors via coded error
- **WHEN** `$SHELL` ends in neither `/zsh` nor `/bash`
- **THEN** `mnmd alias-save` returns a coded error and exits non-zero without modifying any file

### Requirement: Shell startup binds persisted aliases with no subprocess spawns
The persisted rc line SHALL be a direct call to the `_monom_bind_alias` shell function defined by `src/monom`, so that binding a persisted alias at shell startup SHALL NOT spawn the `mnmd` binary or any other subprocess.

#### Scenario: Startup binding spawns no subprocess
- **WHEN** a shell sources an rc file containing `source ".../src/monom"` followed by `_monom_bind_alias myapp "/path/to/project"`
- **THEN** `myapp` is bound with completion and no subprocess is spawned to perform the binding

### Requirement: Invoking an alias whose root is missing fails cleanly
When an alias is invoked but its pinned root no longer exists or no longer contains an executable `monom` config, the alias SHALL print a clear error to stderr and SHALL NOT exec anything.

#### Scenario: Pinned root was removed
- **WHEN** `myapp deploy` is invoked after the pinned project directory has been deleted or its `monom` config removed
- **THEN** the alias prints a clear error to stderr, exits non-zero, and execs nothing
