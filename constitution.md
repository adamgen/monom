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

## Principle: The User Config Interface Requires a Constitution Amendment to Change

The `monom` config file at the root of any user project currently exposes:

```
$MONOM_USER_CONFIG complete   # prints all discoverable command paths, one per line
$MONOM_USER_CONFIG run        # reads args, prints the resolved file path to execute
```

monom does not care how these are implemented. This interface may evolve, but any change to it requires a corresponding update to this document. Changing the interface without amending the constitution is a process violation.

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
