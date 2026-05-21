## ADDED Requirements

### Requirement: Check runs the project's complete output and reports problems
`monomd check` SHALL invoke `$MONOM_USER_CONFIG complete`, read all output paths, validate each one, and report any problems to stdout. It SHALL exit 0 if no problems are found and exit non-zero if any problems are found.

#### Scenario: All paths valid — success output
- **WHEN** `$MONOM_USER_CONFIG complete` produces only valid slash-delimited paths with no spaces in any segment
- **THEN** `monomd check` prints a success summary (e.g. `✔ N commands OK`) and exits 0

#### Scenario: Path with space in segment — problem reported
- **WHEN** `$MONOM_USER_CONFIG complete` produces a path containing a space in any segment (e.g. `"my command/sub"`)
- **THEN** `monomd check` prints an error identifying the offending path and exits non-zero

#### Scenario: Multiple problems — all reported
- **WHEN** multiple invalid paths are present in the complete output
- **THEN** all problems are reported and the count is included in the summary

### Requirement: Check requires MONOM_USER_CONFIG to be set
`monomd check` SHALL exit non-zero with an error message if `$MONOM_USER_CONFIG` is not set or the file is not executable.

#### Scenario: MONOM_USER_CONFIG not set
- **WHEN** `$MONOM_USER_CONFIG` is not set in the environment
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: MONOM_USER_CONFIG not executable
- **WHEN** `$MONOM_USER_CONFIG` points to a file that is not executable
- **THEN** an error message is printed to stderr and exit code is non-zero
