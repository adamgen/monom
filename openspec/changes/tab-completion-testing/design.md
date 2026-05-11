## Context

We prototyped two approaches for testing bash tab completion in `temp/`:

1. **expect script** (`test_greet.exp`) — Tcl/expect spawns bash, sources the script under test, sends Tab keystrokes, and asserts on output patterns.
2. **Go/pty harness** (`greet_test.go`) — Go test with `github.com/creack/pty` spawns bash, drives it over a PTY, and uses `regexp.MustCompile` + a polling `waitFor` helper to assert on accumulated output.

Both approaches passed. We need to pick one as the canonical pattern for monom.

## Goals / Non-Goals

**Goals:**
- Define the canonical tab-completion testing pattern for CLIs built with monom
- Move `temp/` reference files into the change for exploration during implementation
- Support single-Tab and double-Tab completion assertions
- Be runnable with `make test-go` (or `make test-expect`)

**Non-Goals:**
- Supporting zsh or fish completion testing (bash only, for now)
- Integrating completion testing into monom's own build system (this is a pattern for users of monom)
- Replacing the existing shunit2 unit testing approach

## Decisions

### Decision 1: Go/pty as the canonical approach

**Choice:** Go/pty (`greet_test.go` pattern)

**Rationale:** monom is a Go-adjacent CLI ecosystem; Go tests integrate naturally with `go test ./...`, produce structured output (pass/fail per test), and are more portable than the `expect` binary (which is not always installed). The PTY harness we built is self-contained and has no external tool dependency beyond `bash`.

**Alternative considered:** `expect` script — simpler to write, but requires `expect` to be installed (not default on Linux CI), produces only stdout-based pass/fail, and is harder to integrate with standard CI reporters.

### Decision 2: Keep the `expect` script as a reference artifact

The `test_greet.exp` file documents the approach clearly and serves as a standalone runnable proof-of-concept. It will live alongside the implementation as a reference.

### Decision 3: PTY session struct pattern

The `session` struct (ptm, cmd, buf, waitFor) is the reusable unit. Each test creates a fresh session, sources the script, runs assertions, and defers `close()`. This avoids shared state between tests and makes failures easy to isolate.

## Risks / Trade-offs

- [Risk] Tests require `bash` on PATH → Mitigation: document requirement; skip with `t.Skip` on non-bash hosts if needed
- [Risk] PTY timing sensitivity (`waitFor` polls with 5 s deadline) → Mitigation: 20 ms polling interval is fast enough in practice; deadline is generous for CI
- [Risk] `creack/pty` does not support Windows → Mitigation: completion testing is a bash/POSIX concern; Windows is out of scope

## Open Questions

- Should `make test-go` live in the top-level monom Makefile or only in per-CLI Makefiles? (decide during implementation)
