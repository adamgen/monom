# Monom Architecture

This document describes the current intended architecture of monom. Unlike the constitution, it is descriptive and will evolve as the project develops. It should stay consistent with the principles in `constitution.md`.

---

## The Binary: monomd

There is one compiled Go binary: `monomd`.

All subcommands read environment variables at startup — see [Environment Variables](#environment-variables) for the full reference. In shell scripts, the `monom_cfg` function wrapper is used for readability at call sites:

```bash
monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }
```

---

### `monomd filter <prefix>`

Reads command paths from stdin (one per line), filters to those matching `<prefix>`, and prints the matches to stdout.

Called by the shell completion binding as part of a pipe:

```bash
monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }
COMPREPLY=($(monom_cfg complete | monomd filter "$prefix"))
```

`<prefix>` is the partial command the user has typed so far, with spaces replaced by `/` (e.g. `monom category1 sub` becomes `category1/sub`). An empty prefix returns all top-level commands.

The following example uses the `file_commands` test project, whose `complete` output looks like this:

```
# $monom_cfg complete output for file_commands:
category1/sub_command1
category1/sub_command2
command1
command2
```

Piped through `monomd filter`:

```
$ monom_cfg complete | monomd filter ""
category1
command1
command2

$ monom_cfg complete | monomd filter "com"
command1
command2

$ monom_cfg complete | monomd filter "category1/"
sub_command1
sub_command2

$ monom_cfg complete | monomd filter "categ"
category1
```

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

### `monomd run <args...>`

Called by the `monom()` shell function when the user executes a command.

- Passes `<args...>` to `$MONOM_USER_CONFIG run` to resolve the command to a file path.
- Prints the resolved absolute path to stdout.
- The shell then `exec`s that path.

```
$ monomd run category1 sub_command1
/path/to/project/category1/sub_command1.sh
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

Monom does not care how these are implemented — shell functions, Python, Go, whatever. The user config is the seam between monom's engine and the author's project.

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
    → monomd run <args...>
        → calls monom_cfg run
        → prints resolved absolute file path
    → shell exec's the resolved path
```

---

