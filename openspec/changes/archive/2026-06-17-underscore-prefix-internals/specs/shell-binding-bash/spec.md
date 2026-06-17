## RENAMED Requirements

- FROM: `### Requirement: monom_completion is defined`
- TO: `### Requirement: _monom_completion is defined`

- FROM: `### Requirement: monom_completion always exits 0 and never writes to stderr`
- TO: `### Requirement: _monom_completion always exits 0 and never writes to stderr`

- FROM: `### Requirement: monom_completion populates COMPREPLY via the monomd binary`
- TO: `### Requirement: _monom_completion populates COMPREPLY via the mnmd binary`

## MODIFIED Requirements

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
`_monom_completion()` SHALL call `_setup_monom`, then populate `COMPREPLY` with the output of `_monom_cfg complete | "$_MONOM_BIN" filter "${COMP_WORDS[@]:1}"`. It SHALL invoke the resolved `$_MONOM_BIN` binary rather than the bare name `mnmd` (see the `shell-binding-core` _MONOM_BIN requirement).

#### Scenario: COMPREPLY is populated on Tab
- **WHEN** `_monom_completion` is invoked with `COMP_WORDS=("monom" "")` and `COMP_CWORD=1`
- **THEN** `COMPREPLY` is set to the top-level completions returned by `mnmd filter`

#### Scenario: COMPREPLY is empty and exit 0 on filter failure
- **WHEN** `mnmd filter` exits non-zero or produces no output
- **THEN** `COMPREPLY` is empty and `_monom_completion` exits 0
