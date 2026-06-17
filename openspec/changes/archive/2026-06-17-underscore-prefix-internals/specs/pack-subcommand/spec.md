## MODIFIED Requirements

### Requirement: Pack takes space-separated args, joins with slashes, resolves to an absolute executable path
`mnmd pack <word...>` SHALL take one or more space-separated command tokens as CLI args, join them with `/` internally to produce a relative file path, resolve it against the project root (discovered via the same algorithm as `mnmd root`), and print the resulting absolute path to stdout. The file MUST exist and MUST be executable; otherwise `mnmd pack` SHALL exit non-zero with an error message on stderr.

Pack is the symmetric counterpart to `filter`: both receive what the user typed (space-separated CLI args) and bridge to the slash-delimited file tree. Pack's specific job is to translate the spaces back to slashes and resolve the path.

#### Scenario: Two-token input is joined with slashes and resolved
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack category1 sub_command1`
- **THEN** stdout is `"/home/user/myproject/category1/sub_command1"` and exit code is 0

#### Scenario: Single-token input resolves to a top-level command
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack command1`
- **THEN** stdout is `"/home/user/myproject/command1"` and exit code is 0

#### Scenario: Nested input is joined with slashes
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack infra cloud deploy`
- **THEN** stdout is `"/home/user/myproject/infra/cloud/deploy"` and exit code is 0

#### Scenario: No args causes non-zero exit
- **WHEN** pack is invoked with no args
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: File does not exist causes non-zero exit
- **WHEN** the resolved absolute path does not exist on the filesystem
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: File exists but is not executable causes non-zero exit
- **WHEN** the resolved absolute path exists but does not have the executable bit set
- **THEN** an error message is printed to stderr and exit code is non-zero

### Requirement: Pack discovers the project root internally
`mnmd pack` SHALL discover the project root using the same algorithm as `mnmd root`: if `$_MONOM_PROJECT_ROOT` is set and points to a directory containing an executable `monom` file, use it; otherwise, walk up from `$PWD` looking for a directory containing an executable `monom` file. If no project root can be discovered, `mnmd pack` SHALL exit non-zero with an error message on stderr.

#### Scenario: _MONOM_PROJECT_ROOT is honored when valid
- **WHEN** `$_MONOM_PROJECT_ROOT` is set to a directory containing an executable `monom` file
- **THEN** pack uses that directory as the root without walking from `$PWD`

#### Scenario: Walk from PWD when _MONOM_PROJECT_ROOT is unset
- **WHEN** `$_MONOM_PROJECT_ROOT` is unset and a parent directory of `$PWD` contains an executable `monom` file
- **THEN** pack discovers that directory as the root

#### Scenario: No project root found causes non-zero exit
- **WHEN** `$_MONOM_PROJECT_ROOT` is unset and no ancestor of `$PWD` contains an executable `monom` file
- **THEN** an error message is printed to stderr and exit code is non-zero
