## ADDED Requirements

### Requirement: Root honors $MONOM_PROJECT_ROOT when set to a valid project
`monomd root` SHALL first check `$MONOM_PROJECT_ROOT`. If the env var is set AND points to a directory that contains an executable file named `monom`, `monomd root` SHALL print that absolute path to stdout and exit 0 without walking from `$PWD`.

#### Scenario: MONOM_PROJECT_ROOT is set to a valid project
- **WHEN** `$MONOM_PROJECT_ROOT` is set to a directory containing an executable `monom` file
- **THEN** stdout is that directory's absolute path and exit code is 0; no upward walk happens

#### Scenario: MONOM_PROJECT_ROOT is set but missing the monom file
- **WHEN** `$MONOM_PROJECT_ROOT` is set but the directory does not contain an executable `monom` file
- **THEN** the env var is ignored and the upward walk from `$PWD` proceeds

#### Scenario: MONOM_PROJECT_ROOT is set but does not exist
- **WHEN** `$MONOM_PROJECT_ROOT` is set but the directory does not exist
- **THEN** the env var is ignored and the upward walk from `$PWD` proceeds

### Requirement: Root walks up from PWD to find the nearest project root
When `$MONOM_PROJECT_ROOT` is unset or invalid, `monomd root` SHALL walk upward from `$PWD` directory by directory, looking for a directory that contains an executable file named `monom`. When found, it SHALL print the absolute path of that directory to stdout and exit 0.

#### Scenario: monom file found in current directory
- **WHEN** `$MONOM_PROJECT_ROOT` is unset and `$PWD` contains an executable file named `monom`
- **THEN** stdout is the absolute path of `$PWD` and exit code is 0

#### Scenario: monom file found in a parent directory
- **WHEN** `$MONOM_PROJECT_ROOT` is unset, `$PWD` does not contain `monom`, but a parent directory does
- **THEN** stdout is the absolute path of the nearest ancestor containing `monom` and exit code is 0

#### Scenario: monom file not found in any ancestor
- **WHEN** `$MONOM_PROJECT_ROOT` is unset and no ancestor directory up to the filesystem root contains a `monom` file
- **THEN** an error message is printed to stderr and exit code is non-zero

### Requirement: Root stops at the filesystem root
`monomd root` SHALL stop the upward walk at the filesystem root (`/`) and not loop infinitely.

#### Scenario: Walk reaches root without finding monom
- **WHEN** the upward walk reaches `/` without finding a `monom` file
- **THEN** exit code is non-zero and no path is printed to stdout

### Requirement: Root requires the monom file to be executable
`monomd root` SHALL only accept a directory as the project root if the `monom` file within it is executable (has execute permission). A non-executable `monom` file SHALL be ignored and the walk SHALL continue upward.

#### Scenario: Non-executable monom file is skipped
- **WHEN** a directory contains a file named `monom` that is not executable
- **THEN** that directory is NOT returned; the walk continues to the parent
