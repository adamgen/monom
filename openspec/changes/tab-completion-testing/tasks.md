## 1. Reference Files

- [ ] 1.1 Review `ref/greet` (the sample CLI with bash completion) and decide if it becomes the canonical example or a test fixture
- [ ] 1.2 Review `ref/greet_test.go` (Go/pty harness) and identify what needs to change to be reusable beyond `greet`
- [ ] 1.3 Review `ref/test_greet.exp` (expect script) — confirm it stays as reference documentation only

## 2. Go Test Harness

- [ ] 2.1 Create `src/completion_test_harness/session.go` (or similar) with the `session` struct, `newSession`, `send`, `waitFor`, and `close` extracted from `ref/greet_test.go`
- [ ] 2.2 Parameterise `newSession` to accept a script path rather than hard-coding `greet`
- [ ] 2.3 Add `go.mod` with `github.com/creack/pty` dependency at the appropriate location

## 3. Example Test

- [ ] 3.1 Add a `greet_test.go` (or equivalent) that exercises the five completion scenarios from `specs/completion-testing/spec.md`
- [ ] 3.2 Verify single-Tab unambiguous completion (alice, bob, carol)
- [ ] 3.3 Verify double-Tab ambiguous listing (alice + arthur for prefix `a`)
- [ ] 3.4 Verify double-Tab full listing (all five names with empty prefix)

## 4. Build Integration

- [ ] 4.1 Add `test-go` target to the relevant Makefile (`go test -v ./...`)
- [ ] 4.2 Confirm `make test-go` passes on macOS with `bash` on PATH
- [ ] 4.3 Document the `bash` and optional `expect` prerequisites in a comment or README section
