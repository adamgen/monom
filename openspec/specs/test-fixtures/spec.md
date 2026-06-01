## Purpose

Document the `fixtures/` directory: a set of self-contained example monom projects used as test fixtures and as worked reference examples for contributors. The initial fixture, `demo-project`, models a canonical monom project so tests can exercise discovery, packing, and completion against a realistic tree.

## Requirements

### Requirement: fixtures directory holds reusable example projects
The repository SHALL maintain a `fixtures/` directory at the repo root containing self-contained example monom projects used by tests and documentation. Each subdirectory SHALL be a complete, runnable project that demonstrates the canonical monom layout without requiring any external setup.

#### Scenario: A test needs a realistic project to exercise the binary
- **WHEN** a Go unit test or shUnit2 e2e test needs a realistic project tree to exercise discovery, packing, or completion
- **THEN** it points `MONOM_PROJECT_ROOT` (and, where relevant, `MONOM_USER_CONFIG`) at a fixture under `fixtures/` rather than constructing an ad-hoc tree inline

#### Scenario: A contributor wants a reference example
- **WHEN** a contributor wants to see what a valid monom project looks like end to end
- **THEN** they can read a fixture under `fixtures/` as a worked example

---

### Requirement: demo-project models a canonical monom project
The `fixtures/demo-project/` directory SHALL be a representative example monom project. It SHALL contain a `monom` config file at its root and a tree of executable command scripts organized into folders, so that the file tree itself forms the command tree (folders are command categories, scripts are commands).

#### Scenario: Project root contains the config file
- **WHEN** `fixtures/demo-project/` is inspected
- **THEN** an executable `monom` config file exists at its root, making `fixtures/demo-project/` the project root

#### Scenario: File tree maps to a command tree
- **WHEN** the command scripts under `fixtures/demo-project/` are listed
- **THEN** nested folders (e.g. `infra/cloud`, `infra/local`, `db`) act as command categories and the executable scripts within them act as commands

---

### Requirement: demo-project config implements discovery via complete
The `monom` config file in `fixtures/demo-project/` SHALL implement the `complete` subcommand, printing one command path per line, with paths relative to the project root and using `/` as the separator. The set of printed paths SHALL correspond to the executable command scripts present in the project tree.

#### Scenario: complete lists every command path
- **WHEN** `fixtures/demo-project/monom complete` is run
- **THEN** it prints each command path on its own line, one per executable command in the tree (e.g. `infra/cloud/deploy`, `infra/cloud/teardown`, `infra/local/start`, `infra/local/stop`, `db/migrate`, `db/seed`, `release`)

#### Scenario: Unknown subcommand is a no-op
- **WHEN** the config file is invoked with a subcommand it does not implement
- **THEN** it produces no output and exits without error

---

### Requirement: demo-project command scripts are executable and self-contained
Every command script under `fixtures/demo-project/` SHALL be an executable file with a `#!/usr/bin/env bash` shebang that performs a trivial, side-effect-free action (such as echoing a status line). Scripts SHALL NOT depend on network access, external services, or installed tools beyond a POSIX shell.

#### Scenario: A command script runs in isolation
- **WHEN** any command script under `fixtures/demo-project/` is executed directly
- **THEN** it exits 0 and prints a short status line without contacting any external service

#### Scenario: Scripts carry the executable bit
- **WHEN** the command scripts and the `monom` config file are inspected
- **THEN** each has its executable bit set so monom can discover and run them
