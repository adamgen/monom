## Context

`monom` auto-discovers a single project root by walking up from `$PWD` (`mnmd root`), and `_setup_monom()` **exports** `_MONOM_PROJECT_ROOT`, short-circuiting on it thereafter. Completion is bound to a literal command name (`complete -F _monom_completion monom` / `compdef _monom monom`), and `mnmd()` is already a thin shell-function wrapper over the binary. `architecture.md` reserves a `make_monom_alias` placeholder for this feature and states the intent: push as much as possible into Go.

Constraints that shape this design:
- **Go owns logic, shell owns surface** — shell may only do what is technically impossible in Go (sourcing into the parent shell, registering completion hooks, exec).
- **Minimize subprocess roundtrips** — especially anything on a hot path or repeated per shell startup.
- **Completion must feel instant** — the completion path cannot regress.
- The **required user-config interface is amendment-protected** — this change does not touch it.

## Goals / Non-Goals

**Goals:**
- Bind a chosen command name to a fixed monom project root via `mnmd alias <name> <path>`.
- An aliased command behaves identically to `monom` run from inside that project, but with the root pinned rather than discovered.
- Multiple aliased projects coexist in one shell session without interfering with each other or with bare `monom`.
- The alias is usable immediately in the shell that created it — no re-source, no new terminal.
- Persistence adds **zero subprocess spawns** at shell startup and reuses the existing `mnmd install` rc-writing machinery.

**Non-Goals:**
- `mnmd alias` update / remove / list subcommands — deferred to a follow-up change.
- A central alias registry file, per-project alias files, or making `monom` itself reserve any subcommand.
- Shells other than bash and zsh.

## Decisions

### 1. An alias is `monom` bound to a fixed root, behaviorally identical
An aliased command reuses the full `monom` execution and completion behavior; only the project root differs (pinned, not discovered). *Alternative:* a distinct alias UX — rejected; it adds surface with no benefit and violates the "seamless, native-feeling CLI" goal.

### 2. Pin the root with `local`, never by exporting
Each alias binds its root as `local _MONOM_PROJECT_ROOT="<path>"` inside both its command function and its completion wrapper. bash/zsh dynamic scoping makes the pinned value visible to `_setup_monom` and its callees **for the duration of that call only**, then it vanishes. *Alternative:* exporting the global at alias time — rejected; it poisons bare `monom` auto-discovery (the `_setup_monom` short-circuit) and makes two aliases clobber each other. The `local` is applied on **both** paths because the completion path calls the registered handler directly, never the alias command function — so the handler must pin the root itself.

### 3. Alias management lives on `mnmd`, not `monom`
The command is `mnmd alias <name> <path>`. *Alternative:* `monom alias …` — rejected; `monom`'s entire argument surface belongs to the user's command tree, so reserving `alias` there would make any project's top-level `alias` command unreachable and break the clean `monom` = user-surface / `mnmd` = machinery split. Placing it on `mnmd` mirrors `mnmd install`.

### 4. The `mnmd()` shell wrapper intercepts `alias` to bind in-shell immediately
Because `mnmd()` runs in the parent shell, it can define the alias function + completion right now. *Alternative:* a pure binary — rejected; a binary cannot define a function in the parent shell, so the alias would not exist until the next shell. Interception gives immediate availability and keeps the user-facing surface to a single verb. Everything except `alias` still falls through to the binary unchanged.

### 5. Persist a define-only line, written once by the binary
`mnmd alias` calls `mnmd alias-save` (Go) to append a single `_monom_bind_alias <name> "<path>"` line to the rc file, idempotently, reusing install's rc machinery. At each shell startup that line runs the source-time helper — **no subprocess**. *Alternative A:* write a literal multi-line function block into rc — rejected; not a simple call and goes stale if monom internals change. *Alternative B:* `eval "$(mnmd alias --emit …)"` per alias per startup — rejected; N subprocess spawns on every shell open, directly against "minimize subprocess roundtrips."

### 6. Per-alias generated function + completion, not a shared registry
Each alias gets its own `<name>()` and `_<name>_complete`, both binding to the shared cores. *Alternative:* one shared handler plus a name→root associative array keyed on `COMP_WORDS[0]` — rejected; it still needs a per-alias function and per-alias `complete`/`compdef` registration (neither can bind multiple names), so it saves nothing while adding a stateful shell array and load-order coupling.

### 7. The residual `eval` is surface, not logic
Defining a function whose name is dynamic (`<name>`) requires `eval` of a constructed string in bash/zsh; there is no literal-free alternative except writing the block into rc (rejected in #5). This `eval` only *defines a function and registers a completion hook* — work the constitution explicitly assigns to shell. All behavior-deciding logic remains in Go (`_monom_run` → `mnmd pack`, `_monom_complete` → `mnmd filter`).

### 8. Extract `internal/rc` shared by `install` and `alias-save`
Shell detection, rc-file selection, the idempotency check, and the trailing-newline-safe append currently live unexported in `internal/install`. They move to `internal/rc` so both subcommands share one owner. *Alternative:* export install internals or duplicate them — rejected; two owners of rc-file logic invites drift.

### 9. Validate the root at alias time, and fail cleanly if it later moves
`mnmd alias` verifies `<path>` is a directory containing an executable `monom` config (the same definition `mnmd root` uses) before writing, failing fast. Because a pinned root can still move later, invoking an alias whose config is gone fails with a clear error and execs nothing.

## Risks / Trade-offs

- **`eval` of an attacker-controlled name/path** → `mnmd alias` restricts `<name>` to a shell-identifier-safe charset and rejects others; the path is validated and quoted in the generated line. Names/paths are author-supplied at setup, not user-supplied at runtime.
- **Load order: a bind line runs before `src/monom` is sourced** → install's `source` line always precedes alias lines (install runs before alias); `alias-save` can additionally verify the source line is present and error otherwise.
- **A pinned root is moved or deleted after aliasing** → invocation validates and emits a clear error instead of a cryptic exec failure; the user re-runs `mnmd alias`.
- **No remove/update yet** → the rc line is a single readable line the user can delete by hand; a follow-up change adds `mnmd alias rm/ls`.
- **Name collides with an existing command or function** → the binding overwrites it; `alias-save` prints a notice when replacing an existing alias line for the same name.

## Migration Plan

Purely additive. Refactoring `monom()` into `_monom_run monom` and the completion handlers into `_monom_complete` / `_monom_complete_zsh` cores is behavior-preserving and covered by the existing `shell-binding-*` scenarios. Rollback is removing `internal/alias`, the `alias-save` dispatch, and the `mnmd()` interception, and reverting the core extractions.

## Open Questions

- Exact alias-name charset — lean `[A-Za-z_][A-Za-z0-9_-]*`, rejecting anything else at `mnmd alias` time.
- Whether replacing an existing same-name alias should warn versus silently update — lean idempotent-replace with a printed notice.
