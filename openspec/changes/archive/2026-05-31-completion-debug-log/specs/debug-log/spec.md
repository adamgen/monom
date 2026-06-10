## ADDED Requirements

### Requirement: MONOM_DEBUG_LOG activates debug logging
When the environment variable `MONOM_DEBUG_LOG` is set to a non-empty string, monom SHALL treat it as a file path and append debug lines to that file from both the shell layer and the Go layer. When `MONOM_DEBUG_LOG` is unset or empty, no file I/O SHALL occur and behaviour SHALL be identical to the current production path.

#### Scenario: No file I/O when MONOM_DEBUG_LOG is unset
- **WHEN** `MONOM_DEBUG_LOG` is unset and a Tab completion is triggered
- **THEN** no file is created or written, and completion behaviour is unchanged

#### Scenario: Log file is created and written when MONOM_DEBUG_LOG is set
- **WHEN** `MONOM_DEBUG_LOG=/tmp/monom-debug.log` and a Tab completion is triggered
- **THEN** `/tmp/monom-debug.log` exists and contains at least one line after the completion

### Requirement: Debug lines are timestamped and append-only
Each debug line SHALL be prefixed with a wall-clock timestamp (HH:MM:SS format) and a layer tag. The log file SHALL be opened in append mode so multiple completions accumulate without overwriting previous output.

#### Scenario: Multiple completions accumulate in the log
- **WHEN** `MONOM_DEBUG_LOG` is set and two Tab completions are triggered in sequence
- **THEN** the log file contains lines from both completions

#### Scenario: Each line has a timestamp prefix
- **WHEN** a debug line is written
- **THEN** it begins with a string matching `[HH:MM:SS]`

### Requirement: Shell completion handler logs its full pipeline
The shell completion handlers (`monom_completion` in bash, `_monom` in zsh) SHALL log the following when `MONOM_DEBUG_LOG` is set:
1. Entry — the words received from the shell (`COMP_WORDS` in bash, `words` in zsh)
2. `setup_monom` outcome — success with discovered root, or failure
3. `monom_cfg complete` output — number of lines returned
4. `monomd filter` output — the final completions (or empty)

#### Scenario: Entry line includes the typed words
- **WHEN** the user types `monom dep<Tab>` and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording the words passed to the completion handler

#### Scenario: setup_monom failure is logged
- **WHEN** `setup_monom` returns non-zero and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line indicating `setup_monom` failed

#### Scenario: monom_cfg complete output count is logged
- **WHEN** `monom_cfg complete` returns N paths and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording the count N

#### Scenario: Final completions are logged
- **WHEN** `monomd filter` produces completions and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line recording those completions

### Requirement: monomd filter logs its inputs and outputs
`monomd filter` SHALL log the following when `MONOM_DEBUG_LOG` is set:
1. The word arguments it received (`os.Args[2:]`)
2. The number of command lines read from stdin
3. The result tokens it is about to print

#### Scenario: filter logs received words
- **WHEN** `monomd filter dep` is called with `MONOM_DEBUG_LOG` set
- **THEN** the log contains a line showing `words=[dep]`

#### Scenario: filter logs stdin line count
- **WHEN** stdin contains 5 command paths and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line showing `commands=5`

#### Scenario: filter logs result tokens
- **WHEN** filter produces `deploy teardown` and `MONOM_DEBUG_LOG` is set
- **THEN** the log contains a line showing those tokens

### Requirement: _monom_log shell helper is a no-op when MONOM_DEBUG_LOG is unset
The `_monom_log` function defined in `src/monom` SHALL return immediately without spawning any subprocess or opening any file when `MONOM_DEBUG_LOG` is unset or empty.

#### Scenario: _monom_log is silent when MONOM_DEBUG_LOG is unset
- **WHEN** `_monom_log "some message"` is called and `MONOM_DEBUG_LOG` is unset
- **THEN** nothing is written anywhere and the function exits 0

### Requirement: Go debuglog.Log is a no-op when MONOM_DEBUG_LOG is unset
The `debuglog.Log` function in `internal/debuglog` SHALL return immediately without any file I/O when `os.Getenv("MONOM_DEBUG_LOG")` is empty.

#### Scenario: Log is a no-op with no env var
- **WHEN** `debuglog.Log("msg")` is called and `MONOM_DEBUG_LOG` is not set
- **THEN** no file is opened, no bytes are written, and the function returns without error
