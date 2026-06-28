## Context

Users who install monom have a working binary, but the `monom` shell function and tab completion only activate once `src/monom` is sourced in the user's shell rc file. Discovering that this step is required — and knowing exactly which line to add to which file — is a friction point for first-time users. The goal is to make that step automatic and discoverable.

Two related problems are solved together:

1. **mnmd install**: a subcommand that writes the source line into the correct rc file.
2. **install nudge**: a hint printed when someone calls `mnmd` directly (rather than via the `monom()` shell function), signalling that the shell integration is not yet active.

## Goals / Non-Goals

**Goals:**
- `mnmd install` appends a `source` line to the user's shell rc/profile file if not already present, then prints a message telling the user to restart their shell or re-source.
- The nudge surfaces to stderr only when `mnmd` is invoked directly without the shell integration active, and only when it would be actionable (i.e. when `mnmd install` would help).
- Idempotent: running `mnmd install` multiple times does not create duplicate lines.
- Supports zsh (primary target) and bash. Falls back gracefully when neither can be detected.

**Non-Goals:**
- Fish, dash, or other shells.
- Modifying system-level files (`/etc/profile`, `/etc/zshrc`).
- Uninstalling or removing a previously added source line.
- Managing `MONOM_LIB_ROOT` — the user is responsible for keeping the install directory stable.

## Decisions

### Decision: Implement install in Go (`internal/install/`)

The install logic (RC file detection, idempotency check, file append) is pure data transformation with no technical shell-only constraint. Per the architecture principle "Go owns logic", this belongs in `internal/install/` as a Go package, not in a shell file.

*Alternative considered*: A shell script helper. Rejected because it would add a subprocess roundtrip, complicate testing, and violate the "Go owns logic" principle.

### Decision: Shell detection by inspecting `$SHELL` env var

At the binary level, we cannot inspect `$ZSH_VERSION` or `$BASH_VERSION` (those are shell variables in the parent process, not environment variables passed to subprocesses). Instead, `mnmd install` reads `$SHELL` from the environment, which is exported by the login shell and reliably set on macOS.

Detection logic:
- `$SHELL` ends in `/zsh` → target `~/.zshrc`
- `$SHELL` ends in `/bash` → prefer `~/.bash_profile` on macOS (where `.bashrc` is not sourced for login shells), fall back to `~/.bashrc`
- Anything else → exit non-zero with an error message listing the detected shell

### Decision: Source line uses a resolved absolute path

The installed line points directly at `src/monom` by absolute path:

```sh
source "/path/to/monom/src/monom"
```

The absolute path is computed inside `internal/install` by reading the location of the running `mnmd` binary (`os.Executable()`) and resolving its `../src/monom` sibling.

*Alternative considered*: Writing `source "$MONOM_LIB_ROOT/monom"` and relying on the variable. Rejected because `MONOM_LIB_ROOT` is only set *by* `src/monom` at source time — it does not exist in the user's environment until after `src/monom` has been sourced. The rc line *is* that first source, so the variable would be empty when the line runs. An absolute path is the only thing available at bootstrap.

### Decision: Idempotency check is string-contains on the source path

Before appending, `internal/install` reads the target rc file and checks whether any line contains the resolved `src/monom` path. If found, it exits 0 with "already installed" message and does nothing.

### Decision: Nudge is controlled by `MONOM_ACTIVE` env var

`src/monom` exports `MONOM_ACTIVE=1` when sourced. When `mnmd` is invoked without this variable set, it prints a one-line nudge to stderr: `hint: run 'mnmd install' to activate shell integration`. The nudge fires only on subcommands that are relevant to end-users at the CLI (i.e. all subcommands except `install` itself, to avoid recursive advice).

*Alternative considered*: Detecting whether the caller is an interactive shell. Rejected as unreliable and over-engineered; the env var is simpler and explicit.

## Risks / Trade-offs

- [Risk: `os.Executable()` returns a symlink path] → Mitigation: call `filepath.EvalSymlinks` to resolve the real path before computing `../src/`.
- [Risk: rc file ends without a trailing newline] → Mitigation: always prepend a newline before the appended line.
- [Risk: nudge fires during automated script use of `mnmd`] → Mitigation: nudge goes to stderr, not stdout, so it doesn't pollute piped output; scripts that call `mnmd` directly and do not want the nudge can set `MONOM_ACTIVE=1` in their environment.
- [Risk: `MONOM_ACTIVE` export adds surface area to `src/monom`] → This is a minimal, well-named variable; acceptable cost.

## Open Questions

- Should `mnmd install` also print the rc file path it modified, for transparency? (Likely yes — include in spec.)
- Should `mnmd install` support a `--dry-run` flag to preview the line without writing? (Out of scope for v1; add if needed.)
