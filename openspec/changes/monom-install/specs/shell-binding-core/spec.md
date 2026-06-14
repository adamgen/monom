## ADDED Requirements

### Requirement: src/monom exports MONOM_ACTIVE=1 when sourced
When `src/monom` is sourced, it SHALL export `MONOM_ACTIVE=1` so that subprocesses (including `monomd`) can detect that the shell integration is active.

#### Scenario: MONOM_ACTIVE is set after sourcing
- **WHEN** a user sources `src/monom`
- **THEN** `$MONOM_ACTIVE` is exported and equals `1`

#### Scenario: monomd sees MONOM_ACTIVE in subprocess environment
- **WHEN** `src/monom` has been sourced and the user invokes `monomd` via the `monom()` function
- **THEN** `$MONOM_ACTIVE` is present in `monomd`'s environment
