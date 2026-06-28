## Why

When users install monom and run `mnmd` directly, they may not have the `monom` shell function active — meaning tab completion and command dispatch won't work until they manually add a `source` line to their shell's rc file. This friction is unnecessary and undiscoverable; `mnmd install` closes the gap by detecting the user's shell and writing the correct source line automatically.

## What Changes

- Add a `mnmd install` subcommand that detects the active shell, selects the correct rc/profile file, and appends a `source "/path/to/src/monom"` line if one is not already present.
- Add a nudge: when `mnmd` is invoked but the `monom` shell function is not active (detectable by checking whether the caller is a raw binary invocation rather than the shell function), print a one-time hint suggesting `mnmd install`.

## Capabilities

### New Capabilities

- `monom-install`: The `mnmd install` subcommand — detects the shell, resolves the rc/profile file, writes the source line idempotently, and reports what it did.
- `install-nudge`: Nudge printed to stderr when `mnmd` is called directly (outside the `monom()` shell function) suggesting the user run `mnmd install`.
- `shell-binding-core`: Updated — the existing shell binding entrypoint may need a small marker or env var so the nudge can tell whether the function is active.

### Modified Capabilities

- `shell-binding-core`: No spec-level requirement changes; the requirement contract is unchanged. Only the `install` subcommand depends on its output path convention (`$MONOM_LIB_ROOT/monom`), which is already documented.

## Impact

- `cmd/mnmd/main.go` — new `install` subcommand dispatch.
- `internal/install/` — new logic package implementing rc-file detection and source-line insertion.
- `src/monom` — may need an exported marker variable (`MONOM_ACTIVE=1`) so the nudge can detect whether the shell function is loaded.
- `tests/mnmd_install_test` — new shUnit2 e2e test covering the install subcommand.
- No changes to the user config interface; `install` is a one-time setup command, not part of the completion or run path.
