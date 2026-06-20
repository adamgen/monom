## 1. Prerequisite

- [ ] 1.1 Implement on top of `underscore-prefix-internals` (uses `_setup_monom`, `_monom_cfg`, `_MONOM_USER_CONFIG`). If that change is not yet applied, rebase the identifier names accordingly.

## 2. Shell

- [ ] 2.1 In `src/monom`, extend `_setup_monom` so that after the project root and `_MONOM_USER_CONFIG` are resolved, it preserves the inherited `MONOM_DEBUG_LOG`, calls `_monom_cfg debug` once, and resolves the effective path per the resolution ladder (multiline → discard; single-line writable → export; single-line unwritable → fall back to inherited; empty/absent → leave inherited). Place the query after the short-circuit and discovery branches so it runs on every successful setup (completion and execution paths).
- [ ] 2.2 Suppress hook stderr (`2>/dev/null`). Reject multiline hook stdout before the writability check.
- [ ] 2.3 Probe writability of a single-line hook candidate before exporting; if not writable, fall back to the inherited global path.
- [ ] 2.4 Diagnostics: Tab path — never stderr; record multiline/unwritable-hook fallback via `_monom_log` when an effective log path exists. Command path — stderr warning for multiline or unwritable hook path; command still runs.
- [ ] 2.5 Add a `_monom_log` line recording the resolved effective debug path when logging is active.
- [ ] 2.6 Confirm no second read site is introduced: `_monom_log` still reads `MONOM_DEBUG_LOG`; no Go changes.

## 3. Documentation

- [ ] 3.1 In `architecture.md`, add a `### Hook: debug — project-local debug log path` subsection under Hooks (alongside `run`): input none, output a single absolute path or nothing, validation (single-line, writable), fallback to global `MONOM_DEBUG_LOG` on invalid or unwritable output, precedence local-overrides-global when usable, cost one unconditional spawn per invocation plus one writability check when the hook prints a path.
- [ ] 3.2 In the `architecture.md` Environment Variables note for `MONOM_DEBUG_LOG`, mention that a project may override it via the `debug` hook when the hook path is valid and writable.

## 4. Tests

- [ ] 4.1 shUnit2: `debug` hook prints a writable path → that path is used; assert the file receives lines from both shell and `mnmd` (global set to a different path, confirm local wins).
- [ ] 4.2 shUnit2: `debug` hook prints a writable path while global `MONOM_DEBUG_LOG` is unset → local path is used (logging active).
- [ ] 4.3 shUnit2: no `debug` hook → global `MONOM_DEBUG_LOG` behavior unchanged (used when set, no I/O when unset).
- [ ] 4.4 shUnit2: `debug` hook prints empty → falls back to global value.
- [ ] 4.5 shUnit2: `debug` hook prints multiline output → falls back to global; Tab records diagnostic via `_monom_log` when global is set; command path emits stderr warning and still runs the command.
- [ ] 4.6 shUnit2: `debug` hook prints a single-line unwritable path → falls back to global when set; same diagnostic split as 4.5.
- [ ] 4.7 Use fixture `monom` config files exposing `debug` (writable, unwritable, multiline, empty) and one without the hook (extend `tests/helpers`/fixtures as needed).

## 5. Validation

- [ ] 5.1 `shellcheck` on `src/monom` — clean, no new suppressions.
- [ ] 5.2 `go vet ./...` and `go test ./...` — pass (no Go changes expected, confirm nothing broke).
- [ ] 5.3 Run the affected shUnit2 suites — all pass.
- [ ] 5.4 Manually confirm: a project with a usable `debug` hook logs to its own file with the global unset; multiline or unwritable hook output falls back to global; a project without a hook is byte-for-byte unchanged.
