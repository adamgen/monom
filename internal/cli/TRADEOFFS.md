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

## `filter` is exempt from the whole mechanism

`runFilter` calls `os.Exit(0)` directly and never reaches the dispatch tail. This is not an oversight — filter is constitutionally forbidden from exiting non-zero, so participating in an exit-code mechanism would be meaningless at best and dangerous at worst. See `internal/filter/TRADEOFFS.md`.
