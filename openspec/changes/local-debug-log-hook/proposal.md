## Why

Debug logging today is global: a user exports `MONOM_DEBUG_LOG=/path` and every monom project in that shell logs to the same file. There is no way for a single project to opt into debug logging on its own (e.g. while developing that project's command tree) without flipping the global switch — and no way to route a project's debug output to a project-local file. A `debug` hook on the `monom` config file lets a project define its own log path, overriding the global setting for that project only.

## What Changes

- Add a **`debug` hook** to the user config interface: `<monom config file> debug` MAY print a single absolute file path on stdout. It is optional, like the existing `run` hook.
- `_setup_monom` resolves the effective debug-log path once per invocation: it calls `_monom_cfg debug`, validates the output (single-line path, writable), and exports `MONOM_DEBUG_LOG` when the hook path is usable (**local overrides global**). Multiline hook output or an unwritable hook path falls back to the inherited global `MONOM_DEBUG_LOG`. If the hook is absent or prints nothing, the global value is left untouched.
- No new env var and no changes to the read sites: both the shell `_monom_log` and the Go `debuglog.Log` continue to read `MONOM_DEBUG_LOG`. mnmd inherits the resolved value because the shell exports it before invoking the binary. This keeps both layers in agreement with zero duplicated resolution logic.
- The resolution costs **one** extra subprocess (`_monom_cfg debug`) per invocation, for **every** project — the same cost model as the existing `run` hook. The config file is opaque, so the only way to learn whether it defines `debug` is to run it; detecting the hook *is* the spawn. Projects without a `debug` hook still pay the one spawn but are unchanged in behavior: it returns empty and the global `MONOM_DEBUG_LOG` (set or unset) stands.

## Capabilities

### New Capabilities

_None._ The `debug` hook is an optional hook on the existing user config interface (the same place the `run` hook lives), so it extends existing capabilities rather than introducing a new spec.

### Modified Capabilities

- `shell-binding-core`: Add a requirement that `_setup_monom` consults the optional `debug` hook and exports `MONOM_DEBUG_LOG` (local-overrides-global) when it returns a path.
- `debug-log`: Document that the effective log path is the project-local `debug` hook output when present, otherwise the global `MONOM_DEBUG_LOG`; the existing read-site behavior is unchanged.

## Impact

- `src/monom` — `_setup_monom` gains a single `_monom_cfg debug` query and a conditional `export MONOM_DEBUG_LOG`. Built on top of [`underscore-prefix-internals`](../underscore-prefix-internals/proposal.md) (uses `_setup_monom`, `_monom_cfg`, `_MONOM_*`); this change is the planned project-local override for the `MONOM_DEBUG_LOG` that `underscore-prefix-internals` keeps unprefixed in its `debug-log` capability.
- `architecture.md` — document the `debug` hook under Hooks, alongside `run`; note the one-spawn-when-present cost and local-overrides-global precedence.
- `tests/` — shUnit2 coverage: hook present (local path used), hook absent (global used), hook empty (global used), hook multiline (global used + diagnostics), hook unwritable (global used + diagnostics), and confirmation that hook-less projects are unchanged.
- No Go code changes: `debuglog.Log` already reads `MONOM_DEBUG_LOG`; mnmd inherits the exported value.
- `MONOM_DEBUG_LOG` stays unprefixed and public, consistent with `underscore-prefix-internals`.
