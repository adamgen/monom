## ADDED Requirements

### Requirement: PTY session harness
The harness SHALL spawn `bash --norc --noprofile -i` attached to a PTY, source the script under test, and expose `send` and `waitFor` methods for driving and asserting on terminal output.

#### Scenario: Session starts and reaches prompt
- **WHEN** `newSession` is called
- **THEN** the harness waits for a shell prompt (`$ `) before returning

#### Scenario: Session sources the target script
- **WHEN** the session is created with a script path
- **THEN** the script is sourced and the shell is ready for commands

### Requirement: Single-Tab unambiguous completion
The harness SHALL assert that typing a unique prefix followed by Tab causes the completed name to appear in the terminal output.

#### Scenario: Unique prefix completes to one candidate
- **WHEN** `send("greet al\t")` is issued
- **THEN** `waitFor(regexp("alice"))` succeeds within the timeout

### Requirement: Double-Tab ambiguous listing
The harness SHALL assert that typing an ambiguous prefix followed by two Tab presses causes all matching candidates to appear in the terminal output.

#### Scenario: Ambiguous prefix lists multiple candidates
- **WHEN** `send("greet a\t\t")` is issued
- **THEN** `waitFor(regexp("alice"))` and `waitFor(regexp("arthur"))` both succeed

### Requirement: Double-Tab with no prefix lists all candidates
The harness SHALL assert that double-Tab with an empty prefix lists every registered completion candidate.

#### Scenario: Empty prefix after space lists all names
- **WHEN** `send("greet \t\t")` is issued
- **THEN** `waitFor` succeeds for each of alice, arthur, bob, carol, dave

### Requirement: Test isolation
Each test SHALL use its own session instance and close it on completion, so failures in one test do not affect others.

#### Scenario: Each test creates and closes its own session
- **WHEN** a test function calls `newSession` and defers `session.close()`
- **THEN** the bash process and PTY are fully torn down before the next test runs
