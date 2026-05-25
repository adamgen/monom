## ADDED Requirements

### Requirement: _monom completion function is defined
`src/monom.zsh` SHALL define `_monom()` — the zsh completion function called by the `compdef` mechanism. The `_<command>` naming convention is the zsh standard: zsh's completion system discovers and invokes completion functions by looking for `_<command>` by name. The bash equivalent uses `monom_completion` because bash has no such convention and uses a descriptive name instead.

#### Scenario: _monom is callable after sourcing
- **WHEN** `src/monom.zsh` is sourced in zsh
- **THEN** `functions[_monom]` is non-empty (the function exists)

### Requirement: zsh completion hook is registered
`src/monom.zsh` SHALL register `_monom` for `monom` using `compdef _monom monom`. `src/monom` is sourced from the user's `.zshrc`, where `compinit` is expected to have already been called — this is the standard zsh setup order and the only ordering monom supports.

#### Scenario: compdef registration on sourcing
- **WHEN** `src/monom.zsh` is sourced in a zsh session where `compinit` has been called
- **THEN** `_monom` is registered as the completion function for `monom`

### Requirement: _monom always exits 0 and never writes to stderr
`_monom()` SHALL always exit 0 and SHALL never write anything to stderr, regardless of any internal failure. Tab completion is interactive — a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt. This mirrors the hard constraint on `monomd filter`.

#### Scenario: empty completions and exit 0 when setup_monom fails
- **WHEN** `_monom` is invoked and `setup_monom` returns non-zero (no project root found)
- **THEN** no completions are added, nothing is written to stderr, and `_monom` exits 0

### Requirement: _monom populates completions via monomd filter
`_monom()` SHALL call `setup_monom`, then pass the output of `monom_cfg complete | monomd filter "${words[@]:1}"` to `compadd`.

#### Scenario: completions are added on Tab
- **WHEN** `_monom` is invoked with `words=("monom" "")` and `CURRENT=2`
- **THEN** `compadd` is called with the top-level completions returned by `monomd filter`

#### Scenario: empty completion list on filter failure
- **WHEN** `monomd filter` exits non-zero or produces no output
- **THEN** `compadd` is called with an empty list and `_monom` exits 0

### Requirement: src/monom.zsh passes shellcheck
`src/monom.zsh` SHALL pass `shellcheck` with shell `bash` (as proxy for POSIX-compatible checks on zsh files) and no suppressions except those documented inline with an explanation.

#### Scenario: shellcheck clean
- **WHEN** `shellcheck --shell=bash src/monom.zsh` is run
- **THEN** it exits 0
