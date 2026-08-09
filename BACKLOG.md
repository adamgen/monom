# Backlog

Intended work, in no particular order. Each entry is a statement of intent and the reason it matters — not a plan. Design happens when the item is picked up, and lands in `architecture.md` (the contract) plus a co-located `TRADEOFFS.md` (the contested calls). See `CLAUDE.md` for how documentation works in this repo.

Nothing here is committed to. Deleting an entry is a valid outcome.

---

## `mnmd args` — flag parsing for command scripts

`architecture.md` lists `mnmd args` as a subcommand with a TBD output format. Today a CLI author writing a command script hand-rolls `getopts` or ad-hoc parsing, differently in every script and every language.

Sketch: `mnmd args [modifiers...] <flag> -- <raw args...>`, reading both `--prop=value` and `--prop value`, with `--boolean` (presence check, honoring `--no-<flag>`) and `--short <char>` modifiers. Consumed as `PROP=$(mnmd args prop -- "$@")`.

Open: whether one-flag-per-invocation is acceptable, or whether the shape needs to hand back several values at once.

---

## Shell helpers for CLI authors

Follows `mnmd args`. Two functions sourced alongside `monom()`:

- `monom_args` — declare and parse several flags in one call, setting variables in the caller's scope. Removes the one-subprocess-per-flag cost of calling `mnmd args` repeatedly.
- `monom_error` — print to stderr and exit with a code, replacing the `echo >&2; exit 1` boilerplate every author writes.

Constitutional tension worth resolving before starting: both are shell functions containing logic, which the "Go owns logic, shell owns surface" principle forbids. The case for them is that they run in the *author's* script, not monom's own path, and setting variables in the caller's scope is genuinely impossible from Go. That argument needs to be made explicitly and recorded, or the items dropped.

---

## Tab-completion tests driven by a real PTY

monom's entire value proposition is tab completion, and nothing verifies it end to end in a real shell. The existing shUnit2 suites test the pieces the completion path is built from, not the Tab press itself.

Two approaches were explored: a Go harness using `github.com/creack/pty` that spawns `bash --norc -i`, sources a script, sends keystrokes including Tab, and asserts on the output; and an `expect` script doing the same. Cases that matter: single-Tab unambiguous completion, double-Tab listing, full candidate listing.

Costs: a new Go dependency, a hard requirement on `bash` being present, and — for the `expect` variant — the `expect` binary. Purely additive; touches no existing source.

---

## Managed projects and scaffolding

The largest open item, and the only one that would amend the constitution.

The cold-start problem: a new user must understand the `complete`/`run` interface, write discovery logic, and lay out a directory tree before anything works. That friction contradicts "no boilerplate, no registration."

Sketch: introduce **managed projects** (a declarative `monom.yaml`, with `mnmd` handling discovery and resolution internally — fewer subprocess roundtrips) alongside today's **custom projects** (an executable `monom` implementing the hook interface). Add `mnmd init` and `mnmd new command`, generating scripts in bash, python, or node. `monom.yaml` could optionally delegate specific operations to an executable, giving a hybrid.

Touches nearly everything: new terminology, a new constitutional principle, `mnmd root` learning a second project marker, and shell bindings that branch on managed vs. custom. Open question: whether the config should expose anything beyond `complete`/`run` to make scaffolding aware of project structure.

---

## Rich command nodes (command-as-folder)

Today a command is a single executable file, which leaves nowhere to hang per-command richness. The intent: a directory containing an executable `run` file resolves as a runnable command instead of a group, with sibling files adding capability — most valuably a per-command `complete` script so completion continues past the leaf into flags and arguments. Single-file commands stay the zero-ceremony default; the upgrade path is `git mv deploy deploy/run`.

This makes `run`, `complete`, `pre-run`, `post-run` a reserved-name protocol at every tree level (discovery skips them, `mnmd check` flags collisions). The command map already mirrors the node model — a JSON node with a `"run"` key is a command — and reserves the same keys, so the two serializations stay unified. Open question, deliberately shared with the map: whether a node can be a command and a category at once (today the map rejects it).

---

## Lifecycle hooks (`pre-run` / `post-run`)

Hook files that run around command execution, cascading project → category → command on the way in and in reverse on the way out (onion order). The category level is what earns it: a monorepo's `project1/pre-run` activates a toolchain once instead of being copied into every command. Absent file = skip layer, zero ceremony, consistent with the constitution's hook principle.

Settle the runner semantics before starting: any `post-run` in the chain makes `exec` impossible (something must stay resident as parent), which changes signal handling and exit-code flow. Hook existence is a `stat`, so `pack` can detect "plain path, just exec" along its existing walk and projects without lifecycle hooks keep today's exact behavior.

---

## Delete `_archive/`

`_archive/` holds the prototype codebase, kept only so prototyping decisions stay inspectable. It is not functional, not maintained, and not tested. It should be deleted once the `mnmd` implementation is complete and stable — see `_archive/README.md`.
