## 1. src/monom (core shell file)

- [ ] 1.1 Create `src/monom` with shebang `#!/usr/bin/env sh`
- [ ] 1.2 Set and export `MONOM_LIB_ROOT` using `$0`-based resolution at source time
- [ ] 1.3 Define `monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }`
- [ ] 1.4 Define `setup_monom()`: short-circuit if `$MONOM_PROJECT_ROOT` is set; otherwise call `monomd root`; export `MONOM_PROJECT_ROOT` and `MONOM_USER_CONFIG` on success; return non-zero on failure
- [ ] 1.5 Define `monom()`: call `setup_monom`; attempt `run` hook; call `monomd pack` with (hooked or original) args; exec the resolved path; exit non-zero if `monomd pack` fails
- [ ] 1.6 Add shell detection block: source `$MONOM_LIB_ROOT/monom.zsh` if `$ZSH_VERSION` is set, else source `$MONOM_LIB_ROOT/monom.bash` if `$BASH_VERSION` is set, else no-op

## 2. src/monom.bash (bash completion binding)

- [ ] 2.1 Create `src/monom.bash` with shebang `#!/usr/bin/env bash`
- [ ] 2.2 Define `monom_completion()`: call `setup_monom`; set `COMPREPLY` from `monom_cfg complete | monomd filter "${COMP_WORDS[@]:1}"` (always exit 0)
- [ ] 2.3 Register the hook: `complete -F monom_completion monom`

## 3. src/monom.zsh (zsh completion binding)

- [ ] 3.1 Create `src/monom.zsh`
- [ ] 3.2 Define `_monom()`: call `setup_monom`; call `compadd` with output of `monom_cfg complete | monomd filter "${words[@]:1}"` (always exit 0)
- [ ] 3.3 Guard `compdef` call: register `compdef _monom monom` only when `compdef` function is available

## 4. shellcheck

- [ ] 4.1 Run `shellcheck --shell=sh src/monom` — fix all warnings; document any necessary suppressions inline
- [ ] 4.2 Run `shellcheck --shell=bash src/monom.bash` — fix all warnings; document any necessary suppressions inline
- [ ] 4.3 Run `shellcheck --shell=bash src/monom.zsh` — fix all warnings; document any necessary suppressions inline

## 5. shUnit2 e2e tests

- [ ] 5.1 Create `tests/monom_shell_test` following the pattern in `tests/helpers`
- [ ] 5.2 Add `test_monom_lib_root_is_set`: source `src/monom`; assert `$MONOM_LIB_ROOT` equals the absolute path of `src/`
- [ ] 5.3 Add `test_setup_monom_uses_existing_project_root`: set `MONOM_PROJECT_ROOT` to a fixture project; call `setup_monom`; assert `$MONOM_USER_CONFIG` is set correctly and `monomd root` was not called
- [ ] 5.4 Add `test_setup_monom_discovers_root`: unset `MONOM_PROJECT_ROOT`; run from inside a fixture project directory; call `setup_monom`; assert `$MONOM_PROJECT_ROOT` and `$MONOM_USER_CONFIG` are set
- [ ] 5.5 Add `test_setup_monom_fails_outside_project`: unset `MONOM_PROJECT_ROOT`; run from `/tmp`; assert `setup_monom` returns non-zero
- [ ] 5.6 Add `test_monom_cfg_forwards_args`: after `setup_monom`, call `monom_cfg complete`; assert output matches fixture project's `complete` output
- [ ] 5.7 Add `test_monom_completion_defined_in_bash`: source `src/monom.bash` in bash; assert `monom_completion` function exists and `complete -p monom` shows `-F monom_completion`
- [ ] 5.8 Add `test_monom_zsh_function_defined`: source `src/monom.zsh` in zsh (via `zsh -c`); assert `_monom` function exists
- [ ] 5.9 Run `bash tests/monom_shell_test` — all tests pass

## 6. Validation

- [ ] 6.1 Run `make build` — confirm `bin/monomd` exists
- [ ] 6.2 Manually source `src/monom` in bash; run `monom <Tab>` against a fixture project; confirm completions appear
- [ ] 6.3 Manually source `src/monom` in zsh; run `monom <Tab>` against a fixture project; confirm completions appear
- [ ] 6.4 Confirm `go vet ./...` passes (no changes to Go code expected, but verify)
