# runcache tradeoffs

## Replay order instead of interleaving

**Chosen:** a hit writes the stored stdout, then the stored stderr.

**Rejected:** recording one interleaved stream with timestamps or a PTY.

**Why:** separate streams keep `monom build > log` correct; a PTY would change how commands detect a terminal.

**What it costs:** on a hit, interleaved stdout/stderr lines come back grouped. On a miss the command's output is piped, not a terminal, so color and progress detection may differ; interactive commands must not be cached. Undeclared dependencies (a tool's version, files outside `inputs`, env vars not in `env`) are not in the key and can produce stale hits.

## Entries keyed by the whole inner chain

**Chosen:** the key includes next's argv: the middleware inside the cache and the command.

**Rejected:** keying by the command alone.

**Why:** a middleware inside the cache can change what the command prints or writes; changing the chain must not replay results from another chain.

**What it costs:** absolute paths are in the key, so moving the project or reordering the chain starts a fresh cache.
