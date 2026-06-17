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

**mnmd** — the compiled Go binary. The engine of monom. Implements all internal logic: project root discovery, completion filtering, command resolution, and more.

**CLI author** — the developer building a CLI tool using monom.

**CLI user** — the developer using the CLI the author built.
