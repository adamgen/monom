## Context

Debug logging is driven by a single env var, `MONOM_DEBUG_LOG`. Two read sites consult it independently: the shell helper `_monom_log` (`src/monom`) and the Go helper `debuglog.Log` (`internal/debuglog`). When it is a non-empty path, both append timestamped lines to that file; when unset, both are no-ops with no file I/O and (in the shell) no subprocess. The Go binary reads it via `os.Getenv` at the top of `main()` and throughout.

This change adds a project-local override: a `debug` hook on the `monom` config file, mirroring the existing optional `run` hook. The constitutional constraints in play are *Go owns logic, shell owns surface*, *minimize subprocess roundtrips*, and *pluggability via hooks*. The hook is resolved on both the completion (Tab) and execution (`monom <cmd>`) paths, so subprocess cost is a first-class concern.

This change is layered on top of `underscore-prefix-internals` and uses the post-rename identifiers (`_setup_monom`, `_monom_cfg`, `_MONOM_`*). `MONOM_DEBUG_LOG` itself remains unprefixed and public per that change.

## Goals / Non-Goals

**Goals:**

- Let a single project enable debug logging to a project-chosen path without touching the global `MONOM_DEBUG_LOG`.
- Keep the two read sites (`_monom_log`, `debuglog.Log`) unchanged — one resolution point, no duplicated logic, no risk of the layers disagreeing.
- Add exactly one subprocess per invocation — the same cost model as the existing `run` hook. There is no cheaper option: the config file is opaque, so the only way to learn whether it defines `debug` is to run `_monom_cfg debug` and see if it printed a path. Detecting the hook *is* the spawn.
- Reject unusable hook output (multiline) and unwritable hook paths without breaking completion or command dispatch; fall back to the inherited global `MONOM_DEBUG_LOG` when the hook path cannot be used.

**Non-Goals:**

- No new env var. The override reuses `MONOM_DEBUG_LOG`.
- No Go changes. The binary inherits the resolved value from the shell's environment.
- Not letting the local hook explicitly *disable* logging when the global is on (precedence is local-overrides-when-present; an absent or empty hook falls back to global — there is no "force off" sentinel in this change).
- Not adding a `debug` query to mnmd itself (that would spawn the config file from Go, duplicating the shell's resolution and adding a roundtrip).

## Decisions

### Decision: Resolve the path once in the shell, export `MONOM_DEBUG_LOG`

`_setup_monom` already runs once per `monom`/completion invocation and already knows the project root and config file. It is the single natural place to resolve the effective log path. Before overriding, it preserves the inherited global value; after consulting the hook it exports the resolved effective path:

```
_inherited_debug_log=$MONOM_DEBUG_LOG
candidate=$(_monom_cfg debug 2>/dev/null)
# reject multiline hook output (invalid path)
# if single-line and writable: export candidate (local overrides global)
# if single-line and not writable: fall back to _inherited_debug_log
# if empty/absent: leave _inherited_debug_log unchanged
```

Because `_setup_monom` runs before any `mnmd` invocation in the same shell call, exporting `MONOM_DEBUG_LOG` here means the binary inherits the resolved value with no Go changes. Both `_monom_log` and `debuglog.Log` keep reading the same variable, so the shell and Go layers are guaranteed to agree.

- **Alternative — resolve in Go (`mnmd` spawns `_monom_cfg debug`):** rejected. mnmd's hot path (`filter`) spawns no config subprocess today; adding one there violates *minimize subprocess roundtrips*, and it would duplicate the resolution the shell must do anyway for `_monom_log`.
- **Alternative — a second env var (e.g. `_MONOM_DEBUG_LOG_LOCAL`):** rejected. Two variables means two read sites per layer and a precedence check duplicated in shell and Go. Collapsing to one resolved `MONOM_DEBUG_LOG` keeps the read sites untouched.

### Decision: Local overrides global, hook is optional and queried once

`Precedence is local-wins: if _monom_cfg debug prints a non-empty path, it replaces MONOM_DEBUG_LOG for this invocation; otherwise the global value (set or unset) stands. This lets a project turn logging on even when the global is off — which is the whole point — so the hook cannot be skipped merely because the global is unset.`

The hook is optional and follows the established attempt-and-fallback discovery used by `run`: call it, use the output if usable, fall back otherwise. No registration step.

### Decision: One spawn per invocation, unconditionally

`_setup_monom` calls `_monom_cfg debug` once on every successful setup — for *every* project, whether or not it defines the hook. There is no way to spawn "only for projects that define the hook," because the config file is an opaque executable (shell, Python, Go, anything): the sole way to discover whether it exposes a `debug` subcommand is to run it and observe the output. Detection and the spawn are the same act.

A config file with no `debug` subcommand, invoked as `_monom_cfg debug`, exits non-zero / prints nothing, so the `if [ -n "$debug_path" ]` guard falls back to the global value with no behavioral change. The cost is therefore one unconditional `_monom_cfg debug` subprocess per invocation.

This is the same cost model the `run` hook already pays: `monom()` calls `_monom_cfg run "$@"` unconditionally and falls back when it produces nothing. So this change adds no new *kind* of cost — it adds one more hook spawn of the same form, on the same attempt-and-fallback discovery pattern, and that single spawn was explicitly accepted as acceptable on the completion path.

(If this spawn ever proves too costly on the Tab path, the cure is a convention that makes hook presence detectable *without* running the config file — e.g. a marker file, or `complete` output that declares available hooks. That is out of scope here and noted as a possible follow-up, not a promise of this change.)

### Decision: Hook-output validity vs path writability — separate layers, shared fallback

Hook-output **validity** and path **writability** are different checks and MUST NOT be merged into one "bad hook" notion:

- **Validity** — is `_monom_cfg debug` stdout usable as a path? Multiline output is invalid; empty or a single line is valid.
- **Writability** — can the resolved path receive append writes? Checked only after validity passes.

Resolution ladder (runs once in `_setup_monom`, on both Tab and command paths):

```
_inherited=$MONOM_DEBUG_LOG
candidate=$(_monom_cfg debug 2>/dev/null)
if candidate is multiline:           # hook OUTPUT invalid
    diagnose; discard candidate
if candidate is single-line nonempty:
    if candidate is writable: effective = candidate
    else: diagnose; effective = _inherited   # unwritable hook → fall back to global
else:
    effective = _inherited                     # absent/empty hook
export MONOM_DEBUG_LOG only when effective differs from current inherited handling
```

**Diagnostics** (debug logging is a side-channel, never a precondition):

- **Tab completion** — never write to stderr mid-Tab. Record multiline or unwritable-hook fallback via `_monom_log` when an effective log path is already defined; otherwise silent.
- **Command path (`monom <cmd>`)** — stderr warning for multiline hook output or unwritable hook path, then proceed; the command still runs.

Writability of the final effective path (whether from hook or global) is still handled at the read sites: `_monom_log` and `debuglog.Log` swallow open/write errors. The pre-resolution writability probe on the hook candidate exists to decide fallback, not to make logging safe.

## Risks / Trade-offs

- **One extra subprocess on the Tab-completion path, for every project** → Accepted. Bounded to exactly one spawn, centralized in `_setup_monom` (already on that path), and identical in form to the `run` hook's existing spawn. Projects without the hook still pay the spawn but are unaffected in *behavior* — it returns empty and falls back. There is no way to skip the spawn for hook-less projects without a separate presence-detection convention (out of scope).
- **Writability probe on the Tab-completion path** → Accepted. One filesystem check per invocation when the hook prints a single-line path, centralized in `_setup_monom`. Buys correct fallback to global when the hook path is unwritable; read sites still swallow errors on the final effective path.
- **Hook prints multiline output** → Invalid hook output; discard and fall back to inherited global. Tab: diagnostic via `_monom_log` when a log path is defined, else silent. Command: stderr warning, command proceeds.
- **Hook prints a single-line but unwritable path** → Fall back to inherited global `MONOM_DEBUG_LOG` so a writable global still logs. Same diagnostic split as multiline. If neither hook nor global is writable, read sites silently no-op as today.
- **Hook prints a bogus but writable path** → Local path wins; open/write failures at read sites are swallowed as today. `mnmd check` is the place to later add stricter validation if desired (out of scope here).
- **Interaction with a future "force-off" need** → Deliberately deferred (Non-Goals). If a project later needs to suppress logging while the global is on, that is a follow-up that defines an explicit sentinel; this change does not paint over it.

