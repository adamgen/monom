## Purpose

Document the bash completion binding `src/monom.bash`: the `_monom_completion` handler, its registration via the `complete` builtin, the interactive-safety constraint that it always exits 0 without writing to stderr, how it populates `COMPREPLY` from the `mnmd` binary, and the shellcheck cleanliness requirement.

## Requirements

### Requirement: _monom_completion is defined
`src/monom.bash` SHALL define `_monom_completion()` — the bash completion handler called by the `complete` builtin. The name `_monom_completion` follows bash convention (a descriptive `<command>_completion` pattern, since bash enforces no naming scheme) while carrying the leading-underscore prefix that keeps it out of the `monom<Tab>` completion candidate list. The zsh equivalent uses `_monom` because zsh's completion system identifies completion functions by the `_<command>` prefix convention.

#### Scenario: _monom_completion is callable after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `declare -f _monom_completion` succeeds (the function exists)

### Requirement: bash completion hook is registered
`src/monom.bash` SHALL register `_monom_completion` as the completion function for `monom` using `complete -F _monom_completion monom`.

#### Scenario: complete registration is in place after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `complete -p monom` outputs a line containing `-F _monom_completion`

### Requirement: _monom_completion always exits 0 and never writes to stderr
`_monom_completion()` SHALL always exit 0 and SHALL never write anything to stderr, regardless of any internal failure. Tab completion is interactive — a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt. This mirrors the hard constraint on `mnmd filter`.

#### Scenario: COMPREPLY is empty and exit 0 when _setup_monom fails
- **WHEN** `_monom_completion` is invoked and `_setup_monom` returns non-zero (no project root found)
- **THEN** `COMPREPLY` is empty, nothing is written to stderr, and `_monom_completion` exits 0

### Requirement: _monom_completion populates COMPREPLY via the mnmd binary
`_monom_completion()` SHALL call `_setup_monom`, then populate `COMPREPLY` with the output of `_monom_cfg complete | mnmd filter "${COMP_WORDS[@]:1}"`. `mnmd` here resolves via the wrapper function defined in `src/monom`.

#### Scenario: COMPREPLY is populated on Tab
- **WHEN** `_monom_completion` is invoked with `COMP_WORDS=("monom" "")` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` is set to the top-level completions returned by `mnmd filter`

#### Scenario: COMPREPLY is empty and exit 0 on filter failure
- **WHEN** `mnmd filter` exits non-zero or produces no output
- **THEN** `COMPREPLY` is empty and `_monom_completion` exits 0

### Requirement: src/monom.bash passes shellcheck
`src/monom.bash` SHALL pass `shellcheck` with shell `bash` and no suppressions except those documented inline with an explanation.

#### Scenario: shellcheck clean
- **WHEN** `shellcheck --shell=bash src/monom.bash` is run
- **THEN** it exits 0

### Requirement: _mnmd_completion is defined
`src/monom.bash` SHALL define `_mnmd_completion()` — the bash completion handler for the `mnmd` binary. It follows the same `_<command>_completion` naming convention as `_monom_completion` and carries the leading-underscore prefix to keep it out of completion candidate lists.

#### Scenario: _mnmd_completion is callable after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `declare -f _mnmd_completion` succeeds (the function exists)

### Requirement: bash completion hook for mnmd is registered
`src/monom.bash` SHALL register `_mnmd_completion` as the completion function for `mnmd` using `complete -F _mnmd_completion mnmd`.

#### Scenario: complete registration for mnmd is in place after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `complete -p mnmd` outputs a line containing `-F _mnmd_completion`

### Requirement: _mnmd_completion completes the subcommand slot with all known subcommands
`_mnmd_completion()` SHALL populate `COMPREPLY` with the full list of known mnmd subcommands (`filter`, `root`, `pack`, `check`, `install`, `completion`) when `COMP_CWORD` is 1 (the subcommand slot) and the current word is empty.

#### Scenario: all subcommands returned with empty prefix
- **WHEN** `_mnmd_completion` is invoked with `COMP_WORDS=(mnmd '')` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` contains all known mnmd subcommands

### Requirement: _mnmd_completion filters candidates by the typed prefix
`_mnmd_completion()` SHALL use `compgen -W` to filter the subcommand list against the current word, so only candidates matching the typed prefix are returned.

#### Scenario: prefix filters candidates
- **WHEN** `_mnmd_completion` is invoked with `COMP_WORDS=(mnmd fi)` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` contains `filter` and does not contain `root`

### Requirement: _mnmd_completion produces no completions beyond the subcommand slot
`_mnmd_completion()` SHALL leave `COMPREPLY` empty when `COMP_CWORD` is greater than 1 (i.e. the cursor is positioned after the subcommand argument). mnmd subcommands do not have mnmd-managed sub-arguments that warrant shell completion.

#### Scenario: no completions after subcommand
- **WHEN** `_mnmd_completion` is invoked with `COMP_WORDS=(mnmd filter '')` and `COMP_CWORD=2`
- **THEN** `COMPREPLY` is empty
