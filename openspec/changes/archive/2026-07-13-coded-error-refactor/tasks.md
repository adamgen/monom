## 1. internal/cli package

- [x] 1.1 Create `internal/cli/cli.go` with `CodedError` interface, `Base` struct, exit-code registry (`ExitCodes`), and `WrapError` constructor
- [x] 1.2 Add `internal/cli/cli_test.go` covering: Base satisfies CodedError, WrapError wraps with code 1, ExitCode values match expected constants

## 2. pack.GroupError refactor

- [x] 2.1 Update `internal/pack/pack.go`: GroupError embeds `cli.Base`, constructor sets code from `cli.ExitCodes.GroupError`
- [x] 2.2 Update `internal/pack/pack_test.go`: verify GroupError satisfies `cli.CodedError` and returns exit code 3

## 3. main.go uniform dispatch

- [x] 3.1 Refactor `runRoot`, `runPack`, `runCheck`, `runInstall` to return `error` instead of calling `os.Exit` directly
- [x] 3.2 Add a single dispatch tail in `main()` that resolves `CodedError` via `errors.As`, suppresses stderr for `GroupError` code, defaults to `ExitCodes.Error`
- [x] 3.3 Remove now-redundant exit-code comment blocks from main.go

## 4. Documentation

- [x] 4.1 Add "Principle: Errors Carry Their Own Exit Code" to `constitution.md`
- [x] 4.2 Replace `architecture.md`'s inline exit-code table with a reference to `internal/cli/cli.go`

## 5. Validation

- [x] 5.1 Run `go vet ./...` and `go test ./...` — all pass
- [x] 5.2 Run `bash tests/mnmd_pack_test` and `bash tests/monom_run_test` — all pass
- [x] 5.3 Run `shellcheck` on shell files — all pass
