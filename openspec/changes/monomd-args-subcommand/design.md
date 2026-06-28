## Context

`mnmd args` is already documented in `architecture.md` as a planned subcommand with TBD output format. CLI authors writing command scripts today must hand-roll flag parsing in bash — typically with `getopts` (positional only) or fragile `case` loops. `mnmd args` replaces this with a single, consistent call to the binary.

The subcommand follows the project's "CLI arguments by default" principle: the raw args to parse are a bounded, known parameter set at call time — not a stream — so they pass as CLI args, not stdin.

## Goals / Non-Goals

**Goals:**
- Parse `--flag=value` (long equals form) and `--flag value` (long space form) from a raw argument list
- Parse `-f value` and `-f=value` short forms when `--short` modifier is used
- Print the resolved value for a requested flag name to stdout
- Support `--boolean` modifier for presence-only checks, with automatic `--no-<flag>` negation
- Support `--short` modifier to register a single-character alias
- Keep the interface trivially composable — one flag lookup per invocation

**Non-Goals:**
- Defaulting / fallback values — caller handles that with shell `${PROP:-default}`
- Parsing positional arguments — separate concern, separate future subcommand if needed
- Multiple flag lookups in one call — one call per flag keeps the contract simple
- A shell helper function for multi-flag parsing — out of scope for this change (see example below for future reference)

## Decisions

### Decision: `--` separator between mnmd's arguments and raw args

**Choice:** `mnmd args [modifiers...] <flag-name> -- <raw-args...>`

Everything before `--` belongs to mnmd (modifiers + the flag name). Everything after `--` is the raw argument list to search through. The `--` is required.

**Rationale:** The `--` provides an explicit, unambiguous boundary. Without it, the parser must infer where modifiers/flag-name end and raw args begin based on token shape — which becomes fragile as modifiers grow (especially `--short p` where the value is a bare word). The `--` convention is universally understood by POSIX-aware tools and shell authors.

**Alternative considered:** Implicit boundary based on first bare word = flag name, everything after = raw args — rejected because it makes the grammar stateful (parser must know which modifiers take values) and creates potential for subtle misparses.

### Decision: `--boolean` modifier for presence checks with `--no-` negation

**Choice:** `mnmd args --boolean verbose -- "$@"` exits 0 if `--verbose` is present in raw args, exits 1 if absent. No stdout in either case. The parser also recognizes `--no-verbose` as explicit negation (exit 1). Last-wins applies between `--verbose` and `--no-verbose`.

**Rationale:** Boolean flags have no value — their presence IS the meaning. The `--no-` prefix is a widely understood convention (GNU, Go flag, Argbash) for explicitly turning off a boolean. Last-wins semantics let callers override: `--verbose --no-verbose` results in "off."

### Decision: `--short` modifier for single-character aliases with bundling support

**Choice:** `mnmd args --short p prop -- "$@"` (or `--short=p`) searches for both `--prop` and `-p` in the raw args. The `--short` value must be exactly one character. Short-form matching supports `-p value` (space form), `-p=value` (equals form), and bundled forms (`-xp value`, `-xp=value`). Last-wins applies across both long and short forms.

**Bundling rules:** When multiple short flags are combined (e.g., `-abc`), only the last character in the bundle can take a value. Characters before the last are boolean-only. A value flag found in non-last position within a bundle has no value available and is treated as absent.

**Rationale:** While long flags are preferred for self-documentation in command scripts, some CLI authors want to offer short aliases to their CLI users. The modifier is opt-in — only flags that explicitly declare `--short` get short-form matching. Bundling is a standard POSIX convention that users expect to work when short flags are available.

### Decision: All modifiers support equals and space forms

**Choice:** Every modifier accepts both `--mod=value` and `--mod value` for its arguments.

**Rationale:** Consistency. The same rules that apply to flag parsing in raw args apply to mnmd's own modifiers. CLI authors can use whichever form is more readable at the call site.

### Decision: Exit non-zero when flag is absent

**Choice:** Exit code 1 and empty stdout when a flag is not found.

**Rationale:** The binary signals absence through the exit code. It does not handle "required" semantics — that is the caller's responsibility. The exit code gives the caller a clean hook to enforce required flags, provide defaults, or silently accept absence as needed.

**Required flags are a caller-side concern:**

```bash
# Optional — empty string if absent, script continues
port=$(mnmd args port -- "$@")

# Optional with default
port=$(mnmd args port -- "$@")
port="${port:-8080}"

# Required — caller enforces with || and a message
name=$(mnmd args name -- "$@") || { echo "error: --name is required" >&2; exit 1; }

# Required under set -e — script aborts on absence automatically
set -e
name=$(mnmd args name -- "$@")
```

**Alternative considered:** Exit 0 with empty stdout always — rejected because it makes absence indistinguishable from a flag present with an empty value.

### Decision: Last-wins for duplicate flags

**Choice:** If `--prop=a --prop=b` appear, return the last occurrence. When `--short` is active, last-wins applies across both long and short forms.

**Rationale:** Consistent with most CLI parsers (including Go's `flag` package). Predictable and easy to reason about in scripts. Allows overriding defaults by appending flags.

## Risks / Trade-offs

- **`--flag value` vs `--flag=value` ambiguity** → Mitigated by parsing equals-form first; space-form only consumes the next token if it does not start with `--` or `-`.
- **Empty string values** (`--prop=`) are valid and print empty string to stdout with exit 0 → This makes absence (exit 1, no output) cleanly distinguishable from "present but empty" (exit 0, empty output). Acceptable trade-off.
- **Unknown modifiers** → If mnmd encounters a `--`-prefixed token it doesn't recognize before `--`, it should error (unknown modifier) rather than silently passing it through.

## Future Reference: Shell Helper for Multi-Flag Parsing

Out of scope for this change, but the binary interface enables a thin shell function that parses multiple flags in one declaration block:

```bash
monom_args() {
  local _specs=() _raw=()
  while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--" ]]; then shift; _raw=("$@"); break; fi
    _specs+=("$1"); shift
  done
  local _i=0
  while (( _i < ${#_specs[@]} )); do
    local _modifiers="" _name=""
    while [[ "${_specs[$_i]}" == --* ]]; do
      _modifiers+="${_specs[$_i]} "; ((_i++))
    done
    _name="${_specs[$_i]}"; ((_i++))
    printf -v "$_name" '%s' "$(mnmd args $_modifiers$_name -- "${_raw[@]}")"
  done
}

# Usage:
monom_args --short=e env --short p port --boolean verbose -- "$@"
echo "$env" "$port" "$verbose"
```

This keeps the binary simple (one flag per call) while giving CLI authors a concise multi-flag declaration when needed. The function delegates entirely to `mnmd args` — no parsing logic in shell.
