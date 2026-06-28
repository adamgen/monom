## ADDED Requirements

### Requirement: Managed project detection
The system SHALL recognize a project as "managed" when a `monom.yaml` file exists at the project root. The system SHALL recognize a project as "custom" when an executable `monom` file exists at the project root without a `monom.yaml`.

#### Scenario: Directory contains monom.yaml
- **WHEN** `mnmd root` walks up from `$PWD` and finds a directory containing `monom.yaml`
- **THEN** the system identifies this as a managed project and returns that directory as the project root

#### Scenario: Directory contains executable monom only
- **WHEN** `mnmd root` walks up from `$PWD` and finds a directory containing an executable `monom` file but no `monom.yaml`
- **THEN** the system identifies this as a custom project and returns that directory as the project root

#### Scenario: Directory contains both monom.yaml and executable monom
- **WHEN** a directory contains both `monom.yaml` and an executable `monom` file
- **THEN** the system treats the project as managed (`monom.yaml` takes precedence)

### Requirement: monom.yaml minimal schema
The `monom.yaml` file SHALL have `commands` as its only required property. The `commands` property specifies the relative path to the directory containing command scripts.

#### Scenario: Minimal valid monom.yaml
- **WHEN** `monom.yaml` contains only `commands: ./commands`
- **THEN** the system uses `./commands` relative to the project root as the command tree root

#### Scenario: monom.yaml missing commands property
- **WHEN** `monom.yaml` exists but does not contain a `commands` property
- **THEN** the system exits with a non-zero exit code and prints an error explaining that `commands` is required

### Requirement: monom.yaml optional default_language property
The `monom.yaml` file MAY contain a `default_language` property specifying the default language for scaffolded command scripts. Valid values are `bash`, `python`, and `node`.

#### Scenario: default_language set to python
- **WHEN** `monom.yaml` contains `default_language: python`
- **THEN** scaffolding commands generate python scripts by default (with `#!/usr/bin/env python3` shebang)

#### Scenario: default_language not set
- **WHEN** `monom.yaml` does not contain `default_language`
- **THEN** scaffolding commands default to bash scripts (with `#!/bin/bash` shebang)

### Requirement: monom.yaml script delegation
The `monom.yaml` file MAY contain `run` and/or `complete` properties pointing to executable scripts. When present, `mnmd` SHALL delegate that operation to the specified script instead of handling it internally.

#### Scenario: Delegated run operation
- **WHEN** `monom.yaml` contains `run: ./monom-run`
- **THEN** `mnmd run <args>` calls `./monom-run run <args>` instead of resolving the path internally

#### Scenario: Delegated complete operation
- **WHEN** `monom.yaml` contains `complete: ./monom-complete`
- **THEN** `mnmd complete` calls `./monom-complete complete` instead of walking the tree internally

#### Scenario: No delegation specified
- **WHEN** `monom.yaml` does not contain `run` or `complete` properties
- **THEN** `mnmd` handles discovery and resolution internally by walking the `commands` directory

### Requirement: Internal discovery for managed projects
For managed projects without delegated `complete`, `mnmd` SHALL walk the `commands` directory tree and produce command paths. Filenames become command names, directory names become categories.

#### Scenario: Flat command directory
- **WHEN** the commands directory contains files `deploy`, `test`, `build`
- **THEN** `mnmd complete` outputs `deploy`, `test`, `build` (one per line)

#### Scenario: Nested command directory
- **WHEN** the commands directory contains `infra/provision`, `infra/destroy`, `deploy`
- **THEN** `mnmd complete` outputs `infra/provision`, `infra/destroy`, `deploy` (one per line)

#### Scenario: Non-executable files are excluded
- **WHEN** the commands directory contains a file without executable permission (e.g., a README)
- **THEN** that file does not appear in `mnmd complete` output

### Requirement: Internal resolution for managed projects
For managed projects without delegated `run`, `mnmd` SHALL resolve command arguments to the corresponding file path under the `commands` directory.

#### Scenario: Simple command resolution
- **WHEN** user runs `mnmd run deploy` in a managed project with `commands: ./commands`
- **THEN** `mnmd` prints the absolute path to `./commands/deploy`

#### Scenario: Nested command resolution
- **WHEN** user runs `mnmd run infra provision` in a managed project
- **THEN** `mnmd` prints the absolute path to `./commands/infra/provision`

#### Scenario: Command not found
- **WHEN** user runs `mnmd run nonexistent` and no matching file exists
- **THEN** `mnmd` exits with non-zero status and prints an error message

### Requirement: Managed project data flow reduces subprocess roundtrips
For managed projects, tab completion and command execution SHALL each require only one `mnmd` subprocess call (plus the final exec for command execution).

#### Scenario: Tab completion in managed project
- **WHEN** the user presses Tab in a managed project
- **THEN** the shell calls `mnmd complete <prefix>` (one subprocess) and receives filtered completions

#### Scenario: Command execution in managed project
- **WHEN** the user executes a command in a managed project
- **THEN** the shell calls `mnmd run <args>` (one subprocess), receives the path, and exec's it

### Requirement: Constitutional terminology
The terms "managed project" and "custom project" SHALL be added to `terminology.md` with precise definitions. "Managed project" means a monom project using `monom.yaml` where mnmd handles discovery and resolution. "Custom project" means a monom project using an executable `monom` config file implementing the `complete`/`run` interface.

#### Scenario: Terms used consistently
- **WHEN** any documentation, code comment, or error message refers to these project types
- **THEN** it uses "managed project" or "custom project" exactly as defined in terminology.md
