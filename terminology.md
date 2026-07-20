# Terminilogy

The terminology for monom is critical since it's used by llm agents to understand and then execute effectively on task that are related to monom developers and clients.

## Background

monom is a CLI tool framework.

How it works: Organize your executable scripts in a folder structure, add a `monom` configuration file that defines how to discover and run commands, and monom automatically creates a full-featured CLI with tab completion. Your file tree becomes your command tree - folders become command categories, and scripts become executable commands.

## Terms

**Discovery** — the process by which monom finds all possible commands for a single project. Discovery can be as simple or complex as required. Implemented by the `complete` subcommand of the monom config file.

**Command packing** — the process by which a command invocation is transformed back into the full path of the executable to run. Implemented by `mnmd pack`, which takes the user's space-separated command tokens as CLI arguments, joins them with `/`, resolves against the project root, and prints the absolute executable path to stdout.

**monom config file** — the `monom` executable at the project root, written by the CLI author. Exposes `complete` and optional hooks such as `run`. The shell binding locates it as `<project-root>/monom` and invokes it via the internal `_monom_cfg` helper. The internal env var `$_MONOM_USER_CONFIG` holds its resolved path as a shell-to-Go plumbing detail; authors do not need to know or set it.

**Project root** — the directory containing the monom config file (the executable `monom` file). monom discovers it by walking upward from `$PWD`. Authors may pre-set `$_MONOM_PROJECT_ROOT` to skip discovery; this is an internal shell↔Go plumbing affordance, not a required step.

**Command mapping** — table-driven resolution of command words to a value (typically an executable path), as opposed to the convention-driven resolution of command packing. One mapping table — a string of `<value> <key words...>` lines, passed as the first CLI argument — drives both hooks: `mnmd resolve-run` prints the value of the first entry whose key words equal the remaining args — or the args unchanged when nothing matches, so a miss passes through to default resolution — and `mnmd resolve-complete` prints each entry's key as a slash-delimited path for completion discovery. Values may be absolute paths (typically built from the config's own directory) or project-root-relative. Used chiefly while migrating an existing project to monom, when legacy command names must map to files that do not yet follow the file-tree convention.

**mnmd** — the compiled Go binary. The engine of monom. Implements all internal logic: project root discovery, completion filtering, command resolution, and more. Sourcing `src/monom` defines a user-facing `mnmd()` shell function, so `mnmd` is callable by name in the user's shell without `bin/` being on `$PATH`. Hook scripts run as child processes and cannot see that function; every config spawn site instead prepends the binary's directory to the child's `$PATH`, so hooks also call `mnmd` by name.

**CLI author** — the developer building a CLI tool using monom.

**CLI user** — the developer using the CLI the author built.

**alias** — a named command bound to a fixed monom project root, created via `mnmd alias <name> <path>`. Invoking the alias (`<name> …`) behaves identically to invoking `monom …` from inside that project — same discovery, hooks, completion, and command packing — except the project root is pinned to `<path>` instead of discovered by walking up from `$PWD`. This lets multiple monom projects be active under distinct command names in a single shell session. An alias is implemented as a generated shell function plus a matching completion binding, **not** a shell `alias` builtin — the builtin cannot carry the pinned root nor bind completion.
