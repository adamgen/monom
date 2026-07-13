## ADDED Requirements

### Requirement: CodedError interface

`internal/cli` SHALL export a `CodedError` interface that extends `error` with an `ExitCode() int` method. An embeddable `Base` struct SHALL provide a default `ExitCode()` implementation so concrete error types get the method by embedding it.

#### Scenario: Concrete error type satisfies CodedError

- **WHEN** a concrete error type embeds `cli.Base` with a code value
- **THEN** the type satisfies the `CodedError` interface and `ExitCode()` returns the embedded code

### Requirement: Central exit-code registry

`internal/cli` SHALL export a single registry struct containing all exit codes used by `mnmd`. The registry MUST define at minimum:
- `Success` (0) — leaf resolved / normal output
- `Error` (1) — generic real error
- `GroupError` (3) — pack command-group signal, payload-free

No exit-code integer literal SHALL appear outside the registry. All `CodedError` constructors and the dispatch tail in `main.go` SHALL reference the registry.

#### Scenario: Registry is the sole source of exit-code values

- **WHEN** a contributor searches the codebase for exit-code integer literals (0, 1, 3 used as exit codes)
- **THEN** they appear only in the registry definition in `internal/cli`, not in `main.go` or `internal/*` subcommand packages
- **AND** `architecture.md` references `internal/cli/cli.go` instead of duplicating the values inline

#### Scenario: Filter exit 0 is exempt

- **WHEN** `runFilter` completes
- **THEN** it exits 0 directly — it does not participate in the `CodedError` dispatch

### Requirement: Generic error wrapper

`internal/cli` SHALL export a constructor that wraps any `error` as a `CodedError` using the registry's `Error` code (1).

#### Scenario: Plain error wrapped as CodedError

- **WHEN** a subcommand returns a plain `fmt.Errorf(...)` and `main.go` wraps it
- **THEN** the resulting `CodedError` has `ExitCode() == 1` and its `Error()` returns the original message

### Requirement: GroupError carries its own exit code

`pack.GroupError` SHALL embed `cli.Base` and derive its exit code from the registry's `GroupError` entry (3). `main.go` SHALL resolve the exit code via the `CodedError` interface, not a type-specific `errors.As` check.

#### Scenario: Pack returns a command group

- **WHEN** `mnmd pack` resolves tokens to a directory
- **THEN** `pack.Pack` returns a `*GroupError` that satisfies `CodedError` with `ExitCode() == 3`
- **AND** the binary exits 3 with empty stdout and empty stderr

### Requirement: Uniform dispatch tail in main.go

`main.go` SHALL have one error-handling tail after each `runX` call. The tail SHALL:
1. Check if the error satisfies `CodedError` via `errors.As`.
2. If yes, use `ce.ExitCode()` as the exit code.
3. If `ExitCode()` equals the registry's `GroupError` code, suppress stderr output.
4. Otherwise, print `mnmd <subcommand>: <message>` to stderr.
5. If the error does not satisfy `CodedError`, default to the registry's `Error` code (1).

#### Scenario: Subcommand returns a CodedError

- **WHEN** a `runX` function returns a `CodedError` with code 1
- **THEN** the binary prints the error message to stderr prefixed with `mnmd <subcommand>:` and exits 1

#### Scenario: Subcommand returns a plain error

- **WHEN** a `runX` function returns a plain `error` that does not satisfy `CodedError`
- **THEN** the binary prints the error message to stderr and exits with the registry's `Error` code (1)

#### Scenario: GroupError suppresses stderr

- **WHEN** a `runX` function returns a `CodedError` with `ExitCode() == ExitCodes.GroupError`
- **THEN** the binary writes nothing to stdout or stderr and exits 3

### Requirement: Constitution principle

`constitution.md` SHALL contain a new principle: *An error carries its own exit code.* The principle SHALL state that a subcommand's outcome is determined by the typed error it returns, not by its call site in `main.go`.

#### Scenario: Principle documented

- **WHEN** a contributor reads `constitution.md`
- **THEN** they find a principle stating errors carry their own exit codes
