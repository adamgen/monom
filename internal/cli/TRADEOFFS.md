# Tradeoffs — exit codes and `CodedError`

Decisions here were contested. The code shows *what* the registry does; this file records *why* the losing option lost.

---

## The error carries the exit code, not the call site

**Chosen:** every subcommand returns an `error`; `main.go` has one dispatch tail that reads `ExitCode()` off it via `errors.As`.

**Rejected:** each `case` in `main.go`'s switch inspecting the concrete error type and picking an exit code itself.

**Why.** With per-call-site branching, the knowledge "a command group means 3" lives in `main.go` while the condition that produces it lives in `internal/pack`. Adding an outcome to a subcommand then requires editing a file that has no other reason to change, and forgetting to do so is silent — the outcome degrades to a generic exit 1 and the shell's branching logic quietly stops matching. Putting the code on the error makes the two inseparable: you cannot construct the error without stating its exit code.

**What it costs.** `errors.As` and an interface for something a switch statement could express in ten lines. The complexity is real but bounded and paid once, in one function.

---

## Exit codes live in one registry, with no literals elsewhere

**Chosen:** `ExitCodes` in this package is the sole definition; nothing outside it writes `3` as an exit code, and `architecture.md` references this file instead of restating the numbers.

**Rejected:** named constants declared next to the code that uses them, e.g. `pack.ExitGroup = 3`.

**Why.** Exit codes are a shared namespace with exactly one enforced property: no two outcomes may collide. That property is only checkable by looking at all of them at once. Distributing the constants makes a collision invisible until a shell binding branches on 3 and gets an outcome it never meant to handle. One struct is the cheapest way to make the whole namespace readable in a single glance.

The same reasoning applies to documentation: duplicating the values into `architecture.md` creates a second copy that drifts. The doc names the file; the file names the numbers.

**What it costs.** A package-level import from every subcommand that returns an error, and a small amount of indirection when reading a single subcommand in isolation.

---

## Exit 3 is the group signal for every resolver, not a pack-private code

**Chosen:** exit 3 means "these tokens name a category" wherever it appears on the run path — `pack` emits it for a directory, `map resolve` for a map category, and `monom()` treats a run hook's exit 3 identically to pack's, rendering the same child listing.

**Rejected:** keeping 3 exclusive to pack and giving hook-signalled groups a different code (or reporting them as hook failures).

**Why.** A command-map category has no directory on disk, so pack cannot see it — the only process that knows `monom db` names a group is the run hook. The user-facing outcome is *identical* to pack's directory case, and the shell already has the rendering machinery; a second code would duplicate the branch and the docs for one behavior, and "hook failure" would turn a correct answer into an error message.

**What it costs.** Run hooks lose exit 3 as an ordinary failure code. An author's hook that happens to exit 3 for a real error gets a group listing instead of its error message — a silent misrendering monom cannot detect. The contract is documented in `architecture.md`'s run-hook table, but it is a genuine carve-out from the author's exit-code namespace.

---

## `filter` is exempt from the whole mechanism

`runFilter` calls `os.Exit(0)` directly and never reaches the dispatch tail. This is not an oversight — filter is constitutionally forbidden from exiting non-zero, so participating in an exit-code mechanism would be meaningless at best and dangerous at worst. See `internal/filter/TRADEOFFS.md`.
