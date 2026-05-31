## Why

CLI authors writing command scripts in bash or any other shell language have no standard way to parse named flags passed to their commands. Without a helper, each script must hand-roll `getopts` or ad-hoc argument parsing, which is error-prone and inconsistent across a project. `monomd args` provides a single, reliable way to extract flag values from a command invocation.

## What Changes

- Implement the `monomd args [modifiers...] <flag> -- <raw args...>` subcommand in the Go binary
- The subcommand accepts optional modifiers, a flag name, a `--` separator, and the raw argument list
- Supported flag forms in raw args: `--prop=value` (equals form) and `--prop value` (space form)
- Modifiers control behavior (all support both `--mod=val` and `--mod val` forms):
  - `--boolean` — presence check only; also recognizes `--no-<flag>` as explicit negation
  - `--short <char>` — also search for `-<char>` short form in raw args
- CLI authors consume it as:
  - `PROP=$(monomd args prop -- "$@")` — optional value
  - `PROP=$(monomd args --short p prop -- "$@")` — long and short form
  - `if monomd args --boolean verbose -- "$@"; then ...` — boolean with `--no-verbose` support

## Capabilities

### New Capabilities

- `args-subcommand`: The `monomd args` subcommand — parses flags from a raw argument list with support for value flags, boolean presence checks with `--no-` negation, and short-form aliases

### Modified Capabilities

## Impact

- `cmd/monomd/main.go` — dispatch wired for `args` subcommand
- `internal/args/` — new logic package with flag parsing and `*_test.go`
- `tests/monomd_args_test` — new shUnit2 e2e test
