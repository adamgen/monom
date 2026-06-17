## MODIFIED Requirements

### Requirement: Filter bridges slash-delimited stdin and space-separated word arguments
`mnmd filter [word...]` SHALL read newline-delimited command paths from stdin (slash-delimited, e.g. `category1/sub_command1`) and accept zero or more word arguments (space-separated tokens the user has typed, e.g. `category1` `sub`). It SHALL join the word arguments with `/` internally to produce a prefix, then match that prefix against the stdin paths to determine next-level completions. No word arguments SHALL return all top-level tokens.

A trailing empty string argument signals that the user has completed the preceding token and pressed Tab — filter SHALL treat this as "drill into that level" (equivalent to appending `/` to the prefix).

#### Scenario: No arguments returns all top-level tokens
- **WHEN** stdin contains `"category1/sub1\ncategory1/sub2\ncommand1\ncommand2"` and no word arguments are given
- **THEN** stdout is `"category1\ncommand1\ncommand2"` (deduplicated top-level tokens)

#### Scenario: Partial word matches at top level
- **WHEN** stdin contains `"command1\ncommand2\ncategory1/sub1"` and the word argument is `"com"`
- **THEN** stdout is `"command1\ncommand2"`

#### Scenario: Partial category word returns the category token
- **WHEN** stdin contains `"category1/sub1\ncategory1/sub2\ncommand1"` and the word argument is `"categ"`
- **THEN** stdout is `"category1"`

#### Scenario: Complete category word followed by empty word drills into children
- **WHEN** stdin contains `"category1/sub_command1\ncategory1/sub_command2\ncommand1"` and word arguments are `"category1"` and `""`
- **THEN** stdout is `"sub_command1\nsub_command2"`

#### Scenario: Partial word within a category matches children
- **WHEN** stdin contains `"category1/sub_command1\ncategory1/sub_command2"` and word arguments are `"category1"` and `"sub_c"`
- **THEN** stdout is `"sub_command1\nsub_command2"`

#### Scenario: Nested drill-down — two complete words plus empty
- **WHEN** stdin contains `"infra/cloud/deploy\ninfra/cloud/teardown\ninfra/local/start"` and word arguments are `"infra"`, `"cloud"`, `""`
- **THEN** stdout is `"deploy\nteardown"`

#### Scenario: No matches at top level returns empty output
- **WHEN** stdin contains `"command1\ncommand2"` and the word argument is `"xyz"`
- **THEN** stdout is empty and exit code is 0

#### Scenario: Non-existent child of an existing category returns empty output
- **WHEN** stdin contains `"category1/sub1\ncategory1/sub2"` and word arguments are `"category1"` and `"xyz"`
- **THEN** stdout is empty and exit code is 0

#### Scenario: Drilling into a non-existent category returns empty output
- **WHEN** stdin contains `"category1/sub1\ncommand1"` and word arguments are `"nonexistent"` and `""`
- **THEN** stdout is empty and exit code is 0

#### Scenario: Duplicate top-level tokens are deduplicated
- **WHEN** stdin contains `"category1/sub1\ncategory1/sub2"` and no word arguments are given
- **THEN** stdout is `"category1"` (only once)

### Requirement: Filter silently ignores stdin paths that contain spaces
`mnmd filter` SHALL silently skip any stdin line that contains a space character in any path segment. Those paths SHALL be excluded from the output as if they were not present. No error SHALL be printed and exit code SHALL remain 0. Use `mnmd check` to surface these problems explicitly.

#### Scenario: Path with space in segment is excluded from output
- **WHEN** stdin contains `"my command/sub\ncommand1\ncommand2"` and no word arguments are given
- **THEN** stdout is `"command1\ncommand2"` (the invalid path is silently excluded)

#### Scenario: All paths invalid results in empty output
- **WHEN** stdin contains only paths with spaces in segments
- **THEN** stdout is empty and exit code is 0

### Requirement: Filter SHALL NEVER exit with a non-zero exit code
`mnmd filter` is invoked during interactive Tab completion. A non-zero exit code or noisy stderr output would degrade the user experience mid-typing. `mnmd filter` SHALL always exit with code 0, regardless of input validity, stdin read errors, or any other failure. On any internal error, it SHALL produce empty stdout and exit 0. Diagnostics belong in `mnmd check`, not here.

#### Scenario: Stdin read failure still exits 0
- **WHEN** stdin produces an I/O error during reading
- **THEN** stdout is empty (or partial) and exit code is 0

#### Scenario: Malformed stdin still exits 0
- **WHEN** stdin contains any malformed or unexpected content
- **THEN** exit code is 0

#### Scenario: Any unexpected error still exits 0
- **WHEN** any internal error occurs during filtering
- **THEN** exit code is 0 and stderr produces no output
