## RENAMED Requirements

- FROM: `### Requirement: _monom populates completions via the monomd binary`
- TO: `### Requirement: _monom populates completions via the mnmd binary`

## MODIFIED Requirements

### Requirement: _monom always exits 0 and never writes to stderr
`_monom()` SHALL always exit 0 and SHALL never write anything to stderr, regardless of any internal failure. Tab completion is interactive — a non-zero exit or any stderr output mid-typing degrades the experience or corrupts the terminal prompt. This mirrors the hard constraint on `mnmd filter`.

#### Scenario: empty completions and exit 0 when _setup_monom fails
- **WHEN** `_monom` is invoked and `_setup_monom` returns non-zero (no project root found)
- **THEN** no completions are added, nothing is written to stderr, and `_monom` exits 0

### Requirement: _monom populates completions via the mnmd binary
`_monom()` SHALL call `_setup_monom`, then pass the output of `_monom_cfg complete | "$_MONOM_BIN" filter "${words[@]:1}"` to `compadd`. It SHALL invoke the resolved `$_MONOM_BIN` binary rather than the bare name `mnmd` (see the `shell-binding-core` _MONOM_BIN requirement).

#### Scenario: completions are added on Tab
- **WHEN** `_monom` is invoked with `words=("monom" "")` and `CURRENT=2`
- **THEN** `compadd` is called with the top-level completions returned by `mnmd filter`
