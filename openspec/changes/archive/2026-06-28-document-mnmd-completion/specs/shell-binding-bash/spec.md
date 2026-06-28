## ADDED Requirements

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
