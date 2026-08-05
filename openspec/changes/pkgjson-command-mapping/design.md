## Context

monom already has table-driven **command mapping**: `mnmd resolve-run` / `mnmd resolve-complete` implement the `run` and `complete` hook contracts from a static `<value> <key words...>` table, built for legacy migration. That machinery assumes a mapping value is an executable file path — `mnmd pack` validates it and the shell `exec`s it, with no arguments (`( exec "$pack_out" )` in `_monom_run`).

package.json scripts break that assumption twice:

1. **A script is not a file.** It runs as `<package-manager> run <script>` in the package's directory. There is nothing on disk to resolve to.
2. **The exec chain carries no arguments.** `mnmd pack` joins *all* its tokens with `/` into one path, and the shell execs exactly one word. "Executable + args" is currently inexpressible end to end.

There is also a composition constraint: the projects that need this are often mid-migration, so one config declares legacy file mappings *and* package scripts. Whatever mechanism exposes scripts must coexist with `resolve-*` in the same `run` and `complete` hooks without ceremony.

So this change has two halves: teach the existing mapping table a manifest-driven entry kind (`@pkgjson`), and make a minimal, backward-compatible widening of the pack→exec contract so a `run` hook can resolve to an executable *invocation* rather than only a bare file.

## Goals / Non-Goals

**Goals:**

- A CLI author exposes a JS package's scripts as monom commands by adding one line to a mapping table they may already have — same one-line-per-hook UX as today's `resolve-*`.
- Legacy file mappings and package scripts coexist in one table, one `mnmd` call per hook; precedence between them is simply entry order.
- Monorepo-first: one directive line per package, each under its own command prefix (`web build`, `api test`).
- Discovery, completion, group listing, and run all behave identically to tree commands — the scripts ride the existing pipelines.
- No constitution amendment: the required user-config interface (`complete`) is untouched; hooks and `pack` evolve in `architecture.md`.
- All new logic in Go; shell changes limited to the technically-unavoidable exec surface.

**Non-Goals:**

- Passing user-typed arguments through to scripts (`monom web test --watch`). No monom command receives args today; that is the future `mnmd args` change. The args-passthrough added to `pack` here carries only *hook-produced* tokens.
- Workspace auto-discovery (globbing `workspaces` patterns). Natural follow-up; the table keeps v1 explicit.
- Dependency/lifecycle management (`npm install`, engines checks).
- Shells beyond bash/zsh; package managers beyond npm/pnpm/yarn/bun.

## Decisions

### D1: A mapping-table directive, not a separate subcommand pair

The mapping table consumed by `resolve-run`/`resolve-complete` gains a second entry kind. Lines whose first field is exactly `@pkgjson` are manifest directives; everything else parses as today:

```
<value> <key words...>                  # file mapping (unchanged)
@pkgjson <package.json-path> [prefix]   # manifest mapping
```

A directive expands **in place** to one entry per script in that manifest: key = `<prefix-words> <script>` (top-level when prefix is omitted), value = the package-manager invocation (D2). `prefix` is a single slash-delimited token (`web`, `apps/web`). Paths are absolute or project-root-relative. A mid-migration config becomes:

```sh
root="$(cd "$(dirname "$0")" && pwd)"

mapping="
$root/scripts/migrate_db.sh    db migrate
@pkgjson $root/apps/web/package.json   web
@pkgjson $root/apps/api/package.json   api
"

case "$1" in
  complete) mnmd resolve-complete "$mapping" ;;
  run)      shift; mnmd resolve-run "$mapping" "$@" ;;
esac
```

- **Why not a separate `pkgjson-run`/`pkgjson-complete` pair** (the initially drafted shape): both pairs would be passthrough-on-miss, so a config using both must chain them — `mnmd pkgjson-run "$pkgs" $(mnmd resolve-run "$legacy" "$@")` — which is fragile (a second unquoted re-split), doubles `mnmd` spawns on the run path *and on every Tab press* (the constitution's subprocess-minimization principle names exactly this cost), and makes legacy-vs-script precedence an emergent property of nesting order instead of a visible table ordering. The directive keeps "one mapping table drives both hooks" literally true.
- **First match wins across both entry kinds, in table order** — one rule, same as today's `resolve-run`, now also expressing migration precedence (put the legacy line first to keep a legacy name shadowing a same-named script, or vice versa).
- **Backward compatible**: `@pkgjson` in the value position was previously a dead line (an `@pkgjson`-named executable could never resolve); no existing table changes meaning.
- Alternative rejected — *a `pkgjson-table` expander that prints resolve-format lines for inclusion in a table*: resolve values are single whitespace-free tokens, and a script's value is a multi-word invocation. The format cannot carry it; the expansion must happen inside the parser where values are structured data.

### D2: A directive hit resolves to "executable + args" via a pack contract extension

On a directive hit, `resolve-run` prints a space-separated exec line whose first token is the package manager's **absolute** path (resolved via `$PATH` lookup at run time), e.g.:

```
/usr/local/bin/npm --prefix /repo/apps/web run build
```

`monom()`'s existing hook plumbing re-splits this on whitespace and hands it to `mnmd pack`. Pack's new rule: **when the first token is an absolute path and more tokens follow, the rest are the command's arguments** — pack validates the first token exactly as today (exists, executable, not a directory) and prints all tokens **one per line**. `_monom_run` splits pack's output on newlines and execs the argv array.

File-mapping hits are untouched: their value is a single token, pack resolves it as today, and one output line reaches the shell.

Alternatives rejected:

- **Generated shim files** (materialize each script as an executable, either synced ahead of time or written on demand): keeps pack single-token but adds filesystem side effects to the run path, staleness/cleanup obligations, and violates pack's pure-resolver spirit ("returns a path xor a signal").
- **Discovery-only** (author writes the run side): doesn't deliver the feature — running is the point.
- **Shell composes `npm run` itself**: logic in shell; there is a technical reason it *can* be in Go, so per the constitution it must be.

The extension is strictly gated: relative-first-token input behaves byte-for-byte as today (all tokens joined with `/`), and absolute single-token input prints one line as today. Only the previously-nonsensical case (absolute first token + extra tokens, which today resolves to garbage like `/usr/bin/npm/run/build` and fails) gains meaning.

### D3: Pack output is newline-delimited tokens, not a space-separated line

First line = executable path, subsequent lines = args. Two reasons:

- **Backward compatibility is free**: single-command output (one path + `\n`) is byte-identical to today.
- **Spaced project roots keep working**: today `( exec "$pack_out" )` quotes the whole path, so a project root containing spaces works. Splitting pack output on whitespace would regress that; splitting on newlines preserves it (the path occupies the whole first line).

Shell split idiom (works in bash 3.2 and zsh — command substitution is IFS-split in both, unlike unquoted parameter expansion which zsh does not split):

```sh
local IFS=$'\n'
# shellcheck disable=SC2207
pack_words=( $(printf '%s' "$pack_out") )
( exec "${pack_words[@]}" )
```

This mirrors the existing `hook_out` re-split idiom in `_monom_run`, including its known glob-expansion caveat (paths containing `*`/`?` are already out of scope there).

### D4: Package manager detection via the `packageManager` field, defaulting to npm

Per directive: read the manifest's `packageManager` field (`"pnpm@9.1.0"` → `pnpm`); absent → `npm`. Recognized: `npm`, `pnpm`, `yarn`, `bun`. The binary is resolved to an absolute path with a `$PATH` lookup at `resolve-run` time, on a hit only — never at completion time, so Tab never depends on the PM being installed.

Invocation forms (directory pinned via the PM's own flag, so no `cd` and no cwd logic in shell):

| PM   | Exec line                          |
| ---- | ---------------------------------- |
| npm  | `npm --prefix <dir> run <script>`  |
| pnpm | `pnpm -C <dir> run <script>`       |
| yarn | `yarn --cwd <dir> run <script>`    |
| bun  | `bun --cwd <dir> run <script>`     |

`<dir>` is the absolute directory containing that directive's package.json.

- Alternative rejected — *lockfile sniffing*: more magic, ambiguous in monorepos (lockfile lives at the workspace root, not next to the package), and corepack has made `packageManager` the authoritative declaration.
- Unrecognized `packageManager` value → hard error on run (loud, names the value); completion still works (discovery never needs the PM).

### D5: Error posture differs per hook path, matching monom's diagnostics philosophy

- **`resolve-complete`** (runs on the Tab path): never breaks discovery. A directive whose manifest is unreadable or invalid contributes nothing, emits a stderr warning (invisible mid-Tab, visible when the author runs the config manually), and the subcommand still exits 0 so every other entry survives. Script names containing whitespace or `/` are silently skipped — they cannot be represented in the slash-delimited discovery format (same posture as `filter`'s silent skip). File-mapping entries behave exactly as today.
- **`resolve-run`** (user explicitly invoked a command): loud. If the typed words *shape-match* a directive (its prefix words + exactly one more word) but the manifest is unreadable, that is a hard error — falling through to tree resolution would produce a baffling "command not found" for a command that tab-completed moments ago. Same for PM-not-on-PATH, unrecognized PM, and whitespace in the package directory or script name (whitespace cannot cross the space-split hook-stdout boundary). A genuine miss stays a passthrough non-event, exactly as today. Directive manifests are read lazily — only when the typed words shape-match — so a table full of directives costs nothing on a legacy-entry hit.
- All errors use the existing `CodedError` registry; no new exit codes.

### D6: Manifest script order is preserved

Scripts are emitted in the order they appear in the file (via `json.Decoder` token walking in the new `internal/pkgjson` helper package, not a Go map, which would be nondeterministic). The author's manifest ordering is intentional, and stable output keeps completion deterministic and testable. Manifest parsing and the PM table live in `internal/pkgjson`; `internal/resolve` consumes it, keeping each package's testing surface small.

## Risks / Trade-offs

- [PM flag semantics vary — e.g. `npm --prefix run` has historical quirks, yarn berry vs classic `--cwd`] → The per-PM invocation table is centralized in one Go function; the e2e fixture actually executes a script through the full `monom` path with a stub PM on `$PATH`, and the exact flags are verified against real PMs during implementation. Adjusting a flag is a one-line change + test update.
- [Whitespace in package directory paths is unsupported on the run path] → Inherited from the hook stdout contract (space-separated words), same stance as file-mapping values ("can never contain whitespace"). Mitigation: `resolve-run` fails loudly with a message naming the offending path instead of mangling it silently.
- [resolve-* loses purity — directive entries do file IO] → Scoped: file-mapping-only tables remain pure lookups with identical behavior; IO happens only for directive lines, lazily on the run path. The distinction is documented in `architecture.md`.
- [Widening pack's contract could ripple] → The new branch is unreachable from tree resolution (tree tokens are never absolute) and from every existing documented producer. Existing single-token behavior is locked by unchanged e2e assertions.
- [Unquoted command-substitution split in `_monom_run` glob-expands tokens containing `*`/`?`] → Pre-existing hazard of the established hook-split idiom; kept consistent rather than introducing a second idiom. Documented inline.
- [Completion silently omits scripts from a broken manifest] → stderr warning surfaces when the config is run by hand; a future `mnmd check` extension (validating directive manifests) is noted as follow-up, not blocking.

## Migration Plan

No migration. The change is purely additive at every surface: a table line shape that previously could never resolve, a pack input shape that previously always failed, and a shell split that is a no-op for single-line pack output. Rollback = revert the commit.

## Open Questions

- Exact cwd flag for yarn berry (`--cwd` was yarn-classic; berry may need the script run via `yarn workspaces` or plain `--cwd` still works) and bun (`--cwd` position) — verify against installed PMs during implementation; the spec pins npm/pnpm forms, and the per-PM table centralizes any correction.
- Should `mnmd check` learn to validate mapping tables (manifest readable, script names representable)? Deferred to a follow-up change.
