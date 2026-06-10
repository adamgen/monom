## Why

Tab completion produces no results in a real shell session but the pipeline has three separate layers (shell binding → `monom_cfg complete` → `monomd filter`) and every layer suppresses errors to protect the interactive experience. There is currently no way to see what is actually happening inside a live completion without removing those suppressions. A debug log file solves this: when `$MONOM_DEBUG_LOG` is set, every step writes a timestamped trace, making it trivial to identify where the pipeline breaks without modifying production behaviour.

## What Changes

- Introduce `MONOM_DEBUG_LOG` as a reserved environment variable: when set to a file path, monom layers MAY append debug lines to that file.
- Add `_monom_log` shell helper to `src/monom` — no-op when `MONOM_DEBUG_LOG` is unset; appends a timestamped line to the file when set. Used by any shell layer that wants to log.
- Add `internal/debuglog` Go package — `Log(format, args...)` no-op when `MONOM_DEBUG_LOG` is unset; appends a timestamped line otherwise. Available to any `monomd` subcommand that wants to log.

What each layer actually logs is wiring detail, done incrementally outside this spec as debugging needs arise.

## Capabilities

### New Capabilities

- `debug-log`: The `MONOM_DEBUG_LOG` env var, the `_monom_log` shell helper, and the `internal/debuglog` Go package — the infrastructure that any layer can call into.

### Modified Capabilities

## Impact

- New file: `src/monom` gains `_monom_log()`
- New package: `internal/debuglog/debuglog.go`
- No changes to any public CLI contract (stdout, stderr, exit codes)
- No new dependencies
- Both helpers are strict no-ops when `MONOM_DEBUG_LOG` is unset

