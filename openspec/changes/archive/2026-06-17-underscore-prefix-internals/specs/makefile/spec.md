## RENAMED Requirements

- FROM: `### Requirement: build target compiles the monomd binary`
- TO: `### Requirement: build target compiles the mnmd binary`

## MODIFIED Requirements

### Requirement: build target compiles the mnmd binary
The Makefile SHALL provide a `build` target that compiles `cmd/mnmd` and writes the output to `bin/mnmd`. The target SHALL create the `bin/` directory if it does not exist.

#### Scenario: First build on a clean checkout
- **WHEN** the developer runs `make build` and `bin/` does not exist
- **THEN** `bin/` is created and `bin/mnmd` is produced with exit code 0

#### Scenario: Build on a repo with existing binary
- **WHEN** the developer runs `make build` and `bin/mnmd` already exists
- **THEN** the binary is recompiled and the command exits 0

### Requirement: clean target removes build artifacts
The Makefile SHALL provide a `clean` target that removes `bin/mnmd` and any other build artifacts. It SHALL be safe to run on a clean checkout (no error if artifacts do not exist).

#### Scenario: Removing an existing binary
- **WHEN** the developer runs `make clean` and `bin/mnmd` exists
- **THEN** `bin/mnmd` is removed and the command exits 0

#### Scenario: Running clean on a clean checkout
- **WHEN** the developer runs `make clean` and no build artifacts exist
- **THEN** the command exits 0 without error
