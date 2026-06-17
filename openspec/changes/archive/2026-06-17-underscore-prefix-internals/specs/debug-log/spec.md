## RENAMED Requirements

- FROM: `### Requirement: monomd filter logs its inputs and outputs`
- TO: `### Requirement: mnmd filter logs its inputs and outputs`

## MODIFIED Requirements

### Requirement: Shell completion handler logs its full pipeline
The shell completion handlers (`_monom_completion` in bash, `_monom` in zsh) SHALL log the following when `MONOM_DEBUG_LOG` is set:
1. Entry — the words received from the shell (`COMP_WORDS` in bash, `words` in zsh)
2. `_setup_monom` outcome — success with discovered root, or failure
3. `_monom_cfg complete` output — number of lines returned
4. `mnmd filter` output — the final completions (or empty)

`MONOM_DEBUG_LOG` itself is NOT renamed: it is an opt-in input the user exports to enable logging, not a monom-defined internal identifier, so it keeps its public, unprefixed name.

#### Scenario: Entry line includes the typed words
- **WHEN** the user types `monom dep<Tab>` and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording the words passed to the completion handler

#### Scenario: _setup_monom failure is logged
- **WHEN** `_setup_monom` returns non-zero and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line indicating `_setup_monom` failed

#### Scenario: _monom_cfg complete output count is logged
- **WHEN** `_monom_cfg complete` returns N paths and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording the count N

#### Scenario: Final completions are logged
- **WHEN** `mnmd filter` produces completions and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording those completions

### Requirement: mnmd filter logs its inputs and outputs
`mnmd filter` SHALL log the following when `MONOM_DEBUG_LOG` is set:
1. The word arguments it received (`os.Args[2:]`)
2. The number of command lines read from stdin
3. The result tokens it is about to print

#### Scenario: filter logs received words
- **WHEN** `mnmd filter dep` is called with `MONOM_DEBUG_LOG` set
- **THEN** the log contains a line showing `words=[dep]`

#### Scenario: filter logs stdin line count
- **WHEN** stdin contains 5 command paths and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line showing `commands=5`

#### Scenario: filter logs result tokens
- **WHEN** filter produces `deploy teardown` and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line showing those tokens
