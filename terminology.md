# Terminilogy

The terminology for monom is critical since it's used by llm agents to understand and then execute effectively on task that are related to monom developers and clients.

## Background

Monom is a CLI tool framework.

How it works: Organize your executable scripts in a folder structure, add a `monom` configuration file that defines how to discover and run commands, and monom automatically creates a full-featured CLI with tab completion. Your file tree becomes your command tree - folders become command categories, and scripts become executable commands.

## Terms

`Discovery` is the process by which monom finds all possible commands for a single project. The discovery can be as simple or complex as is required, and is done, like everythin in monom, with an executable file.

The implementation of the discovery feature is the complete command that is available on the monom config file.

`Command packing` When you run a command the packing is the process by which the parameters passed to monom (`$*`) transformed back into a file path.
