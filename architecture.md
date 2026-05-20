# monom Architecture

This document describes the current intended architecture of monom. Unlike the constitution, it is descriptive and will evolve as the project develops. It should stay consistent with the principles in `constitution.md`.

---

## The Binary: monomd

There is one compiled Go binary: `monomd`.

All subcommands read environment variables at startup — see [Environment Variables](#environment-variables) for the full reference. In shell scripts, the `monom_cfg` function wrapper is used for readability at call sites:

```bash
monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }
```

---

### `monomd filter [word...]`

Reads command paths from stdin (one per line, slash-delimited) and accepts zero or more space-separated word arguments. Filters stdin to paths matching the given words, and prints the next-level completions to stdout (one per line).

**The two formats:** stdin uses `/` to separate category levels (e.g. `infra/cloud/deploy`) because that is how the CLI author's `complete` output encodes the command tree. The word arguments use spaces because they come directly from what the user typed — the shell passes them as separate args with no transformation. `monomd filter` bridges the two: it joins the word arguments with `/` internally to produce a prefix, then matches that prefix against the slash-delimited stdin paths.

Called by the shell completion binding as part of a pipe:

```bash
monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }
COMPREPLY=($(monom_cfg complete | monomd filter $COMP_WORDS))
```

The shell passes `$COMP_WORDS` (the raw typed tokens) directly — no transformation in shell. Stdin lines containing spaces in any path segment are silently ignored — emitting a hard error during a Tab press would produce noise in the terminal mid-typing. Use `monomd check` to surface these issues explicitly during development.

A trailing empty word in the argument list — which bash appends to `$COMP_WORDS` when the user has typed a complete token followed by a space — signals "drill into this level" rather than "match partially."

| User types | `$COMP_WORDS` (args to filter) | stdin format | filter output |
|---|---|---|---|
| `monom <Tab>` | _(none)_ | `category1/sub1\ncommand1` | `category1\ncommand1` |
| `monom com<Tab>` | `com` | `command1\ncommand2\ncategory1/sub1` | `command1\ncommand2` |
| `monom category1 <Tab>` | `category1` `""` | `category1/sub1\ncategory1/sub2` | `sub1\nsub2` |
| `monom category1 sub<Tab>` | `category1` `sub` | `category1/sub1\ncategory1/sub2` | `sub1\nsub2` |
| `monom category1 sub1 <Tab>` | `category1` `sub1` `""` | `category1/sub1/leaf` | `leaf` |

**Single-level example** — the `file_commands` test project:

```
# stdin: $monom_cfg complete output
category1/sub_command1
category1/sub_command2
command1
command2
```

```
# user types: monom <Tab>
$ monom_cfg complete | monomd filter
category1
command1
command2

# user types: monom com<Tab>
$ monom_cfg complete | monomd filter com
command1
command2

# user types: monom categ<Tab>
$ monom_cfg complete | monomd filter categ
category1

# user types: monom category1 <Tab>  (trailing space → bash appends "")
$ monom_cfg complete | monomd filter category1 ""
sub_command1
sub_command2
```

**Nested example** — a project with two levels of categories:

```
# stdin: $monom_cfg complete output
infra/cloud/deploy
infra/cloud/teardown
infra/local/start
infra/local/stop
release
```

```
# user types: monom <Tab>
$ monom_cfg complete | monomd filter
infra
release

# user types: monom infra <Tab>
$ monom_cfg complete | monomd filter infra ""
cloud
local

# user types: monom infra cl<Tab>
$ monom_cfg complete | monomd filter infra cl
cloud

# user types: monom infra cloud <Tab>
$ monom_cfg complete | monomd filter infra cloud ""
deploy
teardown
```

Stdin lines with spaces in any path segment are silently ignored and excluded from completions.

**Why a pipe instead of a single `monomd` call:**
Calling the user config is a subprocess spawn either way. Shell pipes natively; Go needs goroutines and io plumbing to do the same. The rejected alternative:

```bash
# ❌ Rejected: COMPREPLY=($(monomd complete "$prefix"))
```
```go
// ❌ What monomd would do internally:
cmd := exec.Command(os.Getenv("MONOM_USER_CONFIG"), "complete")
out, _ := cmd.Output()
matches := filterByPrefix(strings.Split(strings.TrimSpace(string(out)), "\n"), prefix)
```

---

### `monomd pack <args...>`

Called by the `monom()` shell function when the user executes a command. Reads a file path from stdin (output of `monom_cfg run <args...>`) and resolves it to an absolute path, printing it to stdout. The shell then `exec`s that path.

```bash
monom_cfg run category1 sub_command1 | monomd pack
# → /path/to/project/category1/sub_command1.sh
```

---

### `monomd root`

Walks up from `$PWD`, looking for a directory containing a `monom` file. Prints the absolute path of the first one found, or exits non-zero if none is found.

Used by `setup_monom()` when `MONOM_PROJECT_ROOT` is not already set.

```
$ monomd root
/path/to/project
```

---

### `monomd check`

Validates that the current monom project is healthy. Runs `monom_cfg complete`, inspects every path in the output, and reports any problems to stdout. Exits non-zero if any problems are found.

Currently checks:
- Every path is slash-delimited with no spaces in any segment. A path with spaces would be silently skipped by `monomd filter` during completion, making that command undiscoverable.

Intended to be run by the CLI author during development and in CI. Not called on the completion or execution path.

---

### `monomd args <args...>`

A helper for CLI authors writing command scripts. Parses the arguments passed to a command and outputs them in a structured form, making it easier to read named flags and positional arguments from any script language.

The exact output format is TBD.

---

## Shell Files

Shell files exist only where a technical constraint makes Go impossible — primarily because env vars, shell functions, and competion hooks must live in the parent shell process.

| File | Purpose |
|---|---|
| `src/monom` | Sourced by user's rc file. Defines `monom()` and `setup_monom()`. Delegates everything to `monomd`. |
| `src/monom.bash` | Registers bash completion hook (`complete -F monom_completion monom`). |
| `src/monom.zsh` | Registers zsh completion hook (`compdef _monom monom`). |

The aliasing feature (`make_monom_alias`) exists to let users bind a named command (e.g. `acme`) to a specific project root. How much of this lives in shell vs. Go is still being determined — the principle is to push as much as possible into `monomd`.

No shell file should contain logic beyond what is technically impossible to move to Go.

---

## The User Config Interface

The user config (`MONOM_USER_CONFIG`) exposes two subcommands that monom calls:

```
$MONOM_USER_CONFIG complete   # prints all discoverable command paths, one per line
$MONOM_USER_CONFIG run        # reads args, prints the resolved file path to execute
```

monom does not care how these are implemented — shell functions, Python, Go, whatever. The user config is the seam between monom's engine and the author's project.

---

## Environment Variables

| Variable | Set by | Description |
|---|---|---|
| `MONOM_LIB_ROOT` | `src/monom` at source time | Absolute path to the monom install directory. |
| `MONOM_PROJECT_ROOT` | user override, alias, or `monomd root` discovery | Path to the currently active monom project root. A single user can have many monom projects; this holds whichever is active for the current invocation. Can be set before sourcing monom to skip auto-discovery. |
| `MONOM_USER_CONFIG` | `setup_monom()` | Path to the user's config entry point — the `monom` executable at `$MONOM_PROJECT_ROOT/monom`. This file is written by the CLI author and exposes the `complete` and `run` subcommands that monom calls. In shell scripts, wrap it as `monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }` for readability. |

---

## Data Flow

### Completion (Tab press)

```
user presses Tab
  → monom_completion() [shell — registers COMPREPLY]
    → monom_cfg complete                 [user's script — prints all paths]
    → monomd filter <prefix>             [Go — filters by prefix, prints matches]
    → COMPREPLY=(...)
```

### Command execution

```
user runs: monom <args...>
  → monom() [shell]
    → monom_cfg run <args...>      [user's script — prints raw command path]
    → monomd pack                  [Go — resolves to absolute file path]
    → shell exec's the resolved path
```

---

