## ADDED Requirements

### Requirement: Parse multiple flags in a single declaration
`monom_args [modifiers...] <name> ... -- <raw args...>` SHALL parse the declaration, call `monomd args` for each declared flag, and set a variable with the flag's name in the caller's scope.

#### Scenario: Single value flag
- **WHEN** `monom_args prop -- --prop=hello` is called in a script
- **THEN** the variable `$prop` is set to `hello` in the caller's scope

#### Scenario: Multiple value flags
- **WHEN** `monom_args name env port -- --name=adam --env=prod --port=8080` is called
- **THEN** `$name` is `adam`, `$env` is `prod`, and `$port` is `8080`

#### Scenario: Absent optional flag sets empty variable
- **WHEN** `monom_args prop -- --other=foo` is called
- **THEN** the variable `$prop` is set to `""` (empty string)

### Requirement: Modifiers apply to the flag they precede
Modifiers (`--short`, `--boolean`) placed before a flag name SHALL apply to that specific flag, matching the `monomd args` modifier syntax.

#### Scenario: Short modifier on one flag
- **WHEN** `monom_args --short p prop -- -p=hello` is called
- **THEN** `$prop` is `hello`

#### Scenario: Boolean modifier on one flag
- **WHEN** `monom_args --boolean verbose -- --verbose` is called
- **THEN** `$verbose` is `true`

#### Scenario: Mixed modifiers across multiple flags
- **WHEN** `monom_args --short=e env --boolean verbose --short p port -- --env=prod --verbose -p 8080` is called
- **THEN** `$env` is `prod`, `$verbose` is `true`, and `$port` is `8080`

### Requirement: Boolean flags set "true" or empty string
For flags declared with `--boolean`, `monom_args` SHALL set the variable to `"true"` if the flag is present and `""` if absent.

#### Scenario: Boolean flag present
- **WHEN** `monom_args --boolean verbose -- --verbose` is called
- **THEN** `$verbose` is `true`

#### Scenario: Boolean flag absent
- **WHEN** `monom_args --boolean verbose -- --other=x` is called
- **THEN** `$verbose` is `""` (empty string)

#### Scenario: Boolean flag negated with no-prefix
- **WHEN** `monom_args --boolean verbose -- --no-verbose` is called
- **THEN** `$verbose` is `""` (empty string)

### Requirement: Double-dash separates declarations from raw args
The `--` token SHALL separate the flag declarations (left side) from the raw argument list (right side). The raw args are passed to every `monomd args` invocation.

#### Scenario: Missing double-dash separator
- **WHEN** `monom_args prop --prop=hello` is called (no `--`)
- **THEN** an error is produced (stderr) and non-zero exit

### Requirement: Works in both bash and zsh
`monom_args` SHALL function identically in bash and zsh environments.

#### Scenario: Variable assignment in bash
- **WHEN** `monom_args prop -- --prop=hello` is called in a bash script
- **THEN** `$prop` is `hello`

#### Scenario: Variable assignment in zsh
- **WHEN** `monom_args prop -- --prop=hello` is called in a zsh script
- **THEN** `$prop` is `hello`
