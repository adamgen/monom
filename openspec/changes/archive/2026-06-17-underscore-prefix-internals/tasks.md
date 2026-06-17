## 1. Shell files

- [x] 1.1 In `src/monom`, rename definitions and call sites: `setup_monom`→`_setup_monom`, `monom_cfg`→`_monom_cfg`, `MONOM_BIN`→`_MONOM_BIN`, `MONOM_LIB_ROOT`→`_MONOM_LIB_ROOT`, `MONOM_PROJECT_ROOT`→`_MONOM_PROJECT_ROOT`, `MONOM_USER_CONFIG`→`_MONOM_USER_CONFIG`. Update binary resolution to `mnmd` (`whence -p mnmd` / `type -P mnmd`, fallback `$_MONOM_LIB_ROOT/../bin/mnmd`). Leave `monom`, `_monom_log`, `_monom_self`, and `MONOM_DEBUG_LOG` unchanged.
- [x] 1.2 In `src/monom.zsh`, update all references to the renamed names (`_setup_monom`, `_monom_cfg`, `_MONOM_BIN`) and any `monomd` mentions in comments. Leave the `_monom` function name and `compdef _monom monom` registration unchanged.
- [x] 1.3 In `src/monom.bash`, rename `monom_completion`→`_monom_completion` (definition and `complete -F _monom_completion monom`) and update references to `_setup_monom`, `_monom_cfg`, `_MONOM_BIN`, and any `monomd` mentions.

## 2. Binary rename (monomd → mnmd)

- [x] 2.1 `git mv cmd/monomd cmd/mnmd` (package stays `main`). Update any `monomd`-name strings inside `cmd/mnmd/main.go` (usage text) and `cmd/mnmd/nudge_test.go`.
- [x] 2.2 Update `Makefile`: `build` (`go build -o bin/mnmd ./cmd/mnmd`), `clean` (`rm -f bin/mnmd`), `test-e2e`/`check`/`lint` globs (`tests/mnmd_*_test`), and any comments mentioning `monomd`.
- [x] 2.3 Update `.gitignore` (`bin/mnmd`).
- [x] 2.4 Update `internal/install/install.go` (and `install_test.go`) for any embedded `monomd` binary-name string.

## 3. Go env-var read sites

- [x] 3.1 In `internal/root/root.go`, change `os.Getenv("MONOM_PROJECT_ROOT")` to `os.Getenv("_MONOM_PROJECT_ROOT")` and update the doc comment.
- [x] 3.2 In `internal/check/check.go`, change `os.Getenv("MONOM_USER_CONFIG")` to `os.Getenv("_MONOM_USER_CONFIG")` and update the four error message strings to read `_MONOM_USER_CONFIG`.
- [x] 3.3 In `cmd/mnmd/main.go`, change `os.Getenv("MONOM_USER_CONFIG")` to `os.Getenv("_MONOM_USER_CONFIG")` and update the related comment.
- [x] 3.4 Confirm `internal/debuglog` is left untouched — `MONOM_DEBUG_LOG` keeps its public name.

## 4. Tests

- [x] 4.1 Update Go tests that set the renamed env vars: `internal/root/root_test.go` and `internal/pack/pack_test.go` (`MONOM_PROJECT_ROOT`→`_MONOM_PROJECT_ROOT`). Leave `internal/debuglog/debuglog_test.go` (`MONOM_DEBUG_LOG`) untouched.
- [x] 4.2 `git mv` the e2e test files: `tests/monomd_test`→`tests/mnmd_test`, `tests/monomd_root_test`→`tests/mnmd_root_test`, `tests/monomd_pack_test`→`tests/mnmd_pack_test`, `tests/monomd_filter_test`→`tests/mnmd_filter_test`, `tests/monomd_install_test`→`tests/mnmd_install_test`. Update their `MONOMD=...bin/mnmd` paths and any `monomd` invocations.
- [x] 4.3 Update shUnit2 tests referencing renamed identifiers/binary: `tests/monom_shell_test`, `tests/monom_run_test`, `tests/monom_source_bash_test`, `tests/monom_source_zsh_test`, and `tests/helpers`.
- [x] 4.4 Add or adjust an assertion confirming `monom` completes but no `monom`/`MONOM`-prefixed internal identifier and no `monomd` binary appears under `monom<Tab>` (guards the core goal — one command in the namespace).

## 5. Documentation

- [x] 5.1 Reword `constitution.md` (Pluggability and Required-Interface principles) so the seam and stability contract are described in terms of the executable `monom` config file and its required `complete` subcommand — naming no env var. Rename any `monomd` references to `mnmd`. No amendment is triggered (the required interface is unchanged).
- [x] 5.2 Update `architecture.md`: rename `monom_cfg`/`setup_monom`/`monom_completion`, the `_MONOM_*` references, and `monomd`→`mnmd` throughout; reframe the Environment Variables table as internal shell↔Go plumbing; surface "user MAY pre-set the project root to skip discovery" as the one public affordance; note `MONOM_DEBUG_LOG` is intentionally unprefixed. (Env var/seam parts largely done already — verify and finish; the `mnmd` rename and Entry Points fix are new.)
- [x] 5.3 Fix the Entry Points contradiction in `architecture.md`: it already calls the binary "machinery," so update the note that says `monom<Tab>` offers "the two real commands, `monom` and `monomd`" to state only `monom` is in the namespace; ensure Entry Points never describes the binary as a user-run command. NOTE: the live `architecture.md` (main checkout) has an **Entry Points section the worktree copy lacks**, written with old pre-rename names (`setup_monom()`, `monom_cfg()`, `monomd()`, `MONOM_LIB_ROOT`). Apply must reconcile against main's current content — rename those names AND the binary, and remove `monomd()` from the list of functions `source monom` defines (no `monomd()` wrapper is created; see shell-binding-core).
- [x] 5.4 Update `terminology.md`: define "monom config file" and "project root" by concept; demote the env var to an internal-mechanism footnote; rename the `monomd` definition to `mnmd`.
- [x] 5.5 Update `CLAUDE.md` and `README.md`: rename `monomd`→`mnmd` (build commands, subcommand examples, file paths).

## 6. Validation

- [x] 6.1 Run `go vet ./...` and `go test ./...` — both pass.
- [x] 6.2 `make build` produces `bin/mnmd`; run the shUnit2 suites (`tests/mnmd_*_test`, `tests/monom_*_test`) — all pass.
- [x] 6.3 Run `shellcheck` on `src/monom`, `src/monom.zsh`, `src/monom.bash`, and the renamed test files — clean, no new suppressions.
- [x] 6.4 Grep the tree (excluding `openspec/changes/archive/**` and other unapplied change dirs) for `\bmonomd\b`, `\bMONOM_(BIN|LIB_ROOT|PROJECT_ROOT|USER_CONFIG)\b`, and the bare `setup_monom`/`monom_cfg`/`monom_completion` names; confirm zero live references remain.
