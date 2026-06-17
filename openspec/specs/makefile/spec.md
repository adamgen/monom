### Requirement: help target self-documents available targets
The Makefile SHALL provide a `help` target that prints all public targets with their descriptions, extracted from `##` comments on the same line as the target definition. Output SHALL be column-aligned and human-readable. The `help` target SHALL be the default target (first in the file).

#### Scenario: Running make with no arguments
- **WHEN** the developer runs `make` with no arguments
- **THEN** the `help` target runs and prints all documented targets with descriptions

#### Scenario: Running make help explicitly
- **WHEN** the developer runs `make help`
- **THEN** all targets annotated with `## <description>` are printed, one per line, with the target name left-aligned and the description right-aligned in a second column

---

### Requirement: build target compiles the mnmd binary
The Makefile SHALL provide a `build` target that compiles `cmd/mnmd` and writes the output to `bin/mnmd`. The target SHALL create the `bin/` directory if it does not exist.

#### Scenario: First build on a clean checkout
- **WHEN** the developer runs `make build` and `bin/` does not exist
- **THEN** `bin/` is created and `bin/mnmd` is produced with exit code 0

#### Scenario: Build on a repo with existing binary
- **WHEN** the developer runs `make build` and `bin/mnmd` already exists
- **THEN** the binary is recompiled and the command exits 0

---

### Requirement: test target runs Go unit tests
The Makefile SHALL provide a `test` target that runs `go test ./...` from the repo root.

#### Scenario: All Go tests pass
- **WHEN** the developer runs `make test` and all Go tests pass
- **THEN** the command exits 0 and prints test output to stdout

#### Scenario: A Go test fails
- **WHEN** the developer runs `make test` and at least one test fails
- **THEN** the command exits non-zero

---

### Requirement: test-e2e target runs shUnit2 e2e suites
The Makefile SHALL provide a `test-e2e` target that builds the binary and runs all shUnit2 e2e suites under `tests/`. It SHALL declare `build` as a prerequisite.

#### Scenario: All e2e suites pass
- **WHEN** the developer runs `make test-e2e` and all shUnit2 suites pass
- **THEN** the command exits 0

#### Scenario: An e2e suite fails
- **WHEN** `make test-e2e` is run and a shUnit2 suite reports a failure
- **THEN** the command exits non-zero

---

### Requirement: check target runs all validation gates
The Makefile SHALL provide a `check` target that, in order: builds the binary, runs Go tests, runs all shUnit2 e2e suites under `tests/`, and runs shellcheck on all shell files. `check` SHALL declare `build` as a prerequisite so the binary is always fresh before e2e tests run.

#### Scenario: All checks pass
- **WHEN** the developer runs `make check` and all gates pass
- **THEN** the command exits 0

#### Scenario: Go tests fail during check
- **WHEN** `make check` is run and a Go test fails
- **THEN** the command exits non-zero before running e2e tests

#### Scenario: shellcheck finds an error
- **WHEN** `make check` is run and shellcheck reports a violation in any shell file
- **THEN** the command exits non-zero

---

### Requirement: lint target runs shellcheck on all shell files
The Makefile SHALL provide a `lint` target that runs `shellcheck` on all shell files under `tests/` (and `src/` once shell bindings exist). It SHALL respect `.shellcheckrc` at the repo root.

#### Scenario: No shellcheck violations
- **WHEN** the developer runs `make lint` and all shell files are clean
- **THEN** the command exits 0 with no output

#### Scenario: A shell file has a violation
- **WHEN** `make lint` is run and a shell file contains a shellcheck-flagged pattern
- **THEN** the command exits non-zero and shellcheck prints the violation to stdout

---

### Requirement: clean target removes build artifacts
The Makefile SHALL provide a `clean` target that removes `bin/mnmd` and any other build artifacts. It SHALL be safe to run on a clean checkout (no error if artifacts do not exist).

#### Scenario: Removing an existing binary
- **WHEN** the developer runs `make clean` and `bin/mnmd` exists
- **THEN** `bin/mnmd` is removed and the command exits 0

#### Scenario: Running clean on a clean checkout
- **WHEN** the developer runs `make clean` and no build artifacts exist
- **THEN** the command exits 0 without error

---

### Requirement: targets are PHONY
All Makefile targets (`help`, `build`, `test`, `test-e2e`, `check`, `lint`, `clean`) SHALL be declared `.PHONY` so they always run regardless of files with matching names.

#### Scenario: A file named test exists in the repo root
- **WHEN** a file named `test` exists at the repo root and the developer runs `make test`
- **THEN** the test target still runs (not skipped due to the existing file)
