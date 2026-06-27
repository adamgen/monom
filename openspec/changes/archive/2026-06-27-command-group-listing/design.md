## Context

`mnmd pack <word...>` joins the user's tokens with `/`, resolves the result against the project root, and prints the absolute executable path for the shell to `exec`. Today, if the resolved path is a directory, `pack` returns a generic error and `cmd/mnmd/main.go` prints `mnmd pack: <err>` to stderr and exits 1 — indistinguishable from "command not found" or "not executable". The `monom()` shell function forwards that raw stderr verbatim.

But a directory is not a failure: it is a **command group** (a noun in monom's noun→verb file tree). Per clig.dev, invoking a group should list available sub-commands, not emit creator-only output. And per clig.dev's "no catch-all subcommand", monom must not auto-pick a default leaf — that would be a time bomb when a real leaf is later added.

Two project principles shape *how* the listing is produced:
- **Go Owns Logic, Shell Owns Surface** — the decision *that this is a group* belongs in Go; only the final user-facing rendering belongs in shell.
- **Discovery is the `complete` hook's job** (terminology.md). The set of commands under a group is owned by `complete`, not re-derived elsewhere.

## Goals / Non-Goals

**Goals:**
- Turn "resolved path is a directory" from a generic error into a distinct, machine-readable **group signal** so `monom()` can branch without string-matching stderr.
- Print a concise, user-facing listing of the group's children (`available: cloud, local`) with no `mnmd pack:` prefix.
- Keep `complete` the single source of truth for the command tree — the group listing must match what `monom <group> <Tab>` shows.
- Preserve the existing escape hatch: authors make a group runnable via the `run` hook, not a monom-level flag.

**Non-Goals:**
- No recursive/tree listing — only the group's immediate children.
- No new top-level flag or config to make a namespace "also a command."
- No change to completion (`filter`) behavior or the required user config interface.
- No change to how leaf resolution, missing-file, or not-executable errors behave.

## Decisions

### Decision: A directory resolution is a typed signal, not a generic error

`pack.Pack` returns a typed sentinel `*pack.GroupError` carrying **only** the resolved path, alongside the existing `(string, error)` signature:

```go
// internal/pack/pack.go
type GroupError struct {
    Path string // absolute path of the resolved group directory
}
func (e *GroupError) Error() string { return "resolved path is a command group: " + e.Path }
```

Callers use `errors.As(err, &ge)` to detect the group case. `pack` stays a pure resolver: it returns an executable path **xor** signals a non-leaf outcome, and it never carries data alongside an error.

- **Alternative considered — keep a plain error and have main.go regex the message.** Rejected: stringly-typed control flow, fragile, and violates "Go owns logic."

### Decision: pack does NOT enumerate the group's children (single source of truth)

`pack` deliberately does not read the directory to list its children. The child listing is sourced by the caller from the canonical discovery pipeline (`complete | mnmd filter`). `GroupError` therefore has no `Children` field.

- **Alternative considered (rejected) — pack reads the directory and returns/​prints the sorted children.** This was the initial design. Rejected because it makes `pack` a *second discoverer* of the command tree, parallel to `complete`. The two can disagree: an author whose `complete` hides files or exposes a virtual/​remapped tree (which the `run` hook explicitly enables) would see `pack`'s raw directory listing diverge from completion. It also makes `pack` return error-with-data, breaking the "error xor value" shape. Sourcing children from `complete` keeps one source of truth, makes the `monom <group>` listing identical to `monom <group> <Tab>`, and lets `pack` stay a pure, side-effect-free resolver.
- **Consequence:** a directory that exists on disk but whose children `complete` does not surface lists nothing (reported as a group with no `available:` line) — which is *more* correct, since those entries are not commands in that project.

### Decision: Reserved exit code 3 is a payload-free signal from the binary

`cmd/mnmd/main.go` `runPack` maps the `GroupError` outcome to **exit code 3**, writing nothing to stdout and nothing to stderr. Rationale for 3: `0` = resolved leaf, `1` = generic/real error (not found, not executable, no root), `2` is conventionally reserved for CLI-usage/​misuse — so `3` is the lowest free code for this distinct, non-error signal.

- Exit 3 is reserved **exclusively** for the command-group outcome and is documented in `architecture.md` so the contract is discoverable.
- Because pack emits no payload, the signal is purely the exit code; the shell does the rest.

### Decision: `monom()` renders the message and sources children from `complete | filter`

On exit code 3, `monom()` prints a concise message to **stderr** and returns non-zero (the user did not run a command):

```
monom: 'infra' is a command group
available: cloud, local
```

- The group label is the **last token the user typed** (found by iterating the unpacked args array — works identically in bash and zsh without shell-specific indexing).
- The `available:` list comes from `monom_cfg complete | mnmd filter "${unpacked_args[@]}" ""`. The trailing `""` tells `filter` to drill into this level's children — the exact pipeline tab-completion uses. No `mnmd pack:` prefix appears.
- On exit 0 it execs the path (unchanged); on any other non-zero it forwards pack's stderr (unchanged).

The shell owns only formatting/surface; the *fact that it is a group* comes from Go (exit 3) and the *list* comes from `complete` (the discoverer). This respects both principles while keeping the message human-friendly.

### Decision: Empty / unlisted group still signals group, omits the available line

If `complete | filter` yields no children — an empty directory, or a group whose children `complete` does not surface — `monom()` still prints `monom: 'infra' is a command group` but omits the `available:` line. An empty group is still not a runnable command, so it must not fall through to "command not found."

### Decision: Author override stays in the `run` hook (no new surface)

No monom-level flag is added. Making a namespace runnable is an author concern, satisfied by the existing `run` hook (e.g. mapping `infra` → `infra cloud deploy`). `architecture.md`'s `pack` section and `run` hook note state this explicitly so the escape hatch is discoverable. This avoids the clig.dev catch-all time bomb.

## Risks / Trade-offs

- **[Exit code 3 collides with a future meaning]** → Reserve and document it in `architecture.md` as "pack: resolved to a command group"; an e2e test asserts the exact code so a regression is caught.
- **[A user's leaf and a group share a name at different levels]** → Not possible at one path: a single resolved path is either a file or a directory, never both. No ambiguity introduced.
- **[Extra `complete | filter` roundtrip on the group path]** → Acceptable: it occurs only on the cold/​error path (user typed an incomplete command), never on leaf execution or during completion, and it reuses the existing pipeline rather than adding new listing logic. The "minimize subprocess roundtrips" principle is about the hot paths.
- **[zsh vs bash capture of exit code + array iteration]** → `monom()` captures `mnmd pack`'s status and iterates the unpacked-args array for the last token; both behave identically in bash and zsh. Covered by e2e in `tests/monom_run_test`.

## Open Questions

- None outstanding. (The earlier question about filtering raw directory entries to executables/​dirs is moot under this design: children come from `complete`, which already defines the surface command set rather than raw filesystem contents.)
