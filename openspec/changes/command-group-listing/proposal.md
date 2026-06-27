## Why

When a CLI user runs a `monom` command that resolves to a folder rather than a leaf script (e.g. `monom infra`, where `infra/` is a command group containing `cloud/` and `local/`), monom currently aborts with an internal, developer-oriented error: `mnmd pack: pack: resolved path is a directory, not a command: ...`. This leaks `mnmd` internals to the CLI user and dead-ends them instead of guiding them to the runnable commands underneath. Invoking a command group is the natural way a user discovers what verbs are available under a noun, and clig.dev calls for exactly this: show concise help and "suggest commands the user should run" rather than emitting creator-only output.

## What Changes

- When `mnmd pack` resolves the user's tokens to a **directory** (a command group) instead of an executable file, it SHALL no longer treat this as a generic error. Instead it SHALL exit with a **distinct, reserved exit code (3)** as a pure signal — writing nothing to stdout or stderr — so the shell can distinguish "this is a group" from "this is an error". `mnmd pack` SHALL NOT enumerate the group's children.
- The `monom()` shell function SHALL detect that reserved exit code and print a concise, user-facing group listing to stderr (e.g. `monom: 'infra' is a command group` followed by `available: cloud, local`), exiting non-zero. The internal `mnmd pack:` prefix SHALL NOT appear in this case.
- The child listing SHALL be sourced from the canonical discovery pipeline — `monom_cfg complete | mnmd filter <tokens> ""`, the same path tab-completion uses — so `complete` stays the single source of truth. This makes the listing identical to `monom <group> <Tab>` and honors any `run`-hook surface tree, and avoids giving the command tree a second discoverer (pack reading the filesystem) that could disagree with `complete`.
- No new top-level monom flag or override is introduced to make a group "also runnable." Authors who want a group to do something keep using the existing `run` hook (documented escape hatch) — this avoids the clig.dev "catch-all subcommand" time bomb where adding a real leaf later would silently change the meaning of an existing group invocation.
- `architecture.md` SHALL document the exit-code-3 group signal of `pack`, the `complete | filter` listing path, and explicitly point authors at the `run` hook as the sanctioned way to make a namespace runnable.

## Capabilities

### New Capabilities
<!-- None. This refines existing behavior of pack and the shell dispatch. -->

### Modified Capabilities
- `pack-group-signal`: a resolved path that is a directory changes from a generic error to a distinct, payload-free group signal — reserved exit code 3, empty stdout/stderr, no executable produced, and (deliberately) no child enumeration by pack.
- `monom-group-dispatch`: the `monom()` dispatch gains a branch that recognizes the reserved group exit code from `mnmd pack` and renders a concise user-facing group listing, sourcing the children from `complete | mnmd filter` instead of forwarding pack's raw stderr.

## Impact

- Code: `internal/pack/pack.go` (directory branch returns a typed group signal, no enumeration), `cmd/mnmd/main.go` (map the group outcome to the reserved exit code), `src/monom` (`monom()` dispatch branch lists children via `complete | filter`).
- Tests: `internal/pack/pack_test.go` (Go: directory returns the group sentinel; nested and empty groups), `tests/mnmd_pack_test` (e2e: reserved exit code 3 with empty stdout/stderr), `tests/monom_run_test` (e2e: user-facing group message with children from `complete`).
- Docs: `architecture.md` (`mnmd pack` exit-code table + `run` hook note).
- No change to the constitution-protected required user config interface. The group listing reuses the existing `complete | filter` pipeline; it adds one `complete`+`filter` roundtrip only on the cold group path (never on leaf execution or completion).
