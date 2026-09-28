# monom Architecture

This document describes the current intended architecture of monom. Unlike the constitution, it is descriptive and will evolve as the project develops. It should stay consistent with the principles in `constitution.md`.

It states **contracts** — what each part promises, and how the parts fit together. It deliberately does not argue for them. Where a decision was contested, the reasoning and the rejected alternatives live in a `TRADEOFFS.md` beside the code that implements it (`internal/filter/`, `internal/pack/`, `internal/cli/`, `internal/install/`, `src/`). Read that file before changing the behavior it describes.

---

## Entry Points

monom has exactly three entry points — the complete public surface through which anything interacts with it. Everything else (`mnmd` subcommands, hooks, internal env vars) is machinery reached *through* these three.

1. **`source monom`** — bootstrap. The user sources `src/monom` from their rc file. This is the one-time introduction of monom into a shell session: it resolves `_MONOM_LIB_ROOT`, defines the `mnmd()`, `monom()`, `_setup_monom()`, and `_monom_cfg()` functions, and sources the shell-specific completion binding (`src/monom.bash` or `src/monom.zsh`). After this, the shell knows the `monom` command and how to complete it. No `mnmd` subcommand runs at source time. Sourcing deliberately adds `mnmd` to the user's namespace: `mnmd()` is a user-facing shell function that makes the binary callable by name without requiring `bin/` on `$PATH`.

2. **`monom <Tab>`** — completion. The user presses Tab while typing a `monom` command. The registered completion function runs discovery (`_monom_cfg complete`) and filters it (`mnmd filter`) to populate `COMPREPLY`. See [Completion (Tab press)](#completion-tab-press) for the full flow. This path must be fast and never error mid-typing.

3. **`monom [command...]`** — execution. The user runs a resolved command. The `monom()` function optionally transforms args via the `run` hook, resolves them to an absolute executable path (`mnmd pack`), and `exec`s it. See [Command execution](#command-execution) for the full flow.

Entry points 2 and 3 both depend on entry point 1 having run first in the session.

`mnmd install` sits before all three: it writes the `source` line that makes entry point 1 happen on every new shell. It is the only `mnmd` subcommand a user is expected to invoke by hand.

---

## Design Principles

Operational guidelines for designing tools within the monom project (currently `mnmd`, and any future tools we build). Unlike the constitution's principles, these are project-design heuristics — they SHOULD be followed, but deviations are allowed when justified. These principles do not apply to the user config interface; CLI authors choose their own conventions there.

### Principle: CLI Arguments by Default, Stdin When the Input Is a Stream

Subcommand inputs SHOULD be CLI arguments. Stdin SHOULD only be used when the input is genuinely a stream — many lines, unbounded data, content piped from another tool, or cases where streaming improves correctness or composability.

**The test:** Before designing a subcommand to read stdin, ask — "is this a parameter or a stream?" Parameters are bounded, named, and known at call time; they belong in args. Streams are unbounded, anonymous, and benefit from pipe composition; they belong in stdin.

Examples in this codebase:
- `mnmd filter` reads commands from stdin — the command list is unbounded and naturally produced by piping `_monom_cfg complete`. Stream.
- `mnmd pack` and `mnmd map resolve` take args — the user's command tokens are a small, known parameter set produced by the shell at call time, not a stream. Parameters.
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

The shell passes `${COMP_WORDS[@]:1}` (the raw typed tokens after `monom`) directly — no transformation in shell. Stdin lines containing spaces in any path segment are silently ignored; `mnmd check` surfaces them explicitly.

**`mnmd filter` never exits with a non-zero status code.** This is a hard constraint, not a guideline. Any internal error produces empty output and exit 0. All diagnostics belong in `mnmd check`.

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

---

### `mnmd pack <word...>`

Called by the `monom()` shell function when the user executes a command. Takes the user's space-separated command path as CLI args (e.g. `mnmd pack category1 sub_command1`), joins the tokens with `/` internally, resolves the result against the project root, and prints the absolute path to stdout. The shell then `exec`s that path.

Pack is self-sufficient: it discovers the project root itself (same algorithm as `mnmd root` — see below), validates the file exists and is executable, and prints the absolute path. The shell `monom()` function reduces to a single `exec` call.

```bash
$ mnmd pack category1 sub_command1
/path/to/project/category1/sub_command1
```

Pack is the symmetric counterpart of `filter`: both take space-separated tokens as CLI args and bridge to the slash-delimited file tree. Pack's specific job is to replace spaces with slashes and resolve to an absolute, executable path.

**Exit codes.** Pack signals its outcome through the exit code so the shell can branch without parsing strings. All exit codes are defined in the central registry at `internal/cli/cli.go` (see constitution: *Errors Carry Their Own Exit Code*).

Exit code `3` is **reserved for the group signal** — the tokens name a category, not a command. From pack that means they resolved to a directory rather than an executable; the same signal may come from a `run` hook (via [`mnmd map resolve`](#mnmd-map-complete--mnmd-map-resolve-word)) for a map category that has no directory on disk. A category is a noun in monom's noun→verb command tree: not a command, but not a failure either. Exit 3 is a *pure signal* — the emitter writes nothing to stdout or stderr and does not enumerate the group's children. Rendering the group for the user is the shell layer's job; see [shell files](#shell-files).

A command group is never runnable by default. An author who wants `monom infra` to *do* something expresses that in their own config via the [`run` hook](#hook-run--transform-args-before-path-resolution), e.g. mapping `infra` → `infra cloud deploy`.

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

When a command map (`monom-map.json`) exists at the project root, check also validates it:

- The file parses. Duplicate keys are a hard error — JSON decoders otherwise keep the last occurrence silently, swallowing a command.
- Every run target exists, is a regular file, is executable, and stays inside the project root.
- No map entry shadows the file tree. A map command over any existing tree path, or a map category over a tree *file*, makes that tree entry unreachable (the run hook fires before pack). A map category over a tree *directory* merges — unmapped children fall back to pack — and is fine.

Intended to be run by the CLI author during development and in CI. Not called on the completion or execution path.

---

### `mnmd map complete` / `mnmd map resolve <word...>`

The **command map** backend: an implementation of the `complete` and `run` hooks that authors wire into their config file, for brownfield projects whose scripts are scattered outside a clean command tree. `mnmd pack` remains the sole resolver — `map resolve` only rewrites the user's tokens into the tokens pack receives; it never validates or executes targets itself.

Both subcommands read `monom-map.json` at the project root (discovered via the `mnmd root` algorithm). The node model mirrors the file tree: a **string value** is a command whose run target is that root-relative path (shorthand for `{"run": <target>}`); an **object with a `"run"` key** is a command node; any **other object** is a category whose keys are children. `"$note"` is allowed anywhere and ignored — it stands in for the comments JSON lacks. The keys `complete`, `pre-run`, and `post-run` are reserved for future per-command hooks and rejected today, as is a node with both a `"run"` target and children.

```json
{
  "db": {
    "migrate": "scripts/legacy/run_migrations.sh"
  },
  "deploy": { "run": "ops/full_deploy.sh" },
  "release": "tools/make_release.sh"
}
```

`map complete` prints the map's command paths, slash-delimited and sorted, one per line — the `complete` hook wire format. `map resolve` takes the user's tokens as CLI args (parameters, like pack) and has three outcomes:

- **Command node matched** — prints the target as space-separated path tokens (`scripts legacy run_migrations.sh`), exit 0. This is the `run` hook output format; pack re-joins the tokens with `/`.
- **Category node matched** — exit 3, the payload-free [group signal](#mnmd-pack-word), exactly like pack's directory outcome.
- **No match** (unknown name, or tokens continuing past a command) — prints nothing, exit 0. Empty output triggers the run-hook fallback, so unresolved tokens drop through to `mnmd pack` against the file tree: **the map is an overlay on the tree, not a replacement.** Mapped legacy scripts and native tree commands coexist, and migrating a script into the tree is just deleting its map entry.

The author's config file wiring, in full:

```bash
#!/usr/bin/env bash
case "$1" in
  complete) mnmd map complete ;;   # plus any tree discovery, for hybrid projects
  run)      shift; mnmd map resolve "$@" ;;
esac
```

A missing or unparsable map file is a loud error (exit 1, stderr): a config that delegates to `mnmd map` without a valid map is misconfigured. `mnmd check` reports map problems proactively. See `fixtures/brownfield-project/` for a complete runnable example, and `internal/mapfile/TRADEOFFS.md` for the contested decisions.

---

### `mnmd install`

Activates the shell integration. Detects the user's shell from `$SHELL`, resolves the absolute path to `src/monom` relative to the running binary, selects the rc/profile file (`~/.zshrc` for zsh; `~/.bash_profile` for bash, falling back to `~/.bashrc`), and appends `source "<path>"` if no non-comment line already references that path. Idempotent: a second run prints `already installed` and changes nothing. Exits non-zero for any shell other than bash or zsh.

```
$ mnmd install
added to /Users/me/.zshrc
restart your shell or run: source /Users/me/.zshrc
```

### The activation nudge

When `mnmd` runs with `$MONOM_ACTIVE` unset — meaning no shell has sourced `src/monom` — every subcommand except `install` prints a one-line hint to **stderr** before doing its work:

```
hint: run 'mnmd install' to activate shell integration
```

Stdout is unaffected, so `$(mnmd pack ...)` and the completion pipe are not polluted. The subcommand's own exit code is unchanged.

---

## Shell Files

Shell files exist only where a technical constraint makes Go impossible — primarily because env vars, shell functions, and completion hooks must live in the parent shell process.

**Target shells: bash and zsh only.** monom targets macOS developers, who use bash or zsh. POSIX sh portability is explicitly out of scope — trying to maintain it restricts implementation options (e.g. no `BASH_SOURCE`, no bash arrays) without meaningful benefit to the target audience. Fish, dash, and other shells are not supported.


| File             | Purpose                                                                                                      |
| ---------------- | ------------------------------------------------------------------------------------------------------------ |
| `src/monom`      | Sourced by user's rc file. Exports `_MONOM_LIB_ROOT` and `MONOM_ACTIVE`, defines `mnmd()`, `monom()`, `_setup_monom()`, `_monom_cfg()`, and `_monom_log()`, then sources the shell-specific binding based on `$ZSH_VERSION` / `$BASH_VERSION`. |
| `src/monom.bash` | Registers bash completion hook (`complete -F _monom_completion monom`).                                      |
| `src/monom.zsh`  | Registers zsh completion hook (`compdef _monom monom`).                                                      |


**Rendering a command group.** The group signal (exit 3, from `pack` or from a `run` hook) carries no payload, so `monom()` produces the user-facing message itself. It names the group by the last token the user typed and lists the children by re-running the discovery pipeline — `_monom_cfg complete | mnmd filter <tokens> ""`, where the trailing empty word drills into the level. This is the same pipeline tab completion uses, so the listing always matches `monom <group> <Tab>`. When the pipeline yields nothing, the `available:` line is omitted.

```
monom: 'infra' is a command group
available: cloud, local
```

The aliasing feature (`make_monom_alias`) exists to let users bind a named command (e.g. `acme`) to a specific project root. Not yet implemented against `mnmd`; the principle is to push as much as possible into Go.

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

The hook's **exit code selects the behavior**:

| Hook result | `monom()` does |
| ----------- | -------------- |
| exit 0, empty stdout | Hook absent, or present and declining. Falls back to the user's original args. These two cases are deliberately indistinguishable. |
| exit 0, non-empty stdout | Splits the output into words and passes them to `mnmd pack` in place of the original args. |
| exit 3 | The group signal: the typed tokens name a category the hook owns (e.g. a command-map category with no directory on disk). Renders the same command-group listing as pack's exit 3 and returns 1. Any hook output is ignored — the signal is payload-free by contract. |
| other non-zero exit | Hook present and failed. Forwards the hook's stderr to the user, returns the hook's exit code, and does not call `mnmd pack` or exec anything. |

Exit 3 is therefore not available to run hooks as an ordinary failure code — see `internal/cli/TRADEOFFS.md`.

The hook may change the *number* of args — that is its purpose. Because a hook is a separate process, it receives argv but can only emit a flat stdout stream, so `monom()` re-splits that stream before handing it to `pack`:

```
monom db migrate
  → "$@"  = ["db", "migrate"]
  → _monom_cfg run db migrate                   # IN: separate args
        ↳ prints "custom-folder db migrate\n"   # OUT: one flat stream (2 → 3 args)
  → (monom re-splits the stream)
  → mnmd pack custom-folder db migrate          # IN: separate args
        ↳ resolves custom-folder/db/migrate
```

Useful for: aliasing, namespace remapping, project-specific routing where the surface command tree differs from the file tree. This is also the sanctioned way to make a **command group runnable** — see [`mnmd pack`](#mnmd-pack-word) exit code 3.

### Hook: `debug` — project-local debug log path

Lets a project route debug logging to its own file — including turning logging on for itself while the global switch is off. Takes no input; prints a single absolute file path on stdout, or nothing.

```
$ _monom_cfg debug
/home/me/proj/.monom-debug.log
```

`_setup_monom` queries the hook once per invocation, on both the completion (Tab) and execution paths. Validation and precedence:

- Hook prints a **single-line path that is writable** → that path is exported as `MONOM_DEBUG_LOG` for the invocation: **local overrides global**.
- Hook prints **multiline output** (invalid) or a **single-line path that is not writable** → monom falls back to the inherited global `MONOM_DEBUG_LOG` (set or unset) and emits a diagnostic: a stderr warning on the command path (the command still runs); via `_monom_log` on the completion path, which never writes to stderr mid-Tab.
- Hook is **absent or prints nothing** → the global value stands untouched.

Both read sites (`_monom_log` in the shell and `debuglog.Log` in Go) keep reading `MONOM_DEBUG_LOG`; the export in `_setup_monom` is the single resolution point, so the layers cannot disagree.

Cost: one unconditional subprocess spawn per invocation, plus one writability check when the hook prints a path.

---

## Environment Variables

These variables are internal shell↔Go plumbing. They are set by `src/monom` and read by `mnmd`. CLI authors and CLI users do not need to set or know these variables during normal use. The one public affordance is that a user MAY pre-set `$_MONOM_PROJECT_ROOT` to skip automatic project root discovery (useful when working outside a project tree or in a custom wrapper).

`MONOM_DEBUG_LOG` and `MONOM_ACTIVE` are intentionally unprefixed — they are user-facing, not internal plumbing.


| Variable                | Set by                                                    | Description                                                                                                                                                                           |
| ----------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `_MONOM_LIB_ROOT`       | `src/monom` at source time                                | Absolute path to the monom install directory (`src/`).                                                                                                                                |
| `MONOM_ACTIVE`          | `src/monom` at source time                                | Set to `1` so subprocesses can detect that the shell integration is live. Its absence is what triggers the [activation nudge](#the-activation-nudge). User-facing: set it to suppress the nudge in scripts. |
| `mnmd()` (function)     | `src/monom` at source time                                | Shell function wrapper that invokes `bin/mnmd`. User-facing — makes `mnmd` callable by name after sourcing, without adding `bin/` to `$PATH`.                                          |
| `_MONOM_PROJECT_ROOT`   | `_setup_monom()` via `mnmd root` discovery, or user       | Path to the currently active monom project root. Pre-setting this skips auto-discovery. All call sites read this via the `mnmd root` algorithm.                                       |
| `_MONOM_USER_CONFIG`    | `_setup_monom()`                                          | Path to the monom config file — the `monom` executable at `$_MONOM_PROJECT_ROOT/monom`. Shell scripts invoke it via `_monom_cfg() { "$_MONOM_USER_CONFIG" "$@"; }` for readability. |
| `MONOM_DEBUG_LOG`       | user (optional), or `_setup_monom()` via the `debug` hook | If set to a file path, `mnmd` and shell functions append timestamped debug lines to that file. Intentionally unprefixed: it is a user-facing diagnostic, not internal plumbing. A project may override the global value via the [`debug` hook](#hook-debug--project-local-debug-log-path) when the hook path is valid (single-line) and writable. |


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
    → _setup_monom                          [resolves root, config path, effective MONOM_DEBUG_LOG]
    → (optional) _monom_cfg run <args...>   [user hook — transforms args; falls back when it exits 0 with no output;
                                             exit 3 → shell renders the command-group listing, returns 1]
    → mnmd pack <args...>                   [Go — discovers root, joins with /, resolves to absolute path]
    → exit 0 → shell exec's the resolved path in a subshell
      exit 3 → shell renders the command-group listing, returns 1
      other  → shell forwards pack's stderr, returns 1
```

`mnmd pack` discovers the project root internally (via the same algorithm as `mnmd root`), so no separate setup step is needed for it on the execution path.

---
