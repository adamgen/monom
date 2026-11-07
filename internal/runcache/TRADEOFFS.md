# runcache tradeoffs

## A built-in middleware, not a user plugin chain (yet)

**Chosen:** caching is built into `mnmd cache run`, configured by `monom-cache.json`. Its before phase can short-circuit and its after phase sees the exit status.

**Rejected:** shipping the general `pre-run` / `post-run` hook chain (BACKLOG.md) first and building caching on it as a user plugin.

**Why:** a cache's before step must be able to *replace* the command, and its after step must capture output and outputs. A user hook protocol that can do that (how a hook says "skip and replay this", what it receives, ordering across project/category/command layers, what a failing hook does) is a product contract, and the backlog already flags the runner semantics as unsettled. Building the cache first on a fixed internal contract proves the semantics without freezing a public one.

**What it costs:** users can't add their own before/after actions yet; the middleware list has one element, so ordered composition is designed (before runs outer→inner, after inner→outer) but not exercised.

## Replay order instead of interleaving

**Chosen:** a hit writes stored stdout, then stored stderr.

**Rejected:** recording one interleaved stream with timestamps or a PTY.

**Why:** separate streams keep `monom build > log` correct; a PTY would change how commands detect a terminal.

**What it costs:** on a hit, interleaved stdout/stderr lines come back grouped. A cached command's output is piped, not a terminal, so color and progress detection may differ; interactive commands must not be cached. Undeclared dependencies (a tool's version, files outside `inputs`, env vars not in `env`) are not in the key and can produce stale hits.
