## Why

The monom binary (`monomd`) is complete, but nothing connects it to the user's shell session. The shell binding files (`src/monom`, `src/monom.bash`, `src/monom.zsh`) defined in `architecture.md` do not yet exist — without them, users cannot source monom, get tab completion, or run commands.

## What Changes

- Add `src/monom` — the shell-agnostic entry file sourced from the user's rc. Defines `monom()`, `setup_monom()`, and `monom_cfg()`. Detects the active shell and sources the correct completion file.
- Add `src/monom.bash` — registers the bash completion hook (`complete -F monom_completion monom`); defines `monom_completion()`.
- Add `src/monom.zsh` — registers the zsh completion hook (`compdef _monom monom`); defines `_monom()`.
This change is implemented in parallel with `tab-completion-testing`, which defines the canonical Go/pty harness for asserting tab-completion behavior interactively. The two changes are independent: this change delivers the shell files under test; `tab-completion-testing` delivers the testing harness. Integration (running the harness against the new bindings) follows after both land.

## Capabilities

### New Capabilities

- `shell-binding-core`: The `src/monom` file — sourcing, `setup_monom()`, `monom()` function, `monom_cfg` wrapper, and shell detection logic.
- `shell-binding-bash`: The `src/monom.bash` file — bash completion registration and `monom_completion()`.
- `shell-binding-zsh`: The `src/monom.zsh` file — zsh completion registration and `_monom()`.

### Modified Capabilities

## Impact

- New files: `src/monom`, `src/monom.bash`, `src/monom.zsh`
- No changes to `monomd` Go binary or existing subcommands
- No changes to existing specs (`filter`, `pack`, `root`, `check`)
- Depends on `monomd` being built and on PATH (tests require `bin/monomd`)
