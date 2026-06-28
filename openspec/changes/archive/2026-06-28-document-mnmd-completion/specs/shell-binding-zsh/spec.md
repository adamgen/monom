## ADDED Requirements

### Requirement: _mnmd completion function is defined
`src/monom.zsh` SHALL define `_mnmd()` — the zsh completion function for the `mnmd` binary. The `_<command>` naming convention follows the zsh standard: zsh's completion system discovers and invokes completion functions by looking for `_<command>` by name.

#### Scenario: _mnmd is callable after sourcing
- **WHEN** `src/monom.zsh` is sourced in zsh
- **THEN** `functions[_mnmd]` is non-empty (the function exists)

### Requirement: zsh completion hook for mnmd is registered when compdef is available
`src/monom.zsh` SHALL register `_mnmd` for `mnmd` using `compdef _mnmd mnmd`, guarded by the same `${+functions[compdef]}` check used for `_monom`. The guard prevents a visible error if sourced before `compinit` has been called.

#### Scenario: compdef registration for mnmd when compinit has been called
- **WHEN** `src/monom.zsh` is sourced in a zsh session where `compinit` has been called
- **THEN** `_mnmd` is registered as the completion function for `mnmd`

#### Scenario: no error when compinit has not been called
- **WHEN** `src/monom.zsh` is sourced before `compinit`
- **THEN** the file sources without error and exits 0 (compdef registration for mnmd is skipped)

### Requirement: _mnmd completes the subcommand slot with all known subcommands
`_mnmd()` SHALL call `compadd` with the full list of known mnmd subcommands (`filter`, `root`, `pack`, `check`, `install`, `completion`) when `${#words[@]}` is 2 or fewer (i.e. the cursor is in the subcommand slot; `words[1]` is `mnmd`).

#### Scenario: all subcommands offered at the subcommand slot
- **WHEN** `_mnmd` is invoked with `words=(mnmd '')`
- **THEN** `compadd` is called with all known mnmd subcommands

### Requirement: _mnmd produces no completions beyond the subcommand slot
`_mnmd()` SHALL not call `compadd` when `${#words[@]}` is greater than 2 (i.e. the cursor is positioned after the subcommand argument).

#### Scenario: no completions after subcommand
- **WHEN** `_mnmd` is invoked with `words=(mnmd filter '')`
- **THEN** `compadd` is not called and no completions are offered
