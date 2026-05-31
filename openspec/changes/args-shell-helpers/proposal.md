## Why

CLI authors using `monomd args` currently call the binary once per flag. For scripts with multiple flags, this means repeating the pattern on many lines. Additionally, there's no standard way to bail out of a command script with an error message — every author writes their own `echo >&2; exit 1` boilerplate. Two small shell helper functions solve both problems and give CLI authors a cohesive scripting experience alongside `monomd args`.

## What Changes

- Add a `monom_args` shell function that parses multiple flags in a single declaration, delegating to `monomd args` under the hood
- Add a `monom_error` shell function that prints a message to stderr and exits with a given code
- Both functions are shipped as part of the monom shell runtime (sourced alongside `monom()`)

## Capabilities

### New Capabilities

- `multi-arg-helper`: The `monom_args` shell function — declares and parses multiple flags in one call, setting variables in the caller's scope
- `error-helper`: The `monom_error` shell function — prints to stderr and exits with a code

### Modified Capabilities

## Impact

- `src/monom` — sources the helper functions so they're available in CLI author scripts
- Shell tests — new shUnit2 tests for both helpers
