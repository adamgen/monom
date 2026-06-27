# monom Architecture

This document describes the current intended architecture of monom. Unlike the constitution, it is descriptive and will evolve as the project develops. It should stay consistent with the principles in `constitution.md`.

---

## Entry Points

monom has exactly three entry points — the complete public surface through which anything interacts with it. Everything else (`mnmd` subcommands, hooks, internal env vars) is machinery reached *through* these three.

1. **`source monom`** — bootstrap. The user sources `src/monom` from their rc file. This is the one-time introduction of monom into a shell session: it resolves `_MONOM_LIB_ROOT`, defines the `mnmd()`, `monom()`, `_setup_monom()`, and `_monom_cfg()` functions, and sources the shell-specific completion binding (`src/monom.bash` or `src/monom.zsh`). After this, the shell knows the `monom` command and how to complete it. No `mnmd` subcommand runs at source time. `mnmd` itself is never added to the user's namespace — `mnmd()` is an internal shell function, not a user-facing command.

2. **`monom <Tab>`** — completion. The user presses Tab while typing a `monom` command. The registered completion function runs discovery (`_monom_cfg complete`) and filters it (`mnmd filter`) to populate `COMPREPLY`. See [Completion (Tab press)](#completion-tab-press) for the full flow. This path must be fast and never error mid-typing.

3. **`monom [command...]`** — execution. The user runs a resolved command. The `monom()` function optionally transforms args via the `run` hook, resolves them to an absolute executable path (`mnmd pack`), and `exec`s it. See [Command execution](#command-execution) for the full flow.

Entry points 2 and 3 both depend on entry point 1 having run first in the session.

---

## Design Principles

Operational guidelines for designing tools within the monom project (currently `mnmd`, and any future tools we build). Unlike the constitution's principles, these are project-design heuristics — they SHOULD be followed, but deviations are allowed when justified. These principles do not apply to the user config interface; CLI authors choose their own conventions there.

### Principle: CLI Arguments by Default, Stdin When the Input Is a Stream

Subcommand inputs SHOULD be CLI arguments. Stdin SHOULD only be used when the input is genuinely a stream — many lines, unbounded data, content piped from another tool, or cases where streaming improves correctness or composability.

**The test:** Before designing a subcommand to read stdin, ask — "is this a parameter or a stream?" Parameters are bounded, named, and known at call time; they belong in args. Streams are unbounded, anonymous, and benefit from pipe composition; they belong in stdin.

Examples in this codebase:
- `mnmd filter` reads commands from stdin — the command list is unbounded and naturally produced by piping `_monom_cfg complete`. Stream.
- `mnmd pack` takes args — the user's command tokens are a small, known parameter set produced by the shell at call time, not a stream. Parameters.
- `mnmd root` and `mnmd check` take no input — neither parameters nor stream.

---

## The Binary: mnmd

There is one compiled Go binary: `mnmd`.

All subcommands read environment variables at startup — see [Environment Variables](#environment-variables) for the full reference. In shell scripts, the `_monom_cfg` function wrapper is used for readability at call sites:

```bash
_monom_cfg() { "$_MONOM_USER_CONFIG" "$@"; }
```

---

### `mnmd filter [word...]`

Reads command paths from stdin (one per line, slash-delimited) and accepts zero or more space-separated word arguments. Filters stdin to paths matching the given words, and prints the next-level completions to stdout (one per line).

**The two formats:** stdin uses `/` to separate category levels (e.g. `infra/cloud/deploy`) because that is how the CLI author's `complete` output encodes the command tree. The word arguments use spaces because they come directly from what the user typed — the shell passes them as separate args with no transformation. `mnmd filter` bridges the two: it joins the word arguments with `/` internally to produce a prefix, then matches that prefix against the slash-delimited stdin paths.

Called by the shell completion binding as part of a pipe:

```bash
_monom_cfg() { "$_MONOM_USER_CONFIG" "$@"; }
COMPREPLY=($(  _monom_cfg complete | mnmd filter "${COMP_WORDS[@]:1}"  ))
```

The shell passes `${COMP_WORDS[@]:1}` (the raw typed tokens after `monom`) directly — no transformation in shell. Stdin lines containing spaces in any path segment are silently ignored — emitting a hard error during a Tab press would produce noise in the terminal mid-typing. Use `mnmd check` to surface these issues explicitly during development.

**`mnmd filter` never exits with a non-zero status code.** This is a hard constraint, not a guideline. Tab completion is interactive — a non-zero exit or stderr noise mid-typing degrades the user experience and can break shells that interpret completion failures unpredictably. Any internal error produces empty output and exit 0. All diagnostics belong in `mnmd check`.

A trailing empty word in the argument list — which bash appends to `$COMP_WORDS` when the user has typed a complete token followed by a space — signals "drill into this level" rather than "match partially."


| User types                   | `$COMP_WORDS` (args to filter) | stdin format                         | filter output         |
| ---------------------------- | ------------------------------ | ------------------------------------ | --------------------- |
| `monom <Tab>`                | *(none)*                       | `category1/sub1\ncommand1`           | `category1\ncommand1` |
| `monom com<Tab>`             | `com`                          | `command1\ncommand2\ncategory1/sub1` | `command1\ncommand2`  |
| `monom category1 <Tab>`      | `category1` `""`               | `category1/sub1\ncategory1/sub2`     | `sub1\nsub2`          |
| `monom category1 sub<Tab>`   | `category1` `sub`              | `category1/sub1\ncategory1/sub2`     | `sub1\nsub2`          |
| `monom category1 sub1 <Tab>` | `category1` `sub1` `""`        | `category1/sub1/leaf`                | `leaf`                |


**Single-level example** — the `file_commands` test project:

```
# stdin: $_monom_cfg complete output
category1/sub_command1
category1/sub_command2
command1
command2
```

```
# user types: monom <Tab>
$ _monom_cfg complete | mnmd filter
category1
command1
command2

# user types: monom com<Tab>
$ _monom_cfg complete | mnmd filter com
command1
command2

# user types: monom categ<Tab>
$ _monom_cfg complete | mnmd filter categ
category1

# user types: monom category1 <Tab>  (trailing space → bash appends "")
$ _monom_cfg complete | mnmd filter category1 ""
sub_command1
sub_command2
```

**Nested example** — a project with two levels of categories:

```
# stdin: $_monom_cfg complete output
infra/cloud/deploy
infra/cloud/teardown
infra/local/start
infra/local/stop
release
```

```
# user types: monom <Tab>
$ _monom_cfg complete | mnmd filter
infra
release

# user types: monom infra <Tab>
$ _monom_cfg complete | mnmd filter infra ""
cloud
local

# user types: monom infra cl<Tab>
$ _monom_cfg complete | mnmd filter infra cl
cloud

# user types: monom infra cloud <Tab>
$ _monom_cfg complete | mnmd filter infra cloud ""
deploy
teardown
```

Stdin lines with spaces in any path segment are silently ignored and excluded from completions.

**Why a pipe instead of a single `mnmd` call:**
Calling the user config is a subprocess spawn either way. Shell pipes natively; Go needs goroutines and io plumbing to do the same. The rejected alternative:

```bash
# ❌ Rejected: COMPREPLY=($(mnmd complete "$prefix"))
```

```go
// ❌ What mnmd would do internally:
cmd := exec.Command(os.Getenv("_MONOM_USER_CONFIG"), "complete")
out, _ := cmd.Output()
matches := filterByPrefix(strings.Split(strings.TrimSpace(string(out)), "\n"), prefix)
```

---

### `mnmd pack <word...>`

Called by the `monom()` shell function when the user executes a command. Takes the user's space-separated command path as CLI args (e.g. `mnmd pack category1 sub_command1`), joins the tokens with `/` internally, resolves the result against the project root, and prints the absolute path to stdout. The shell then `exec`s that path.

Pack is self-sufficient: it discovers the project root itself (same algorithm as `mnmd root` — see below), validates the file exists and is executable, and prints the absolute path. The shell `monom()` function reduces to a single `exec` call.

```bash
$ mnmd pack category1 sub_command1
/path/to/project/category1/sub_command1
```

Pack is the symmetric counterpart of `filter`: both take space-separated tokens as CLI args and bridge to the slash-delimited file tree. Pack's specific job is to replace spaces with slashes and resolve to an absolute, executable path.

**Exit codes.** Pack signals its outcome through the exit code so the shell can branch without parsing strings:

| Exit | Meaning | Output |
| ---- | ------- | ------ |
| `0` | A leaf command resolved | absolute path on stdout |
| `3` | The tokens resolved to a **command group** (a directory, not a runnable file) | none — stdout and stderr both empty |
| `1` | A real error (no args, not found, not executable, no project root) | message on stderr |

Exit code `3` is **reserved exclusively** for the command-group outcome. A directory is a noun in monom's noun→verb file tree — it is not a command, but it is also not a failure, so it gets its own signal. Exit 3 is a *pure signal*: pack writes nothing and, crucially, does **not** enumerate the group's children. Discovery is the `complete` hook's job (see [terminology](terminology.md) and the [user config interface](#the-user-config-interface)); pack stays a pure resolver that returns a path xor a non-leaf signal. The `monom()` function turns exit 3 into a user-facing listing by sourcing the children from the canonical discovery pipeline — `monom_cfg complete | mnmd filter <tokens> ""`, the same path tab-completion uses — so the listing matches `monom <group> <Tab>` and honors any `run`-hook surface tree, rather than re-deriving it from a direct filesystem read. The result is `monom: 'infra' is a command group` / `available: cloud, local`; see [shell files](#shell-files).

**Making a namespace runnable is an author concern, not a monom flag.** monom deliberately does *not* let a group double as a runnable command via some built-in override. Per clig.dev's "don't have a catch-all subcommand", auto-picking a default leaf for a group would be a time bomb: the day the author adds a real leaf with that name, every existing group invocation silently changes meaning. An author who wants `monom infra` to *do* something expresses that explicitly in their own config via the [`run` hook](#hook-run--transform-args-before-path-resolution) (e.g. mapping `infra` → `infra cloud deploy`), keeping the override visible and stable in their project rather than baked into monom's core.

---

### `mnmd root`

Returns the active monom project root.

Algorithm: if `$_MONOM_PROJECT_ROOT` is set and points to a directory containing an executable `monom` file, return it. Otherwise, walk up from `$PWD` looking for a directory containing an executable `monom` file. Print the first match to stdout; exit non-zero if none is found.

This same algorithm is shared internally by all subcommands that need the root (currently `mnmd pack`). Exposed as a standalone subcommand so the shell and CLI authors can query the root explicitly (for sourcing scripts, aliases, debugging).

```
$ mnmd root
/path/to/project
```

---

### `mnmd check`

Validates that the current monom project is healthy. Runs `_monom_cfg complete`, inspects every path in the output, and reports any problems to stdout. Exits non-zero if any problems are found.

Currently checks:

- Every path is slash-delimited with no spaces in any segment. A path with spaces would be silently skipped by `mnmd filter` during completion, making that command undiscoverable.

Intended to be run by the CLI author during development and in CI. Not called on the completion or execution path.

---

### `mnmd args <args...>`

A helper for CLI authors writing command scripts. Parses the arguments passed to a command and outputs them in a structured form, making it easier to read named flags and positional arguments from any script language.

The exact output format is TBD.

---

## Shell Files

Shell files exist only where a technical constraint makes Go impossible — primarily because env vars, shell functions, and completion hooks must live in the parent shell process.

**Target shells: bash and zsh only.** monom targets macOS developers, who use bash or zsh. POSIX sh portability is explicitly out of scope — trying to maintain it restricts implementation options (e.g. no `BASH_SOURCE`, no bash arrays) without meaningful benefit to the target audience. Fish, dash, and other shells are not supported.


| File             | Purpose                                                                                                      |
| ---------------- | ------------------------------------------------------------------------------------------------------------ |
| `src/monom`      | Sourced by user's rc file. Defines `monom()`, `_setup_monom()`, and `_monom_cfg()`. Delegates to `mnmd`.    |
| `src/monom.bash` | Registers bash completion hook (`complete -F _monom_completion monom`).                                      |
| `src/monom.zsh`  | Registers zsh completion hook (`compdef _monom monom`).                                                      |


The aliasing feature (`make_monom_alias`) exists to let users bind a named command (e.g. `acme`) to a specific project root. How much of this lives in shell vs. Go is still being determined — the principle is to push as much as possible into `mnmd`.

No shell file should contain logic beyond what is technically impossible to move to Go.

---

## The User Config Interface

The monom config file (the executable `monom` at the project root) is the seam between monom and the author's project. It exposes one required subcommand and any number of optional hooks (see [Hooks](#hooks) below).

**Required:**

```
<monom-config-file> complete   # prints all discoverable command paths, slash-delimited, one per line
```

monom does not care how the user config is implemented — shell functions, Python, Go, whatever, as long as the required subcommand prints to stdout. The required interface is constitution-protected; changes require an amendment.

## Hooks

Hooks are optional subcommands the CLI author MAY expose on the monom config file to customize monom's default behavior. Each hook has a defined input/output contract and a defined fallback (what monom does when the hook is absent). Hooks are discovered by attempt-and-fallback at the call site — there is no separate registration step. The list of available hooks evolves here in `architecture.md` without requiring a constitution amendment.

### Hook: `run` — transform args before path resolution

Interposes between `monom <user_args...>` and `mnmd pack`. Receives the user's space-separated args, prints transformed space-separated args. monom passes the transformed output to `mnmd pack`.

```
$ _monom_cfg run acme deploy
infra cloud deploy
```

When the hook is present and produces usable output, monom uses it. When the hook is absent or doesn't produce usable output, monom falls back to passing the user's original args to `mnmd pack`. The exact detection-and-fallback semantics are left to the implementation.

Useful for: aliasing, namespace remapping, project-specific routing where the surface command tree differs from the file tree. This is also the sanctioned way to make a **command group runnable**: by default invoking a group (a directory) lists its children and exits non-zero (see [`mnmd pack`](#mnmd-pack-word) exit code 3), but a `run` hook can map a bare group token to a concrete leaf path, keeping that override explicit and per-project instead of a monom-wide catch-all.

---

## Environment Variables

These variables are internal shell↔Go plumbing. They are set by `src/monom` and read by `mnmd`. CLI authors and CLI users do not need to set or know these variables during normal use. The one public affordance is that a user MAY pre-set `$_MONOM_PROJECT_ROOT` to skip automatic project root discovery (useful when working outside a project tree or in a custom wrapper).

`MONOM_DEBUG_LOG` is intentionally unprefixed — it is a user-facing diagnostic toggle, not internal plumbing.


| Variable                | Set by                                                    | Description                                                                                                                                                                           |
| ----------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `_MONOM_LIB_ROOT`       | `src/monom` at source time                                | Absolute path to the monom install directory (`src/`).                                                                                                                                |
| `mnmd()` (function)     | `src/monom` at source time                                | Shell function wrapper that invokes `bin/mnmd`. Internal — not exported or user-facing.                                                                                               |
| `_MONOM_PROJECT_ROOT`   | `_setup_monom()` via `mnmd root` discovery, or user       | Path to the currently active monom project root. Pre-setting this skips auto-discovery. All call sites read this via the `mnmd root` algorithm.                                       |
| `_MONOM_USER_CONFIG`    | `_setup_monom()`                                          | Path to the monom config file — the `monom` executable at `$_MONOM_PROJECT_ROOT/monom`. Shell scripts invoke it via `_monom_cfg() { "$_MONOM_USER_CONFIG" "$@"; }` for readability. |
| `MONOM_DEBUG_LOG`       | user (optional)                                           | If set to a file path, `mnmd` and shell functions append timestamped debug lines to that file. Intentionally unprefixed: it is a user-facing diagnostic, not internal plumbing.      |


---

## Data Flow

### Completion (Tab press)

```
user presses Tab
  → _monom_completion() / _monom() [shell — registers COMPREPLY / calls compadd]
    → _monom_cfg complete                     [user's script — prints all paths, slash-delimited]
    → mnmd filter $COMP_WORDS                 [Go — always exits 0, prints matches]
    → COMPREPLY=(...) / compadd ...
```

### Command execution

```
user runs: monom <args...>
  → monom() [shell]
    → (optional) _monom_cfg run <args...>   [user hook — transforms args; falls back if absent or fails]
    → mnmd pack <args...>                   [Go — discovers root, joins with /, resolves to absolute path]
    → shell exec's the resolved path
```

The shell `monom()` function uses an attempt-and-fallback pattern: it tries `_monom_cfg run "$@"`, captures the output, and falls back to the user's original args if the hook is absent or produced no usable output. `mnmd pack` is then called with the resulting args. `mnmd pack` discovers the project root internally (via the same algorithm as `mnmd root`), so no separate setup step is needed on the execution path.

---
