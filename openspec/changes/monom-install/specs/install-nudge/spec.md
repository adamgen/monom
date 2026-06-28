## ADDED Requirements

### Requirement: mnmd prints a nudge when shell integration is not active
When `mnmd` is invoked directly (not via the `monom()` shell function) and `$MONOM_ACTIVE` is unset, `mnmd` SHALL print a one-line hint to stderr on every subcommand invocation except `install`. The hint SHALL suggest running `mnmd install`.

#### Scenario: Nudge fires when MONOM_ACTIVE is unset
- **WHEN** `mnmd filter foo` is called and `$MONOM_ACTIVE` is not set in the environment
- **THEN** `mnmd` prints `hint: run 'mnmd install' to activate shell integration` to stderr before its normal output, and exits with the subcommand's normal exit code

#### Scenario: Nudge suppressed when MONOM_ACTIVE is set
- **WHEN** `mnmd filter foo` is called and `$MONOM_ACTIVE=1` is set in the environment
- **THEN** no hint is printed to stderr

#### Scenario: Nudge suppressed for the install subcommand itself
- **WHEN** `mnmd install` is called and `$MONOM_ACTIVE` is unset
- **THEN** no hint is printed (the user is already running the install command)

### Requirement: Nudge goes to stderr only
The nudge hint SHALL be written to stderr and SHALL NOT appear on stdout, so it does not pollute piped or captured output from subcommands like `mnmd filter` or `mnmd pack`.

#### Scenario: Stdout is clean despite nudge
- **WHEN** `mnmd pack foo bar` is called with `$MONOM_ACTIVE` unset, capturing stdout
- **THEN** stdout contains only the resolved command path; the hint appears only on stderr
