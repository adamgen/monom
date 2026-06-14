## Purpose

Document the bash completion binding `src/monom.bash`: the `monom_completion` handler, its registration via the `complete` builtin, the interactive-safety constraint that it always exits 0 without writing to stderr, how it populates `COMPREPLY` from the `monomd` binary, and the shellcheck cleanliness requirement.

## Requirements

### Requirement: monom_completion is defined
`src/monom.bash` SHALL define `monom_completion()` — the bash completion handler called by the `complete` builtin. The name `monom_completion` follows bash convention: bash has no enforced naming scheme for completion functions, so a descriptive `<command>_completion` pattern is used. The zsh equivalent uses `_monom` because zsh's completion system identifies completion functions by the `_<command>` prefix convention.

#### Scenario: monom_completion is callable after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `declare -f monom_completion` succeeds (the function exists)

### Requirement: bash completion hook is registered
`src/monom.bash` SHALL register `monom_completion` as the completion function for `monom` using `complete -F monom_completion monom`.

#### Scenario: complete registration is in place after sourcing
- **WHEN** `src/monom.bash` is sourced in bash
- **THEN** `complete -p monom` outputs a line containing `-F monom_completion`

### Requirement: monom_completion always exits 0 and never writes to stderr
`monom_completion()` SHALL always exit 0 and SHALL never write anything to stderr, regardless of any internal failure. Tab completion is interactive — a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt. This mirrors the hard constraint on `monomd filter`.

#### Scenario: COMPREPLY is empty and exit 0 when setup_monom fails
- **WHEN** `monom_completion` is invoked and `setup_monom` returns non-zero (no project root found)
- **THEN** `COMPREPLY` is empty, nothing is written to stderr, and `monom_completion` exits 0

### Requirement: monom_completion populates COMPREPLY via the monomd binary
`monom_completion()` SHALL call `setup_monom`, then populate `COMPREPLY` with the output of `monom_cfg complete | monomd filter "${COMP_WORDS[@]:1}"`. It SHALL invoke the `monomd()` wrapper (which runs the resolved executable) rather than relying on the bare name being on `PATH` (see the `shell-binding-core` monomd wrapper requirement).

#### Scenario: COMPREPLY is populated on Tab
- **WHEN** `monom_completion` is invoked with `COMP_WORDS=("monom" "")` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` is set to the top-level completions returned by `monomd filter`

#### Scenario: COMPREPLY is empty and exit 0 on filter failure
- **WHEN** `monomd filter` exits non-zero or produces no output
- **THEN** `COMPREPLY` is empty and `monom_completion` exits 0

### Requirement: src/monom.bash passes shellcheck
`src/monom.bash` SHALL pass `shellcheck` with shell `bash` and no suppressions except those documented inline with an explanation.

#### Scenario: shellcheck clean
- **WHEN** `shellcheck --shell=bash src/monom.bash` is run
- **THEN** it exits 0
