# monom Constitution

> A constitution defines the invariants of a project — the principles that remain true as everything else changes, and that resolve conflicts before they are argued. It is normative, not descriptive: it says what should be, not what currently is. It does not contain implementation state, migration plans, or unsettled assumptions.
>
> This document governs the monom project. All architectural decisions and AI-assisted work must be validated against it. When in doubt, re-read this document.

---

## Spelling & Casing

The project name is **monom** — always lowercase, even at the start of a sentence. Never "Monom".

---

## Mission

monom is a CLI framework that turns a file tree into a tab-completable command tree. You organize your scripts in folders, add a `monom` config file that defines how to discover and run commands, and monom automatically gives you a full-featured CLI with tab completion — in any shell, for any script language.

The core promise: **your file tree is your command tree**. Folders become command categories. Scripts become commands. No boilerplate, no registration, no framework lock-in.

---

## Audience

There are two distinct roles:

**The CLI Author** — a developer building a CLI tool or monorepo manager using monom. They write the `monom` config file and the command scripts. They work against the monom interface.

**The CLI User** — a developer (often on the same team) who uses the CLI the author built. They type `my-tool <Tab>` and run commands. They never know or care that monom exists underneath.

monom must serve both: give authors a clean, minimal interface to implement; give users a seamless, fast, native-feeling CLI experience.

---

## Goals

1. Make it trivial to build a CLI tool or monorepo manager in any scripting language.
2. Provide free, correct tab completion for bash and zsh (with more shells supported in the future).
3. Support both standalone CLI tools (like `git`) and internal monorepo managers (like `acme ui build`).
4. Remain language-agnostic — commands can be shell, Python, Node, Ruby, anything with a shebang.
5. Follow [clig.dev](https://clig.dev/) guidelines for CLI behavior.
6. Be fast. Completion must feel instant. No perceptible latency from monom's own overhead.
7. Be testable. Every piece of logic has a clear home and a clear testing strategy.

---

## Principle: Go Owns Logic, Shell Owns Surface

This is the primary architectural principle.

**Go handles** all internal decision-making: project root discovery, calling the user's config file, completion filtering, command resolution, argument parsing, and all data transformation.

**Shell handles** only what is technically impossible to do in Go: sourcing into the parent shell session, registering completion hooks, and exec-ing the resolved command file.

**The test:** Before writing any shell code, ask — "is there a technical reason this cannot be in Go?" If the answer is no, it goes in Go.

---

## Principle: Minimize Subprocess Roundtrips

A chain of subprocesses passing data through pipes is hard to reason about, hard to debug, and hard to test in isolation. Each boundary should exist only because there is no alternative.

---

## Principle: Pluggability via Hooks

monom defines a small set of default behaviors (command discovery, path resolution, etc.). For each, the CLI author MAY expose a hook in their user config to intercept, transform, or extend the default. Hooks are discovered by convention: they are subcommands of `$MONOM_USER_CONFIG`. If a hook is absent, monom uses its default behavior.

The pluggability principle guides design: when monom adds a new default behavior, the question is always "should this be hookable?" Authors get a minimal required surface to implement and a clearly-defined seam for everything they want to customize.

The current hooks are documented in `architecture.md`.

---

## Principle: The Required User Config Interface Requires a Constitution Amendment to Change

The user config (`$MONOM_USER_CONFIG`) is the seam between monom and the author's project. The set of subcommands monom *requires* the user config to expose is part of monom's stability contract — projects that comply today must continue to work tomorrow.

The currently required interface is:

```
$MONOM_USER_CONFIG complete   # prints all discoverable command paths, one per line
```

Adding to this required set, removing from it, or changing the contract of any required subcommand requires an amendment to this document. Optional hooks are documented in `architecture.md` and may evolve there without amendment.

monom does not care how the required interface is implemented — shell, Python, Go, anything that prints to stdout.

---

## Principle: Testing Is Layered

Logic and surface are tested with different tools and must not be conflated:

- **Go unit tests** cover Go functions in isolation.
- **e2e tests** cover the full CLI surface — they spawn the binary, assert on stdout, stderr, and exit codes. These are the contract: they define what monom promises its users.
- **Completion e2e tests** execute in a shell environment as close as possible to where tab completion functions and bindings actually run.

If something needs a test and it can live in Go, it should. Shell test files are for CLI surface behavior, not internal logic.

---

## Principle: Static and Lint Checks Must Pass

- All shell files must pass `shellcheck` with no suppressions except those documented inline with an explanation.
- All Go code must pass `go vet` with no errors.

---

## Key Reference Files

- `terminology.md` — canonical definitions of all domain terms. Read before naming anything.
- `architecture.md` — current intended architecture: the binary, shell files, data flow.
- `CLAUDE.md` — AI working guide (how to work in this repo).
