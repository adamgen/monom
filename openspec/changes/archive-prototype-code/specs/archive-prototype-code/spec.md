## ADDED Requirements

### Requirement: Prototype code is moved to _archive
All prototype-era code files SHALL be moved to a top-level `_archive/` directory while preserving their relative directory structure. No file content SHALL be modified during the move.

#### Scenario: Code directories and scripts are archived
- **WHEN** the archive operation is complete
- **THEN** `_archive/` SHALL contain `src/`, `test_projects/`, `dependencies/`, `build`, `check`, `go_e2e_test`, `sh_test_runner`, `shellcheck`, and `install.sh`, each with their original structure intact

### Requirement: Non-code files remain in place
Markdown documents, OpenSpec artifacts, and `go.mod` SHALL NOT be moved to `_archive/`. They SHALL remain at their current paths in the repository.

#### Scenario: Governance documents and config files are untouched
- **WHEN** the archive operation is complete
- **THEN** `constitution.md`, `architecture.md`, `terminology.md`, `CLAUDE.md`, `README.md`, `old_notes.md`, and `go.mod` SHALL remain at the repo root, and all files under `openspec/` SHALL remain unchanged

### Requirement: Prototype code no longer exists at original paths
After archiving, the original locations SHALL be empty or absent. No prototype code file SHALL remain at its pre-archive path.

#### Scenario: Original code locations are absent from root
- **WHEN** the archive operation is complete
- **THEN** `src/`, `test_projects/`, `dependencies/`, `build`, `check`, `go_e2e_test`, `sh_test_runner`, `shellcheck`, and `install.sh` SHALL NOT exist at the repo root
