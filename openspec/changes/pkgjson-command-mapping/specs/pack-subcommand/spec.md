## MODIFIED Requirements

### Requirement: Pack takes space-separated args, joins with slashes, resolves to an absolute executable path
`mnmd pack <word...>` SHALL take one or more space-separated command tokens as CLI args. When the first token is not an absolute path, it SHALL join all tokens with `/` internally to produce a relative file path, resolve it against the project root (discovered via the same algorithm as `mnmd root`), and print the resulting absolute path to stdout. When the first token is an absolute path, no root discovery happens: the first token is the path to resolve, and any remaining tokens are the command's arguments, passed through unvalidated.

The resolved path (the joined tree path, or the absolute first token) MUST exist and MUST be executable; otherwise `mnmd pack` SHALL exit non-zero with an error message on stderr.

Output SHALL be newline-delimited tokens: the resolved absolute path on the first line, followed by one argument per line. Resolution without arguments therefore prints exactly one line — unchanged from the previous single-path contract. Argument lines are possible only in the absolute-first-token case; tree tokens are always path segments.

Pack is the symmetric counterpart to `filter`: both receive what the user typed (space-separated CLI args) and bridge to the slash-delimited file tree. Pack's specific job is to translate the spaces back to slashes and resolve the path — or, for an absolute first token (e.g. a `run`-hook mapping value), to validate it and pass the invocation through.

#### Scenario: Two-token input is joined with slashes and resolved
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack category1 sub_command1`
- **THEN** stdout is `"/home/user/myproject/category1/sub_command1"` and exit code is 0

#### Scenario: Single-token input resolves to a top-level command
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack command1`
- **THEN** stdout is `"/home/user/myproject/command1"` and exit code is 0

#### Scenario: Nested input is joined with slashes
- **WHEN** the project root is `/home/user/myproject` and pack is invoked as `mnmd pack infra cloud deploy`
- **THEN** stdout is `"/home/user/myproject/infra/cloud/deploy"` and exit code is 0

#### Scenario: Absolute single token is used as-is
- **WHEN** pack is invoked as `mnmd pack /opt/tools/deploy` and that file exists and is executable
- **THEN** stdout is exactly `"/opt/tools/deploy"` on one line, exit code is 0, and no project root discovery happens

#### Scenario: Absolute first token with arguments prints newline-delimited tokens
- **WHEN** pack is invoked as `mnmd pack /usr/local/bin/npm --prefix /repo/apps/web run build` and `/usr/local/bin/npm` exists and is executable
- **THEN** stdout is five lines — `/usr/local/bin/npm`, `--prefix`, `/repo/apps/web`, `run`, `build` — and exit code is 0

#### Scenario: Absolute first token that is not executable causes non-zero exit
- **WHEN** pack is invoked with an absolute first token plus arguments and that path exists but is not executable
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: No args causes non-zero exit
- **WHEN** pack is invoked with no args
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: File does not exist causes non-zero exit
- **WHEN** the resolved absolute path does not exist on the filesystem
- **THEN** an error message is printed to stderr and exit code is non-zero

#### Scenario: File exists but is not executable causes non-zero exit
- **WHEN** the resolved absolute path exists but does not have the executable bit set
- **THEN** an error message is printed to stderr and exit code is non-zero
