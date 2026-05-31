## Context

The `monomd args` binary subcommand provides single-flag parsing. CLI authors call it as `PROP=$(monomd args prop -- "$@")`. For scripts with many flags, this pattern repeats on every line. Additionally, there's no standard "bail out with an error" function — authors write ad-hoc `echo >&2; exit 1` every time.

These two shell helpers are thin wrappers that provide ergonomics on top of the binary. They follow the constitution's "Go owns logic, shell owns surface" principle: the parsing logic stays in `monomd args`; the shell function only handles variable assignment (which requires parent-shell scope) and the multi-flag declaration syntax.

## Goals / Non-Goals

**Goals:**
- Provide `monom_args` — a single-call declaration of multiple flags that sets variables in the caller's scope
- Provide `monom_error` — a terse way to print to stderr and exit
- Both functions work identically in bash and zsh
- Both functions are available to CLI author command scripts after sourcing monom

**Non-Goals:**
- Adding parsing logic to shell — `monom_args` delegates entirely to `monomd args`
- Help/usage generation — out of scope
- Argument validation beyond what `monomd args` already provides
- Making these functions available outside of monom-sourced environments

## Decisions

### Decision: `monom_args` declaration syntax

**Choice:**
```bash
monom_args [modifiers...] <name> [modifiers...] <name> ... -- "$@"
```

The `--` separates the flag declarations from the raw args. Before `--`, tokens are grouped: each bare word is a variable name (and flag name), and `--`-prefixed tokens before it are modifiers for that flag. After `--` is the raw argument list passed through to `monomd args`.

**Example:**
```bash
monom_args --short=e env --short p port --boolean verbose -- "$@"
# Sets: $env, $port, $verbose
```

**Rationale:** Mirrors the `monomd args` modifier syntax exactly. CLI authors learn one pattern. The shell function just splits the declaration into individual `monomd args` calls.

### Decision: `monom_args` sets variables via `printf -v`

**Choice:** Use `printf -v "$name" '%s' "$(monomd args ...)"` to set variables in the caller's scope.

**Rationale:** `printf -v` sets a variable by name in the calling scope without `eval`. Works in both bash and zsh. Avoids the security and linting concerns of `eval`.

### Decision: `monom_args` boolean variables get "true" or empty string

**Choice:** For `--boolean` flags, set the variable to `"true"` if the flag is present (exit 0 from `monomd args`), or `""` (empty) if absent.

**Rationale:** Shell idiom: `if [[ "$verbose" ]]; then ...` works naturally. Empty = falsy, non-empty = truthy. Using `"true"` (not `"1"`) is more readable in debug output.

### Decision: `monom_error` interface

**Choice:**
```bash
monom_error <message> [exit_code]
```

Prints `<message>` to stderr and exits with `exit_code` (default: 1).

**Example:**
```bash
monom_error "missing required flag --name"
monom_error "invalid port number" 2
```

**Rationale:** Minimal interface. Message is required, exit code is optional with a sensible default. No formatting opinions (no prefixes like "error:") — the CLI author controls the message.

### Decision: Where the functions live

**Choice:** Define both functions in `src/monom` alongside the existing `monom()` function. They're sourced into the user's shell session and available to any command script that runs within a monom context.

**Rationale:** Command scripts are executed by the shell's `monom()` function (or via `exec`), so they inherit the sourced environment. No separate sourcing step needed for CLI authors.

## Risks / Trade-offs

- **N subprocess spawns for N flags in `monom_args`** → Acceptable. monom command scripts are human-invoked, not tight loops. 5 flags × ~5ms per spawn = imperceptible.
- **`printf -v` zsh compatibility** → In zsh, `printf -v` works but only in function scope with `emulate -L bash` or by using zsh's native `typeset`. Need to verify behavior or use a zsh-compatible assignment. Mitigation: test in both shells during implementation.
- **`monom_error` calls `exit`** → This terminates the script. If called inside a subshell or pipe, it only exits the subshell. This is standard shell behavior and acceptable — document it.
