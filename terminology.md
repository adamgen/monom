# Terminilogy

The terminology for monom is critical since it's used by llm agents to understand and then execute effectively on task that are related to monom developers and clients.

## Background

monom is a CLI tool framework.

How it works: Organize your executable scripts in a folder structure, optionally add a `monom` configuration file that customizes how commands are discovered and run, and monom automatically creates a full-featured CLI with tab completion. Your file tree becomes your command tree - folders become command categories, and scripts become executable commands.

## Terms

**Discovery** — the process by which monom finds all possible commands for a single project. Discovery can be as simple or complex as required. Implemented by the `complete` hook of the monom config file when it prints anything, otherwise by default discovery.

**Default discovery** — monom's built-in discovery, `mnmd discover`: a broad scan of the project root for executables, narrowed by the discovery gate. What makes a project **zero-config**: it needs no monom config file, or an empty one.

**Discovery gate** — the rule a scanned executable must pass to be registered: it has a shebang, its name matches the naming pattern (`^[a-z0-9][a-z0-9_-]*$`), or it is declared in a declarative monom config file.

**Registered command** — a command path that is part of the CLI: the `complete` hook's output, or default discovery's gated set plus declarations. Completion, group listings, and `mnmd check` all see exactly the registered set.

**Command packing** — the process by which a command invocation is transformed back into the full path of the executable to run. Implemented by `mnmd pack`, which takes the user's space-separated command tokens as CLI arguments, joins them with `/`, resolves against the project root, and prints the absolute executable path to stdout.

**monom config file** — the optional file named `monom` at the project root, written by the CLI author. Either a **hook script** (an executable file starting with a shebang, exposing optional hooks such as `complete`, `run`, `debug`, `config`) or a **declarative config file** (any other file, including an empty one: command declarations and `key = value` settings, parsed and never run). The shell binding locates it as `<project-root>/monom` and runs a hook script via the internal `_monom_cfg` helper. The internal env var `$_MONOM_USER_CONFIG` holds its resolved path as a shell-to-Go plumbing detail; authors do not need to know or set it.

**Project root** — the directory that is monom's base for a project. Resolved as: `$_MONOM_PROJECT_ROOT` when it names a directory (what an alias pins); else the nearest directory upward from `$PWD` containing a file named `monom`; else the nearest git root. Pre-setting `$_MONOM_PROJECT_ROOT` is the one public affordance of the internal shell↔Go plumbing.

**mnmd** — the compiled Go binary. The engine of monom. Implements all internal logic: project root discovery, completion filtering, command resolution, and more. Sourcing `src/monom` defines a user-facing `mnmd()` shell function, so `mnmd` is callable by name in the user's shell without `bin/` being on `$PATH`.

**CLI author** — the developer building a CLI tool using monom.

**CLI user** — the developer using the CLI the author built.
