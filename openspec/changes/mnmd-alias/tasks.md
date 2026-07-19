## 1. Extract shared rc-file package

- [x] 1.1 Create `internal/rc` and move `rcFileForShell`, `alreadyInstalled`/line-presence check, `appendSourceLine`/append, and `needsLeadingNewline` out of `internal/install` into it, as exported functions (e.g. `FileForShell`, `HasLine`, `AppendLine`)
- [x] 1.2 Update `internal/install/install.go` to call `internal/rc`; keep `install` behavior unchanged
- [x] 1.3 Add `internal/rc/rc_test.go` covering shell selection, line-presence idempotency, and trailing-newline handling
- [x] 1.4 Run `go test ./internal/rc/... ./internal/install/...` and confirm existing install behavior is preserved

## 2. Implement the alias-save binary subcommand

- [x] 2.1 Create `internal/alias/alias.go` with `Save(name, path string) error`: validate `name` is shell-identifier-safe, validate `path` is a directory containing an executable `monom` config (reuse `internal/root`'s validity definition), resolve `path` to absolute
- [x] 2.2 In `Save`, build the line `_monom_bind_alias <name> "<abs-path>"` and append it via `internal/rc` if absent; report "already saved" when present; print the modified rc path on first save
- [x] 2.3 Return typed/coded errors via `internal/cli` for invalid name, invalid root, and unsupported shell
- [x] 2.4 Add `internal/alias/alias_test.go` covering name validation, root validation, absolute-path resolution, idempotent re-save, and the generated line format
- [x] 2.5 Wire `case "alias-save"` into `cmd/mnmd/main.go` dispatch (reads `os.Args[2]`, `os.Args[3]`) with `handleError`; add `alias`/`alias-save` to `usage()`

## 3. Shell: shared execution and completion cores

- [x] 3.1 In `src/monom`, extract `monom()`'s body into `_monom_run <name> [args...]` that reads `_MONOM_PROJECT_ROOT` from scope and prefixes user-facing messages with `<name>`; redefine `monom()` as `_monom_run monom "$@"`
- [x] 3.2 In `src/monom.bash`, extract `_monom_complete()` core reading root from scope; redefine `_monom_completion` to delegate to it
- [x] 3.3 In `src/monom.zsh`, extract `_monom_complete_zsh()` core (preserving skip-`compadd`-when-empty) reading root from scope; redefine `_monom` to delegate to it
- [x] 3.4 Confirm bare `monom` execution, group listing, and completion are unchanged (existing `shell-binding-*` e2e tests still pass)

## 4. Shell: alias binding and interception

- [x] 4.1 In `src/monom`, define `_monom_bind_alias <name> <root>`: `eval`-define `<name>()` pinning `local _MONOM_PROJECT_ROOT` and delegating to `_monom_run`; `eval`-define a per-name completion function pinning the root and delegating to the completion core; register it via `complete -F` (bash) / `compdef` (zsh) with shell detection
- [x] 4.2 In `src/monom`, make `mnmd()` intercept `alias`: call `_monom_bind_alias "$2" "$3"` then `"$_bin" alias-save "$2" "$3"`; all other subcommands pass through to the binary unchanged
- [x] 4.3 Add `alias` to the `mnmd` subcommand completion lists in both `src/monom.bash` (`_mnmd_completion`) and `src/monom.zsh` (`_mnmd`)
- [x] 4.4 Verify an aliased command fails cleanly (clear error, no exec) when its pinned root is missing

## 5. Tests: alias surface

- [x] 5.1 Add `tests/mnmd_alias_test` (shUnit2 e2e): `mnmd alias` binds + persists, immediate availability in-shell, idempotent re-save, invalid name and invalid root rejected, alias executes a leaf, alias lists a command group with the `<name>:` prefix, alias no-args lists top level
- [x] 5.2 Add e2e coverage that a bare `monom` after an alias call still auto-discovers, and that two aliases to different roots coexist (root-pinning does not clobber globals)
- [x] 5.3 Add completion e2e for an aliased command (`<name> <Tab>` yields the pinned project's top-level completions)
- [x] 5.4 Add an e2e asserting the persisted rc line is a direct `_monom_bind_alias` call (no `mnmd`/subprocess invocation at bind time)

## 6. Docs and final validation

- [x] 6.1 Update `architecture.md`: replace the `make_monom_alias` placeholder with the alias design, document `mnmd alias <name> <path>`, the `_monom_bind_alias` rc line, and the `alias-save` internal subcommand
- [x] 6.2 Run `go vet ./...` and `go test ./...`; fix any failures
- [x] 6.3 Run `shellcheck` on `src/monom`, `src/monom.bash`, `src/monom.zsh` with no new suppressions
- [x] 6.4 Run the full shUnit2 e2e suites and confirm `make check` passes (once build infra exists) against the Validation Checklist in `CLAUDE.md`
