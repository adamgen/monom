## Context

The completion pipeline has three layers, all of which currently suppress errors silently:

1. **Shell binding** (`monom_completion` / `_monom`) — redirects all stderr to `/dev/null`, always exits 0.
2. `**monom_cfg complete`** — runs the user's config file; its output is piped to `monomd filter`.
3. `**monomd filter**` — reads stdin, filters by words, prints results. Any failure returns empty output and exits 0.

When completion produces nothing, the failure could be at any layer: `setup_monom` failing to find the root, the user config not printing anything, `monomd filter` receiving bad input, or the shell binding not passing the right words. Today there is no way to distinguish these cases without removing the error suppression — which corrupts the terminal mid-typing.

The solution is opt-in debug logging infrastructure: when `MONOM_DEBUG_LOG` is set to a file path, any layer can append a timestamped trace line using the shared helpers. The variable is unset by default, so there is zero overhead on the normal path. What each layer logs is wiring detail added incrementally — this change only delivers the infrastructure.

## Goals / Non-Goals

**Goals:**

- Provide `_monom_log` (shell) and `internal/debuglog.Log` (Go) as shared infrastructure any layer can call
- Zero overhead (no file open, no syscall) when `MONOM_DEBUG_LOG` is unset
- Append-only log so multiple completions accumulate without overwriting
- No changes to stdout, stderr, or exit codes on any path

**Non-Goals:**

- Wiring log calls into specific layers (done incrementally, outside this spec)
- Structured/machine-readable log format (plain timestamped lines are enough for debugging)
- Log rotation or size limits (it's a debug tool; the user manages the file)

## Decisions

### Decision 1: `MONOM_DEBUG_LOG` is the activation env var

*A single env var is the simplest opt-in. The user sets it in their shell before sourcing monom or in a debug session:*

```sh
export MONOM_DEBUG_LOG=/tmp/monom-debug.log
source ~/monom/src/monom
monom <Tab>
cat /tmp/monom-debug.log
```

**Alternative considered:** A flag passed to `monomd filter` — rejected because the shell binding would need to forward it, adding complexity to the hot path.

### Decision 2: Shell helper `_monom_log` writes timestamped lines

Defined in `src/monom` (available to both bash and zsh bindings):

```sh
_monom_log() {
  [ -n "$MONOM_DEBUG_LOG" ] || return 0
  printf '[%s] %s\n' "$(date +%T)" "$*" >> "$MONOM_DEBUG_LOG"
}
```

The check is first: if `MONOM_DEBUG_LOG` is unset, the function returns immediately without spawning `date` or opening any file.

**Alternative considered:** Inline `[ -n "$MONOM_DEBUG_LOG" ] && printf ...` at each call site — rejected because it duplicates the timestamp format and is harder to change.

### Decision 3: Go helper package `internal/debuglog`

A single exported function:

```go
// Log appends a timestamped line to $MONOM_DEBUG_LOG if set. No-op otherwise.
func Log(format string, args ...any)
```

Any `monomd` subcommand can import and call it. Using a package rather than inlining keeps the log format consistent and lets the file open/append/close be encapsulated.

**Alternative considered:** Writing directly to a log file path in `main.go` — rejected because it scatters the file-open logic and makes it harder to ensure the no-op path is truly zero-cost.

## Risks / Trade-offs

- [Risk] `_monom_log` spawns `date` on every log call, adding latency when debugging → Accepted: debug mode is explicitly opt-in; latency in debug mode is acceptable.
- [Risk] Log file grows unboundedly → Accepted: this is a debug tool; the user truncates it between sessions. Document this in a comment.
- [Risk] Go layer opens the file once per log call (open/write/close) → Mitigation: acceptable for a debug tool. A persistent file handle would complicate the no-op path.

