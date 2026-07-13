## Context

`cmd/mnmd/main.go` dispatches to `runX` functions that each hardcode their exit code (`os.Exit(1)`, `os.Exit(3)`) and format stderr independently. The exit-code semantics (e.g. `pack.GroupError` → 3) are explained in comments rather than expressed in the type system. Adding a new exit code means editing `main.go`'s switch arms and keeping comments in sync — a maintenance burden that grows with every new subcommand.

## Goals / Non-Goals

**Goals:**

- Every error returned by a subcommand carries its own exit code via the `CodedError` interface.
- A single central registry maps exit codes to their meaning and usage, serving as the sole source of truth.
- `main.go`'s dispatch tail becomes a uniform `errors.As` check — no per-subcommand exit-code logic.
- `pack.GroupError` derives its exit code (3) from the registry, not a literal in `main.go`.
- The exit-code comment blocks in `main.go` become unnecessary and are deleted.
- `constitution.md` gains a new principle; `architecture.md` references the registry.

**Non-Goals:**

- Changing any externally observable exit code, stdout, or stderr format.
- Touching `internal/filter` — it always exits 0 and is handled separately.
- Refactoring stdout-printing paths (success output stays in `runX`).
- Introducing error wrapping chains or sentinel errors beyond what is needed for the registry.

## Decisions

### 1. Package location: `internal/cli`

The `CodedError` interface, embeddable base, and exit-code registry live in `internal/cli`. This is a neutral package importable by all subcommand packages (`internal/pack`, etc.) and by `cmd/mnmd/main.go` without creating import cycles.

*Alternative considered:* `internal/exit` — more specific, but `cli` leaves room for future CLI-wide utilities (e.g. stderr formatting helpers) without creating yet another package.

### 2. Interface + embeddable base

```go
type CodedError interface {
    error
    ExitCode() int
}

type Base struct{ Code int }
func (b Base) ExitCode() int { return b.Code }
```

Concrete error types (like `pack.GroupError`) embed `Base` to get `ExitCode()` for free. A generic `Err` wrapper covers the common "code 1 + message" case.

### 3. Central exit-code registry

A package-level table in `internal/cli` maps each exit code to its meaning:

```go
var ExitCodes = struct {
    Success    int // 0 — leaf resolved / normal output
    Error      int // 1 — generic real error
    GroupError int // 3 — pack command-group signal (payload-free)
}{
    Success:    0,
    Error:      1,
    GroupError: 3,
}
```

All `CodedError` constructors and `main.go` dispatch read from this struct — no integer literals elsewhere. `architecture.md`'s inline exit-code table is replaced with a reference to this file as the source of truth, eliminating the duplication.

### 4. Generic error wrapper

A constructor wraps any `error` as a `CodedError` with a code from the registry:

```go
func WrapError(err error) CodedError { ... } // always uses ExitCodes.Error (1)
```

This covers root, check, install, and pack's non-group error paths — they return `fmt.Errorf(...)` today, and `main.go` wraps them uniformly.

### 5. Uniform dispatch in main.go

Each `runX` function returns `error` (or `nil` for success). `main` has one tail:

```go
err := runXxx()
if err == nil { return }
var ce cli.CodedError
if errors.As(err, &ce) {
    if ce.ExitCode() != cli.ExitCodes.GroupError {
        fmt.Fprintln(os.Stderr, "mnmd <sub>:", ce)
    }
    os.Exit(ce.ExitCode())
}
fmt.Fprintln(os.Stderr, "mnmd <sub>:", err)
os.Exit(cli.ExitCodes.Error)
```

The `filter` case keeps its existing `os.Exit(0)` — it is exempt from the error dispatch.

### 6. GroupError embeds Base

`pack.GroupError` gains an embedded `cli.Base` initialized from the registry:

```go
type GroupError struct {
    cli.Base
    Path string
}
```

The constructor sets `Base.Code = cli.ExitCodes.GroupError`. `main.go` no longer needs `errors.As(err, &ge)` — it just reads `ce.ExitCode()`.

## Risks / Trade-offs

- **Coupling `internal/pack` → `internal/cli`**: Acceptable — `cli` is a leaf dependency with no transitive imports. All subcommand packages already share `internal/root`.
- **Registry struct vs. const block**: A struct groups code + name; a const block is simpler. Chose struct for greppability (one place with names and values). If it feels heavy, constants with a doc comment are an acceptable fallback.
- **GroupError's stderr suppression**: The dispatch tail must know that `ExitCodes.GroupError` produces no stderr. This is a single `if` keyed on the registry constant — the "special case" is small and documented.
