## Prior Art

The problem of parsing `--long=value` / `--long value` flags in shell scripts is well-known and historically under-served:

- **`getopts`** (POSIX built-in) — only handles single-character short flags (`-f value`). No `--long` support.
- **GNU `getopt`** (external) — supports long flags but is absent on macOS by default, behaves differently on BSD, and requires eval-based quoting workarounds. Not portable.
- **[Argbash](https://argbash.dev)** — a code generator: declare flags in a template, run `argbash`, get inlined parsing boilerplate. Adds a build step; the generated code lives in each script.
- **[getoptlong.sh](https://github.com/tecolicom/getoptlong)** — a pure-bash library (source into script). Associative-array option definitions, supports `--long`, arrays, hashes, callbacks. Requires Bash 4.2+.
- **Hand-rolled `case` loop** — the most common approach in practice; every team writes the same pattern repeatedly across scripts.

`mnmd args` inverts this: the script delegates to the binary (`PROP=$(mnmd args prop -- "$@")`), so parsing logic lives once in Go, is consistent across every command in a monom project, and works identically in bash and zsh command scripts.

---

## ADDED Requirements

### Requirement: Double-dash separates mnmd arguments from raw args
`mnmd args [modifiers...] <flag> -- <raw args...>` SHALL treat everything before `--` as its own arguments (modifiers and flag name) and everything after `--` as the raw argument list to search.

#### Scenario: Standard invocation with separator
- **WHEN** `mnmd args prop -- --prop=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Missing separator produces an error
- **WHEN** `mnmd args prop --prop=hello` is invoked (no `--` separator)
- **THEN** stderr contains an error message and exit code is 1

### Requirement: Resolve flag value from equals form
`mnmd args <flag> -- <raw args...>` SHALL parse `--<flag>=<value>` in the raw args and print `<value>` to stdout, exiting 0.

#### Scenario: Flag present in equals form
- **WHEN** `mnmd args prop -- --prop=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Flag value is empty string in equals form
- **WHEN** `mnmd args prop -- --prop=` is invoked
- **THEN** stdout is empty and exit code is 0

### Requirement: Resolve flag value from space form
`mnmd args <flag> -- <raw args...>` SHALL parse `--<flag> <value>` (two adjacent tokens) in the raw args and print `<value>` to stdout, exiting 0.

#### Scenario: Flag present in space form
- **WHEN** `mnmd args prop -- --prop hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Space-form value starts with non-flag character
- **WHEN** `mnmd args prop -- --prop some-value` is invoked
- **THEN** stdout is `some-value` and exit code is 0

### Requirement: Short modifier enables single-character alias
`mnmd args --short <char> <flag> -- <raw args...>` SHALL search for both `--<flag>` and `-<char>` in the raw args. The `--short` value MUST be exactly one character; more than one character SHALL produce an error. Short-form matching supports `-<char>=<value>` (equals) and `-<char> <value>` (space). Last-wins applies across both long and short forms.

#### Scenario: Short form in equals style
- **WHEN** `mnmd args --short p prop -- -p=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Short form in space style
- **WHEN** `mnmd args --short p prop -- -p hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Long form still works with short modifier
- **WHEN** `mnmd args --short=p prop -- --prop=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Last-wins across long and short forms
- **WHEN** `mnmd args --short p prop -- --prop=first -p last` is invoked
- **THEN** stdout is `last` and exit code is 0

#### Scenario: Multi-char short value produces error
- **WHEN** `mnmd args --short pp prop -- -pp=hello` is invoked
- **THEN** stderr contains an error and exit code is 1

### Requirement: Bundled short flags are recognized
When multiple single-character flags are bundled in the raw args (e.g., `-abc`), `mnmd args` SHALL recognize the target short flag within the bundle. The last character in a bundle MAY take a value (equals or space form). Characters before the last are treated as boolean flags.

#### Scenario: Target flag bundled with other flags (boolean)
- **WHEN** `mnmd args --boolean --short v verbose -- -xv` is invoked
- **THEN** exit code is 0 (verbose is present in the bundle)

#### Scenario: Target flag is last in bundle with space-form value
- **WHEN** `mnmd args --short p prop -- -xp hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Target flag is last in bundle with equals-form value
- **WHEN** `mnmd args --short p prop -- -xp=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Target flag is not last in bundle (boolean presence only)
- **WHEN** `mnmd args --boolean --short v verbose -- -vx` is invoked
- **THEN** exit code is 0 (verbose is present in the bundle)

#### Scenario: Target value flag is not last in bundle (no value available)
- **WHEN** `mnmd args --short p prop -- -px` is invoked
- **THEN** stdout is empty and exit code is 1 (value flag not in last position cannot take a value)

### Requirement: Exit non-zero when flag is absent
`mnmd args <flag> -- <raw args...>` SHALL exit 1, produce no stdout, and produce no stderr when the requested flag is not present in the raw args.

#### Scenario: Flag not in raw args
- **WHEN** `mnmd args prop -- --other=foo` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 1

#### Scenario: No raw args after separator
- **WHEN** `mnmd args prop --` is invoked (nothing after `--`)
- **THEN** stdout is empty, stderr is empty, and exit code is 1

### Requirement: Boolean modifier checks presence with no-prefix negation
`mnmd args --boolean <flag> -- <raw args...>` SHALL exit 0 if `--<flag>` is present in the raw args, and exit 1 if absent. No stdout SHALL be produced in either case. The parser SHALL also recognize `--no-<flag>` as explicit negation (treated as absent). Last-wins applies between `--<flag>` and `--no-<flag>`.

#### Scenario: Boolean flag present
- **WHEN** `mnmd args --boolean verbose -- --verbose --other=x` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 0

#### Scenario: Boolean flag absent
- **WHEN** `mnmd args --boolean verbose -- --other=x` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 1

#### Scenario: Boolean flag with equals value is still considered present
- **WHEN** `mnmd args --boolean verbose -- --verbose=yes` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 0

#### Scenario: Boolean flag negated with no-prefix
- **WHEN** `mnmd args --boolean verbose -- --no-verbose` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 1

#### Scenario: Last-wins between flag and no-prefix negation
- **WHEN** `mnmd args --boolean verbose -- --no-verbose --verbose` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 0

#### Scenario: Negation wins when it comes last
- **WHEN** `mnmd args --boolean verbose -- --verbose --no-verbose` is invoked
- **THEN** stdout is empty, stderr is empty, and exit code is 1

### Requirement: Last occurrence wins for duplicate flags
When `--<flag>` appears multiple times in the raw args, `mnmd args` SHALL return the value from the last occurrence and exit 0.

#### Scenario: Duplicate flags in equals form
- **WHEN** `mnmd args prop -- --prop=first --prop=last` is invoked
- **THEN** stdout is `last` and exit code is 0

#### Scenario: Mix of equals and space forms, last wins
- **WHEN** `mnmd args prop -- --prop=first --prop last` is invoked
- **THEN** stdout is `last` and exit code is 0

### Requirement: Ignore other flags while parsing
`mnmd args <flag> -- <raw args...>` SHALL ignore all flags and tokens in the raw args that are not the requested flag. It SHALL NOT error on unknown flags.

#### Scenario: Other flags present alongside target flag
- **WHEN** `mnmd args prop -- --other=x --prop=wanted --extra y` is invoked
- **THEN** stdout is `wanted` and exit code is 0

### Requirement: Space-form does not consume next-token if it looks like a flag
When `--<flag>` is followed by a token starting with `--` or `-`, `mnmd args` SHALL NOT treat that next token as the value.

#### Scenario: Space-form followed by another long flag
- **WHEN** `mnmd args prop -- --prop --other=x` is invoked
- **THEN** stdout is empty and exit code is 1 (no value found for `--prop`)

#### Scenario: Space-form followed by a short flag
- **WHEN** `mnmd args prop -- --prop -o` is invoked
- **THEN** stdout is empty and exit code is 1 (no value found for `--prop`)

### Requirement: Modifiers support both equals and space forms
All modifiers that accept a value (`--short`) SHALL support both `--mod=value` and `--mod value` syntaxes.

#### Scenario: Short modifier with equals form
- **WHEN** `mnmd args --short=p prop -- -p=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

#### Scenario: Short modifier with space form
- **WHEN** `mnmd args --short p prop -- -p=hello` is invoked
- **THEN** stdout is `hello` and exit code is 0

### Requirement: Unknown modifier produces an error
If a `--`-prefixed token appears before `--` and is not a recognized modifier or the flag name, `mnmd args` SHALL exit 1 and print an error to stderr.

#### Scenario: Unrecognized modifier
- **WHEN** `mnmd args --unknown prop -- --prop=hello` is invoked
- **THEN** stderr contains `mnmd args: unknown modifier --unknown` and exit code is 1
