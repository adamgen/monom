## 1. Governance and documentation

- [x] 1.1 Amend `constitution.md`: replace "User Config Interface" principle with "Pluggability via Hooks" + "Required User Config Interface Requires a Constitution Amendment to Change". Required interface reduces to `complete`.
- [x] 1.2 Update `architecture.md`: add "Hooks" section documenting `run` as the first optional hook; update `monomd pack` and `monomd root` sections; update data flow diagram.

## 2. Go package: `root`

- [ ] 2.1 Create `internal/root/root.go` with `FindProjectRoot() (string, error)` — first checks `$MONOM_PROJECT_ROOT` (returns it if set AND points to a directory containing an executable `monom` file); otherwise walks upward from `$PWD` looking for one; errors if none found
- [ ] 2.2 Create `internal/root/root_test.go` covering: env var honored when valid, env var ignored when missing-file/missing-dir, found in current `$PWD`, found in a parent, not found anywhere, non-executable `monom` file is skipped (during walk), walk stops at filesystem root

## 3. Go package: `filter`

- [ ] 3.1 Create `internal/filter/filter.go` with `Filter(commands []string, words []string) []string` — receives the raw typed words, joins them with `/` internally to build the prefix, silently skips any command path with spaces in any segment, returns next-level matching tokens. SHALL never return an error and SHALL never panic — on any unexpected condition, return an empty slice.
- [ ] 3.2 Create `internal/filter/filter_test.go` covering: no words returns all top-level tokens, partial word matches at top level, partial category word returns category token, complete word + empty word drills into children, partial word within a category, nested drill-down (two complete words + empty), no matches returns empty slice, non-existent child of existing category returns empty, drilling into non-existent category returns empty, duplicates are deduplicated, path with space in segment is silently excluded

## 4. Go package: `pack`

- [ ] 4.1 Create `internal/pack/pack.go` with `Pack(words []string) (string, error)` — calls `FindProjectRoot` internally, joins words with `/`, joins to root, validates the result exists and is executable, returns the absolute path
- [ ] 4.2 Create `internal/pack/pack_test.go` covering: single-token path, two-token path (space → slash), nested path (multiple tokens → slashes), no project root discoverable → error, file not found, file exists but not executable, empty words slice → error

## 5. Go package: `check`

- [ ] 5.1 Create `internal/check/check.go` with `Check(userConfig string) ([]string, error)` — runs `$MONOM_USER_CONFIG complete`, validates each output path for spaces in any segment, returns a list of problem descriptions (empty list means healthy)
- [ ] 5.2 Create `internal/check/check_test.go` covering: all valid paths returns empty problems, path with space in segment is reported, multiple invalid paths all reported, missing or non-executable userConfig returns error

## 6. Binary entry point

- [ ] 6.1 Create `cmd/monomd/main.go` as `package main` with a skeleton `main()` function (no dispatch yet — just the file, the package declaration, and any required imports)
- [ ] 6.2 Wire subcommand dispatch in `cmd/monomd/main.go` for `filter`, `root`, `pack`, and `check`
- [ ] 6.3 `filter`: passes `os.Args[2:]` as the words slice to `Filter`, reads stdin as the command list, prints results one per line. SHALL always exit 0 — any read or processing error results in empty output and exit 0 (wrap dispatch in a `recover()` to swallow panics).
- [ ] 6.4 `root`: calls `FindProjectRoot()`, prints result to stdout (exit 1 + stderr on error)
- [ ] 6.5 `pack`: takes `os.Args[2:]` as space-separated tokens, calls `Pack`, prints result to stdout (exit 1 + stderr on error)
- [ ] 6.6 `check`: calls `Check(os.Getenv("MONOM_USER_CONFIG"))`, prints problems to stdout, exits non-zero if any found
- [ ] 6.7 Unknown subcommand or no args: print usage to stderr and exit 1

## 7. shUnit2 e2e tests

- [ ] 7.1 Set up minimal `monomd` build invocation for tests (e.g. `go build -o bin/monomd ./cmd/monomd` invoked from the test setup)
- [ ] 7.2 Write shUnit2 e2e test for `monomd filter`: pipe slash-delimited paths via stdin, assert stdout for no args, partial word, complete word + empty word (drill-down)
- [ ] 7.3 Write shUnit2 e2e test for `monomd filter` invalid input: verify paths with spaces are silently excluded
- [ ] 7.4 Write shUnit2 e2e test for `monomd filter` always exits 0: test with broken stdin, malformed input, missing args — all SHALL exit 0
- [ ] 7.5 Write shUnit2 e2e test for `monomd root` env-var honored: set `MONOM_PROJECT_ROOT` to a valid temp project dir, assert that dir is returned
- [ ] 7.6 Write shUnit2 e2e test for `monomd root` walk: unset `MONOM_PROJECT_ROOT`, cd into a subdir of a temp project, assert the project root is returned
- [ ] 7.7 Write shUnit2 e2e test for `monomd root` not found: cd into a directory with no `monom` ancestor, assert non-zero exit
- [ ] 7.8 Write shUnit2 e2e test for `monomd pack`: create a temp project with an executable script, invoke `monomd pack category sub` from within it, assert the absolute path output uses slashes
- [ ] 7.9 Write shUnit2 e2e test for `monomd check`: set `MONOM_USER_CONFIG` to a clean script (expect exit 0), then to a script printing a path with a space (expect exit non-zero)

## 8. Verify

- [ ] 8.1 `go test ./...` passes
- [ ] 8.2 All shUnit2 e2e tests pass when run directly
