## MODIFIED Requirements

### Requirement: _mnmd_completion completes the subcommand slot with all known subcommands
`_mnmd_completion()` SHALL populate `COMPREPLY` with the full list of known mnmd subcommands (`filter`, `root`, `pack`, `check`, `install`, `alias`, `completion`) when `COMP_CWORD` is 1 (the subcommand slot) and the current word is empty.

#### Scenario: all subcommands returned with empty prefix
- **WHEN** `_mnmd_completion` is invoked with `COMP_WORDS=(mnmd '')` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` contains all known mnmd subcommands, including `alias`

## ADDED Requirements

### Requirement: _monom_complete is the reusable bash completion core
`src/monom.bash` SHALL define `_monom_complete()` as the reusable bash completion core: it SHALL call `_setup_monom`, then populate `COMPREPLY` from `_monom_cfg complete | mnmd filter "${COMP_WORDS[@]:1}"`, reading the project root from `_MONOM_PROJECT_ROOT` in scope rather than requiring it to be exported globally. It SHALL always exit 0 and never write to stderr. `_monom_completion` (the handler bound to `monom`) SHALL delegate to `_monom_complete`, and each alias's per-name completion function SHALL pin `_MONOM_PROJECT_ROOT` as a `local` and then delegate to `_monom_complete`, so the same core serves `monom` and every alias.

#### Scenario: _monom_complete is callable after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `declare -f _monom_complete` succeeds (the function exists)

#### Scenario: monom handler delegates to the core
- **WHEN** `_monom_completion` is invoked with `COMP_WORDS=("monom" "")` and `COMP_CWORD=1`
- **THEN** it delegates to `_monom_complete`, which populates `COMPREPLY` with the top-level completions from `mnmd filter`

#### Scenario: Core reads the pinned root from scope for an alias
- **WHEN** an alias's per-name completion function sets `local _MONOM_PROJECT_ROOT="/path/to/project"` and calls `_monom_complete`
- **THEN** `COMPREPLY` is populated from the pinned project's `complete` output, and nothing is written to stderr
