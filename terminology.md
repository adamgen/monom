# Terminilogy

The terminology for monom is critical since it's used by llm agents to understand and then execute effectively on task that are related to monom developers and clients.

## Background

monom is a CLI tool framework.

How it works: Organize your executable scripts in a folder structure, add a `monom` configuration file that defines how to discover and run commands, and monom automatically creates a full-featured CLI with tab completion. Your file tree becomes your command tree - folders become command categories, and scripts become executable commands.

## Terms

**Discovery** — the process by which monom finds all possible commands for a single project. Discovery can be as simple or complex as required. Implemented by the `complete` subcommand of the monom config file.

**Command packing** — the process by which a command invocation is transformed back into the full path of the executable to run. Implemented by `monomd pack`, which reads the raw path printed by `monom_cfg run <args...>` from stdin and resolves it to an absolute file path.

**monom config file** — the `monom` executable at the project root, written by the CLI author. Exposes `complete` and `run`. Referenced via env var `$MONOM_USER_CONFIG`. In shell scripts, wrapped as `monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }` for readability.

**Project root** — the directory containing the monom config file. Referenced via env var `$MONOM_PROJECT_ROOT`.

**monomd** — the compiled Go binary. The engine of monom. Implements all internal logic.

**CLI author** — the developer building a CLI tool using monom.

**CLI user** — the developer using the CLI the author built.
