## 1. Shell Binding: Export MONOM_ACTIVE

- [x] 1.1 Add `export MONOM_ACTIVE=1` to `src/monom` (near the top, after `MONOM_LIB_ROOT` is set)
- [x] 1.2 Add shUnit2 test in `tests/monom_source_zsh_test` (or equivalent) verifying `$MONOM_ACTIVE` equals `1` after sourcing

## 2. Install Logic Package

- [x] 2.1 Create `internal/install/install.go` with a `Run()` function implementing the install logic
- [x] 2.2 Implement shell detection via `$SHELL` → zsh → `~/.zshrc`, bash → `~/.bash_profile`
- [x] 2.3 Implement binary path resolution via `os.Executable()` + `filepath.EvalSymlinks` → `../src/monom`
- [x] 2.4 Implement idempotency check: scan rc file for a line containing the resolved `src/monom` path
- [x] 2.5 Implement append with newline-prefix guard (check last byte of file)
- [x] 2.6 Print modified file path and restart hint on success; print "already installed" if already present
- [x] 2.7 Add `internal/install/install_test.go` covering: zsh detection, bash detection, idempotency, symlink resolution, unknown shell error, newline guard

## 3. Install Nudge

- [x] 3.1 Add a `checkNudge()` helper in `cmd/mnmd/main.go` (or a shared `internal/nudge/` package) that prints the hint to stderr when `$MONOM_ACTIVE` is unset
- [x] 3.2 Call `checkNudge()` at the top of `main()`, skipping it when the subcommand is `install`
- [x] 3.3 Add Go unit test verifying nudge fires when `MONOM_ACTIVE` is unset and is suppressed when set
- [x] 3.4 Add shUnit2 e2e scenario to `tests/mnmd_install_test` verifying nudge appears on stderr for non-install subcommands when `MONOM_ACTIVE` is unset

## 4. mnmd install Subcommand

- [x] 4.1 Wire `install` dispatch in `cmd/mnmd/main.go` calling `internal/install.Run()`
- [x] 4.2 Create `tests/mnmd_install_test` with shUnit2 e2e tests covering: zsh rc write, bash rc write, idempotency, unknown shell error, stdout/stderr separation

## 5. Validation

- [x] 5.1 Run `go vet ./...` — no errors
- [x] 5.2 Run `go test ./...` — all tests pass
- [x] 5.3 Run `shellcheck src/monom` — no new warnings
- [x] 5.4 Run `bash tests/monom_source_zsh_test` — passes
- [x] 5.5 Run `bash tests/mnmd_install_test` — passes
