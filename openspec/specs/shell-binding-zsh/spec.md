## Purpose

Document the zsh completion binding `src/monom.zsh`: the `_monom` completion function, its registration via the guarded `compdef` mechanism, the interactive-safety constraint that it always exits 0 without writing to stderr, how it populates completions from the `mnmd` binary, the skip-`compadd`-when-empty behavior that avoids a spurious trailing space at leaf commands, and the shellcheck cleanliness requirement.

## Requirements

### Requirement: _monom completion function is defined
`src/monom.zsh` SHALL define `_monom()` — the zsh completion function called by the `compdef` mechanism. The `_<command>` naming convention is the zsh standard: zsh's completion system discovers and invokes completion functions by looking for `_<command>` by name. The bash equivalent uses `_monom_completion` because bash has no such convention and uses a descriptive name instead.

#### Scenario: _monom is callable after sourcing
- **WHEN** `src/monom.zsh` is sourced in zsh
- **THEN** `functions[_monom]` is non-empty (the function exists)

### Requirement: zsh completion hook is registered when compdef is available
`src/monom.zsh` SHALL register `_monom` for `monom` using `compdef _monom monom`, guarded by a check that `compdef` is defined. Standard `.zshrc` order has `compinit` run before monom is sourced, so in practice the guard is always true. The guard prevents a visible error if sourced in an unusual environment (e.g. a script or test runner) where `compinit` has not been called.

#### Scenario: compdef registration when compinit has been called
- **WHEN** `src/monom.zsh` is sourced in a zsh session where `compinit` has been called
- **THEN** `_monom` is registered as the completion function for `monom`

#### Scenario: no error when compinit has not been called
- **WHEN** `src/monom.zsh` is sourced before `compinit`
- **THEN** the file sources without error and exits 0 (compdef registration is skipped)

### Requirement: _monom always exits 0 and never writes to stderr
`_monom()` SHALL always exit 0 and SHALL never write anything to stderr, regardless of any internal failure. Tab completion is interactive — a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt. This mirrors the hard constraint on `mnmd filter`.

#### Scenario: empty completions and exit 0 when _setup_monom fails
- **WHEN** `_monom` is invoked and `_setup_monom` returns non-zero (no project root found)
- **THEN** no completions are added, nothing is written to stderr, and `_monom` exits 0

### Requirement: _monom populates completions via the mnmd binary
`_monom()` SHALL call `_setup_monom`, then pass the output of `_monom_cfg complete | mnmd filter "${words[@]:1}"` to `compadd`. `mnmd` here resolves via the wrapper function defined in `src/monom`.

#### Scenario: completions are added on Tab
- **WHEN** `_monom` is invoked with `words=("monom" "")` and `CURRENT=2`
- **THEN** `compadd` is called with the top-level completions returned by `mnmd filter`

### Requirement: _monom skips compadd when there are no completions
When the filter produces no output (e.g. a leaf command has been fully typed, or the filter fails), `_monom()` SHALL return without calling `compadd`. Splitting empty filter output with `${(@f)...}` yields a single empty-string element; calling `compadd -- ""` registers an empty match that zsh "completes" by inserting a trailing space on every Tab press. Skipping `compadd` entirely avoids that spurious space.

#### Scenario: no spurious space at a leaf
- **WHEN** `mnmd filter` exits non-zero or produces no output
- **THEN** `compadd` is not called, no empty match is registered, and `_monom` exits 0

### Requirement: src/monom.zsh passes shellcheck
`src/monom.zsh` SHALL pass `shellcheck` with shell `bash` (as proxy for POSIX-compatible checks on zsh files) and no suppressions except those documented inline with an explanation.

#### Scenario: shellcheck clean
- **WHEN** `shellcheck --shell=bash src/monom.zsh` is run
- **THEN** it exits 0

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
