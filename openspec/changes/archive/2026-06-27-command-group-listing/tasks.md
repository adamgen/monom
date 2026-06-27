## 1. Pack: group outcome in Go

- [x] 1.1 In `internal/pack/pack.go`, add an exported `GroupError` type (`Path string`) implementing `error`.
- [x] 1.2 In `Pack`, replace the `fi.IsDir()` error branch: return a `*GroupError` carrying the resolved directory path. Pack does not read or enumerate the directory's contents — discovery is the `complete` hook's job. Keep not-found and not-executable branches unchanged.
- [x] 1.3 Add Go unit tests in `internal/pack/pack_test.go`: directory returns `*GroupError` with the correct path; nested directory returns `*GroupError`; empty directory returns `*GroupError`; leaf/not-found/not-executable cases still behave as before (detect group via `errors.As`).

## 2. Binary: reserved exit code 3

- [x] 2.1 In `cmd/mnmd/main.go` `runPack`, use `errors.As` to detect `*pack.GroupError`; on match, write nothing to stdout or stderr, and `os.Exit(3)`. Leave the existing error path (exit 1, stderr) for all other errors.
- [x] 2.2 Add a `debuglog.Log` line for the group outcome (mirrors the existing resolved/failed logging).

## 3. Shell: monom() group branch

- [x] 3.1 In `src/monom`, capture `mnmd pack`'s stdout and exit status separately. On exit 0, exec as today.
- [x] 3.2 On exit 3, print `monom: '<last-token>' is a command group` to stderr; source children from `_monom_cfg complete | mnmd filter <tokens> ""` and, when that pipeline yields children, also print `available: <children joined by ", ">`; return non-zero without exec'ing. Use the last element of the unpacked args as `<last-token>`.
- [x] 3.3 On any other non-zero exit, forward `mnmd pack`'s stderr and return non-zero (preserve current behavior).
- [x] 3.4 Confirm the branch works identically in bash and zsh (array capture, `$?` after command substitution).

## 4. Tests: e2e surface

- [x] 4.1 In `tests/mnmd_pack_test`, add cases: `mnmd pack <group>` exits 3 with empty stdout and empty stderr; nested group exits 3 with empty stdout; empty group exits 3 with empty stdout.
- [x] 4.2 In `tests/monom_run_test`, add cases: `monom <group>` prints the `is a command group` / `available:` message to stderr and exits non-zero without running anything; group message uses the last typed token; empty group reported without an `available:` list. Cover both bash and zsh per the file's existing pattern.

## 5. Docs

- [x] 5.1 Update `architecture.md` `mnmd pack` section to document the exit-code-3 group signal (empty stdout/stderr, reserved code 3).
- [x] 5.2 Add a note (in the `pack` section and the `run` hook section) that making a namespace runnable is an author concern handled via the `run` hook, not a monom-level flag — referencing the clig.dev "no catch-all" rationale.

## 6. Validation

- [x] 6.1 Run `go vet ./...` and `go test ./...` — all green.
- [x] 6.2 Run the shUnit2 suites (`bash tests/mnmd_pack_test`, `bash tests/monom_run_test`) — all green.
- [x] 6.3 Run `shellcheck` on `src/monom` (and all shell files) — no new suppressions.
- [x] 6.4 Manually verify in the `fixtures/demo-project`: `monom infra` lists `cloud, local`; `monom infra cloud deploy` still runs.
