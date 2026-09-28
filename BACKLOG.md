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

## Scaffolding

Zero-config discovery removed the first cold-start hurdle: a project needs no config file, and a declarative `monom` file can declare commands and settings without writing any code. What is left is creating things.

Sketch: `mnmd init` (write an empty or commented declarative `monom` file) and `mnmd new command <path>` (generate a script with a shebang in bash, python, or node, so it passes the discovery gate). Open question: whether scaffolding should also offer a hook-script template for authors who outgrow declarations.

---

## Delete `_archive/`

`_archive/` holds the prototype codebase, kept only so prototyping decisions stay inspectable. It is not functional, not maintained, and not tested. It should be deleted once the `mnmd` implementation is complete and stable — see `_archive/README.md`.
