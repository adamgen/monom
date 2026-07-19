## MODIFIED Requirements

### Requirement: _mnmd completes the subcommand slot with all known subcommands
`_mnmd()` SHALL call `compadd` with the full list of known mnmd subcommands (`filter`, `root`, `pack`, `check`, `install`, `alias`, `completion`) when `${#words[@]}` is 2 or fewer (i.e. the cursor is in the subcommand slot; `words[1]` is `mnmd`).

#### Scenario: all subcommands offered at the subcommand slot
- **WHEN** `_mnmd` is invoked with `words=(mnmd '')`
- **THEN** `compadd` is called with all known mnmd subcommands, including `alias`

## ADDED Requirements

### Requirement: _monom_complete_zsh is the reusable zsh completion core
`src/monom.zsh` SHALL define `_monom_complete_zsh()` as the reusable zsh completion core: it SHALL call `_setup_monom`, then pass the output of `_monom_cfg complete | mnmd filter "${words[@]:1}"` to `compadd`, reading the project root from `_MONOM_PROJECT_ROOT` in scope rather than requiring it to be exported globally. It SHALL preserve the skip-`compadd`-when-empty behavior (no spurious trailing space at a leaf), always exit 0, and never write to stderr. `_monom` (the function registered for `monom`) SHALL delegate to `_monom_complete_zsh`, and each alias's per-name completion function SHALL pin `_MONOM_PROJECT_ROOT` as a `local` and then delegate to `_monom_complete_zsh`, so the same core serves `monom` and every alias.

#### Scenario: _monom_complete_zsh is callable after sourcing
- **WHEN** `src/monom.zsh` is sourced in zsh
- **THEN** `functions[_monom_complete_zsh]` is non-empty (the function exists)

#### Scenario: monom function delegates to the core
- **WHEN** `_monom` is invoked with `words=("monom" "")` and `CURRENT=2`
- **THEN** it delegates to `_monom_complete_zsh`, which calls `compadd` with the top-level completions from `mnmd filter`

#### Scenario: Core reads the pinned root from scope for an alias
- **WHEN** an alias's per-name completion function sets `local _MONOM_PROJECT_ROOT="/path/to/project"` and calls `_monom_complete_zsh`
- **THEN** `compadd` receives the pinned project's completions, `compadd` is skipped when the filter is empty, and nothing is written to stderr
