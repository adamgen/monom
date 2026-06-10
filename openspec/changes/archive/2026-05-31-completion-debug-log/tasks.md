## 1. Go: internal/debuglog package

- [x] 1.1 Create `internal/debuglog/debuglog.go` with `Log(format string, args ...any)` — no-op when `MONOM_DEBUG_LOG` is unset; appends a `[HH:MM:SS] message` line to the file otherwise
- [x] 1.2 Add `internal/debuglog/debuglog_test.go` — unit tests: no file created when env var unset; file created and line written when set; second call appends (does not overwrite)
- [x] 1.3 Run `go test ./...` — all tests pass

## 2. Shell: _monom_log helper in src/monom

- [x] 2.1 Add `_monom_log()` to `src/monom` — no-op when `MONOM_DEBUG_LOG` is unset; appends `[HH:MM:SS] $*` to `$MONOM_DEBUG_LOG` otherwise
- [x] 2.2 Run `shellcheck --shell=bash src/monom` — passes clean
- [x] 2.3 Run `bash tests/monom_bash_source_test` and `bash tests/monom_zsh_source_test` — still exits 0 and prints nothing
