# middleware tradeoffs

## Executable chain, not hook events or an in-process interface

**Chosen:** a run middleware is any executable that receives the rest of the chain as argv (`mw <next...>`), and `monom()` runs `( exec mw1 mw2 … cmd )`. Passing through, wrapping and short-circuiting are just `exec "$@"`, running `"$@"`, and not running it.

**Rejected:** (a) `pre-run` / `post-run` event hooks (BACKLOG); (b) shell functions sourced into `monom()`; (c) a Go interface inside `mnmd`.

**Why:** (a) events can't replace the command, so a cache can't be built on them, and a `post-run` forces a resident parent on every command anyway. (b) would be bash/zsh only and run inside the user's shell. (c) closes the door to authors' own middleware, against the constitution's "Pluggability via Hooks". The chain matches how commands already run: a subshell that `exec`s the resolved path. A pass-through middleware keeps that exactly; only a middleware that needs "after" pays for staying resident.

**What it costs:** a wrapping middleware sits between the terminal and the command. It must relay signals itself (`mnmd cache` uses `internal/proc`, which relays SIGINT/SIGTERM/SIGHUP/SIGQUIT; on a terminal Ctrl-C the child can see SIGINT twice, once from the terminal's process group and once relayed). A middleware that captures output turns the command's stdout/stderr into pipes, so terminal detection, colors and stdout/stderr interleaving change. The chain runs in the command's subshell, so middleware can set env for the command but never for the user's shell. Each command run of a hook-script project costs one `mnmd middleware` process and one hook call.

## Chain from a hook, not a declarative key or folder files

**Chosen:** the optional `middleware` hook prints the chain for one command path.

**Rejected:** a `middleware = …` declarative setting; `middleware` files cascading project → category → command.

**Why:** a hook needs no new parsing or precedence rules and can choose per command with any logic; the other two can be added on top later (BACKLOG).

**What it costs:** projects with a declarative or no config can't use middleware without switching to a hook script.
