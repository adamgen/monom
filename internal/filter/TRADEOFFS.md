# Tradeoffs — `mnmd filter`

Decisions here were contested. The code shows *what* filter does; this file records *why* the losing option lost.

---

## filter never exits non-zero

**Chosen:** any internal failure produces empty output and exit 0. `runFilter` even wraps itself in a `recover()`.

**Rejected:** returning a real exit code and a stderr message on bad input, the way every other subcommand does.

**Why.** filter runs during a Tab press, inside a command substitution the shell is expanding while the user is mid-word. Stderr at that moment paints text into a half-typed prompt line, and some shells treat a non-zero completion function as a reason to fall back to their own default completion — which would silently produce *wrong* candidates rather than none. "No completions" is a legible outcome to a user; a corrupted prompt is not.

**What it costs.** Genuine bugs in filter are invisible at the call site. That cost is paid back by `mnmd check`, which exists precisely to be the loud diagnostic that filter refuses to be. Any new failure mode added to filter must also get a check for it — otherwise the failure has nowhere to surface.

---

## Paths containing spaces are dropped silently

**Chosen:** a stdin line with a space in any segment is skipped and excluded from completions.

**Rejected:** quoting or escaping such paths so they survive into `COMPREPLY`.

**Why.** monom's completion protocol is whitespace-delimited end to end — the shell hands filter the typed tokens as separate args, and hands the results back as a word-split array. Supporting spaces in command names would mean inventing an escaping convention that the CLI author's `complete` hook, both shell bindings, and `mnmd pack` all have to implement identically. The feature bought is thin (command names with spaces are bad CLI design regardless) and the surface it adds is wide.

**What it costs.** An author whose `complete` emits such a path gets a command that is undiscoverable with no explanation. `mnmd check` reports exactly this case — it is the primary reason `check` exists.

---

## The completion path is a shell pipe, not one `mnmd` call

**Chosen:**

```bash
COMPREPLY=($( _monom_cfg complete | mnmd filter "${COMP_WORDS[@]:1}" ))
```

**Rejected:** `COMPREPLY=($(mnmd complete "$prefix"))`, with `mnmd` spawning the user config itself via `exec.Command`.

**Why.** Both designs spawn the user config — that subprocess is unavoidable either way, so "minimize subprocess roundtrips" does not discriminate between them. What differs is who owns the plumbing. The shell pipes for free; Go would need `exec.Command`, a pipe, a goroutine, and its own error handling to reproduce it, and would then own the user config's failure modes (hang, partial output, non-zero exit) inside a code path constitutionally forbidden from reporting errors. Keeping the spawn in the shell keeps filter a pure function of `(stdin, args)`, which is why it is testable by piping a fixture into it.

**What it costs.** One more moving part visible in the shell bindings, and the pipeline's exit status is the *last* command's — so a failing `complete` hook is indistinguishable from one that legitimately returned nothing. On the Tab path that collapse is acceptable, because both outcomes render as "no completions" anyway.
