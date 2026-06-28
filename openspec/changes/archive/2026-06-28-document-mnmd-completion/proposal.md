## Why

`mnmd` is an internal binary with a fixed set of subcommands, but pressing `mnmd <Tab>` in the terminal produces nothing — the shell has no completion registered for it. Adding `_mnmd_completion` (bash) and `_mnmd` (zsh) alongside the existing `monom` completion handlers gives developers the same Tab-complete experience for the internal tooling surface.

## What Changes

- `src/monom.bash` gains `_mnmd_completion()` — a bash completion handler registered via `complete -F _mnmd_completion mnmd` that completes the first positional argument (the subcommand slot) with the known mnmd subcommands.
- `src/monom.zsh` gains `_mnmd()` — a zsh completion function registered via `compdef _mnmd mnmd` (guarded by the same `compinit` availability check used for `_monom`) that populates the subcommand slot via `compadd`.
- Both handlers only complete at the subcommand position (the first argument); no completion is offered for positions beyond the subcommand.
- The subcommand list is static and hardcoded: `filter root pack check install completion`.

## Capabilities

### New Capabilities

*(none — this documents additions to existing shell binding capabilities)*

### Modified Capabilities

- `shell-binding-bash`: new requirement — `_mnmd_completion` is defined and registered for `mnmd`; filters by prefix at the subcommand slot; produces no completions beyond position 1.
- `shell-binding-zsh`: new requirement — `_mnmd` is defined and registered for `mnmd` via guarded `compdef`; populates the subcommand slot via `compadd`; produces no completions beyond position 1.

## Impact

- `src/monom.bash` — appended; no existing behavior changed.
- `src/monom.zsh` — appended; no existing behavior changed.
- `tests/mnmd_completion_test` — new shUnit2 e2e test file covering both shells (8 tests).
- No Go changes. No new env vars. No install changes.
