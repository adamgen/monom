## ADDED Requirements

### Requirement: Print message to stderr and exit
`monom_error <message> [exit_code]` SHALL print `<message>` to stderr and exit the current script with `exit_code`.

#### Scenario: Default exit code
- **WHEN** `monom_error "something went wrong"` is called
- **THEN** stderr contains `something went wrong` and the script exits with code 1

#### Scenario: Custom exit code
- **WHEN** `monom_error "invalid input" 2` is called
- **THEN** stderr contains `invalid input` and the script exits with code 2

#### Scenario: Exit code zero is allowed
- **WHEN** `monom_error "done, but with warnings" 0` is called
- **THEN** stderr contains `done, but with warnings` and the script exits with code 0

### Requirement: Default exit code is 1
When no exit code argument is provided, `monom_error` SHALL use exit code 1.

#### Scenario: No exit code argument
- **WHEN** `monom_error "fail"` is called
- **THEN** exit code is 1

### Requirement: Message is printed verbatim
`monom_error` SHALL NOT add prefixes, formatting, or color to the message. The message is printed exactly as provided.

#### Scenario: No prefix added
- **WHEN** `monom_error "missing --name"` is called
- **THEN** stderr is exactly `missing --name` followed by a newline

### Requirement: Works in both bash and zsh
`monom_error` SHALL function identically in bash and zsh environments.

#### Scenario: Bash execution
- **WHEN** `monom_error "fail"` is called in a bash script
- **THEN** stderr contains `fail` and exit code is 1

#### Scenario: Zsh execution
- **WHEN** `monom_error "fail"` is called in a zsh script
- **THEN** stderr contains `fail` and exit code is 1
