## Why

monom is a CLI builder, but has no testing infrastructure for verifying that shell tab completion actually works end-to-end. We explored two approaches — an `expect` script and a Go/pty harness — and need to codify the chosen pattern so CLIs built with monom can include completion tests.

## What Changes

- Add a testing pattern for bash tab completion driven by a real PTY session
- Provide a reusable Go test harness (`session` helper) that spawns `bash --norc -i`, sources a script, sends keystrokes including Tab, and asserts completions appear in the output
- Ship a reference `expect`-based script as an alternative/documentation artifact
- Extend the project Makefile conventions to support `test-expect` and `test-go` targets

## Capabilities

### New Capabilities

- `completion-testing`: A testing pattern and harness for asserting bash tab completion behaviour via PTY. Covers single-Tab unambiguous completion, double-Tab listing, and full listing of all candidates.

### Modified Capabilities

## Impact

- New Go module dependency: `github.com/creack/pty`
- Tests require `bash` (not `zsh`) to be available on the host
- `expect` tests additionally require the `expect` binary (macOS: `brew install expect`)
- No changes to existing monom source; this is purely additive test infrastructure
