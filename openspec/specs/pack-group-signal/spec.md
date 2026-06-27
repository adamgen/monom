## MODIFIED Requirements

### Requirement: Pack takes space-separated args, joins with slashes, resolves to an absolute executable path
`mnmd pack <word...>` SHALL take one or more space-separated command tokens as CLI args, join them with `/` internally to produce a relative file path, resolve it against the project root (discovered via the same algorithm as `mnmd root`), and print the resulting absolute path to stdout. The file MUST exist and MUST be executable; otherwise `mnmd pack` SHALL exit non-zero with an error message on stderr.

When the resolved path exists but is a **directory** (a command group), `mnmd pack` SHALL NOT treat this as a generic error. Instead it SHALL exit with the **reserved exit code 3** as a pure signal, writing nothing to stdout and nothing to stderr. Exit code 3 is reserved exclusively for the "resolved path is a command group" outcome and SHALL NOT be used for any other condition. Exit code 0 means a leaf command was resolved; any other non-zero exit means a real error (no args, not found, not executable, no project root).

`mnmd pack` SHALL NOT enumerate the group's children. Discovery of the command tree is the responsibility of the `complete` hook (the single source of truth); listing a group's children is the caller's job, performed by piping `complete` output through `mnmd filter`. This keeps `pack` a pure resolver that returns an executable path xor signals a non-leaf outcome, and prevents the command tree from having two independent discoverers that could disagree (e.g. when a `run` hook maps a surface tree that differs from the file tree).

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

#### Scenario: Resolved path is a command group exits 3 with no output
- **WHEN** the project root contains a directory `infra/`, and pack is invoked as `mnmd pack infra`
- **THEN** stdout is empty, stderr is empty, and exit code is 3

#### Scenario: Nested command group exits 3 with no output
- **WHEN** `infra/cloud/` is a directory, and pack is invoked as `mnmd pack infra cloud`
- **THEN** stdout is empty, stderr is empty, and exit code is 3

#### Scenario: Pack does not enumerate the group's children
- **WHEN** the resolved path is a directory containing several entries
- **THEN** pack writes nothing to stdout — the child listing is sourced from `complete | mnmd filter` by the caller, not produced by pack

#### Scenario: Empty command group exits 3 with no output
- **WHEN** the resolved path is a directory with no entries
- **THEN** stdout is empty, stderr is empty, and exit code is 3
