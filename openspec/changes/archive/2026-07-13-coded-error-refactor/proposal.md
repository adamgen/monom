## Why

Each `runX` function in `cmd/mnmd/main.go` hardcodes its exit code and stderr formatting. The relationship between error kind and exit code is implicit, duplicated, and explained only in comments. `pack.GroupError` → exit 3 is the clearest symptom: the meaning lives in `internal/pack`, but its exit code lives in `main.go`. This makes the dispatcher brittle and hard to extend.

## What Changes

- Add a new constitution principle: *An error carries its own exit code.* A subcommand's outcome is determined by the typed error it returns, not by its call site in `main.go`.
- Introduce a `CodedError` interface and embeddable base in `internal/cli`, plus a single central exit-code registry (code ↔ message ↔ usage) as the sole source of truth.
- Convert `pack.GroupError` to embed the base, deriving its code from the registry.
- Wrap all generic error paths (root, check, install, pack real-errors) through the registry's code-1 entry.
- Replace per-`runX` exit-code logic in `main.go` with one uniform `errors.As` dispatch tail.
- Delete the now-redundant exit-code comment blocks in `main.go`.
- Replace `architecture.md`'s inline exit-code table with a reference to `internal/cli/cli.go` as the source of truth.

## Capabilities

### New Capabilities

- `coded-error`: CodedError interface, embeddable base, central exit-code registry, and uniform dispatch in `main.go`.

### Modified Capabilities

_(none — external CLI behavior and exit codes are unchanged; this is an internal structural refactor)_

## Impact

- `cmd/mnmd/main.go` — dispatcher rewritten to uniform error→exit-code tail.
- `internal/cli/` — new package containing `CodedError`, base, registry.
- `internal/pack/pack.go` — `GroupError` embeds coded base.
- `constitution.md` — new principle added.
- `architecture.md` — inline exit-code table replaced with a reference to `internal/cli/cli.go`.
- `internal/filter` — **unchanged**; filter must always exit 0.
