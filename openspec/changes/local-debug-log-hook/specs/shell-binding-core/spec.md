## ADDED Requirements

### Requirement: _setup_monom resolves a project-local debug log path via the debug hook
`_setup_monom()` SHALL consult the optional `debug` hook on the `monom` config file to resolve a project-local debug log path. After the project root and `_MONOM_USER_CONFIG` are known, it SHALL preserve the inherited `MONOM_DEBUG_LOG`, call `_monom_cfg debug` once, and resolve the effective path as follows:

1. If hook stdout is **multiline**, treat it as invalid hook output: discard it, emit a diagnostic (Tab: `_monom_log` when an effective log path exists, else silent; command: stderr warning), and fall back to the inherited global value.
2. If hook stdout is a **single non-empty line** and the path is **writable**, export `MONOM_DEBUG_LOG` to that path for the current invocation, overriding the inherited global (**local overrides global**).
3. If hook stdout is a **single non-empty line** but the path is **not writable**, emit a diagnostic (same Tab/command split as above), and fall back to the inherited global value.
4. If the hook is **absent or prints nothing**, leave `MONOM_DEBUG_LOG` unchanged.

The query SHALL add at most one subprocess per invocation and SHALL NOT change observable behavior for a config file that does not expose `debug` (the call returns empty and the global value, set or unset, stands). Diagnostics on the command path SHALL NOT prevent the command from running. The resolved value is exported so that both the shell `_monom_log` and the `mnmd` binary's `debuglog.Log` read the same effective path without any change to their read sites.

#### Scenario: Local debug path overrides global
- **WHEN** `_setup_monom` runs, the global `MONOM_DEBUG_LOG` is set to `/tmp/global.log`, and `_monom_cfg debug` prints a writable `/proj/.monom-debug.log`
- **THEN** `MONOM_DEBUG_LOG` is exported as `/proj/.monom-debug.log` for the current invocation and subsequent shell and `mnmd` debug lines are written there

#### Scenario: Local debug path applies even when global is unset
- **WHEN** `_setup_monom` runs, `MONOM_DEBUG_LOG` is unset, and `_monom_cfg debug` prints a writable `/proj/.monom-debug.log`
- **THEN** `MONOM_DEBUG_LOG` is exported as `/proj/.monom-debug.log` and debug logging is active for this project

#### Scenario: No debug hook leaves global untouched
- **WHEN** `_setup_monom` runs and the config file does not expose `debug` (so `_monom_cfg debug` produces no output)
- **THEN** `MONOM_DEBUG_LOG` is left exactly as inherited (set or unset) and behavior is identical to before this change

#### Scenario: Empty debug hook output falls back to global
- **WHEN** `_setup_monom` runs, the global `MONOM_DEBUG_LOG` is `/tmp/global.log`, and `_monom_cfg debug` prints an empty string
- **THEN** `MONOM_DEBUG_LOG` remains `/tmp/global.log`

#### Scenario: Multiline hook output falls back to global on Tab
- **WHEN** `_setup_monom` runs on the Tab-completion path, the global `MONOM_DEBUG_LOG` is `/tmp/global.log`, and `_monom_cfg debug` prints multiple lines
- **THEN** `MONOM_DEBUG_LOG` remains `/tmp/global.log`, no stderr is written, and a diagnostic line is recorded via `_monom_log`

#### Scenario: Multiline hook output warns on command path
- **WHEN** `_setup_monom` runs on the command path, the global `MONOM_DEBUG_LOG` is `/tmp/global.log`, and `_monom_cfg debug` prints multiple lines
- **THEN** `MONOM_DEBUG_LOG` remains `/tmp/global.log`, a warning is written to stderr, and the invoked command still runs

#### Scenario: Unwritable hook path falls back to global
- **WHEN** `_setup_monom` runs, the global `MONOM_DEBUG_LOG` is `/tmp/global.log`, and `_monom_cfg debug` prints a single-line path that is not writable
- **THEN** `MONOM_DEBUG_LOG` remains `/tmp/global.log` and debug lines are written there when writable
