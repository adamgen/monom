## ADDED Requirements

### Requirement: Effective debug log path may be set per-project via the debug hook
The effective debug log path SHALL be resolved as follows: `_setup_monom` consults the `debug` hook once. If the hook prints a valid single-line path that is writable, that path is the effective `MONOM_DEBUG_LOG` for the invocation; if the hook output is multiline, empty, or a single-line path that is not writable, the inherited global `MONOM_DEBUG_LOG` (set or unset) is used. Resolution happens once in `_setup_monom`, which exports the result so both the shell `_monom_log` helper and the Go `debuglog.Log` helper read the same value. Neither read site changes: both continue to treat `MONOM_DEBUG_LOG` as the file path and remain no-ops when it is empty.

#### Scenario: Project-local log file receives both shell and Go debug lines
- **WHEN** a project's `debug` hook prints a writable `/proj/.monom-debug.log` and a Tab completion is triggered
- **THEN** `/proj/.monom-debug.log` contains debug lines from both the shell completion handler and the `mnmd` binary

#### Scenario: Global log path used when no debug hook is present
- **WHEN** `MONOM_DEBUG_LOG=/tmp/global.log`, the config file has no `debug` hook, and a Tab completion is triggered
- **THEN** debug lines are written to `/tmp/global.log` exactly as before this change

#### Scenario: Global log path used when hook path is unwritable
- **WHEN** `MONOM_DEBUG_LOG=/tmp/global.log`, the `debug` hook prints a single-line path that is not writable, and a Tab completion is triggered
- **THEN** debug lines are written to `/tmp/global.log`

#### Scenario: No file I/O when neither local nor global path is set
- **WHEN** `MONOM_DEBUG_LOG` is unset and the config file has no `debug` hook (or its `debug` hook prints nothing, multiline output, or an unwritable single-line path)
- **THEN** no file is created or written and behavior is identical to the production path
