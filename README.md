# monom

monom is a CLI framework that turns a file tree into a tab-completable command tree. Organize your scripts in folders, add a config file, and monom gives you a full-featured CLI with tab completion — in any shell, for any scripting language.

**Your file tree is your command tree.** Folders become command categories. Scripts become commands. No boilerplate, no registration, no framework lock-in.

## Philosophy

monom is built on a few hard principles:

- **Go owns logic, shell owns surface.** All decision-making — discovery, filtering, resolution, argument parsing — lives in a compiled Go binary (`mnmd`). Shell code exists only where technically unavoidable: sourcing into the parent process, registering completion hooks, and exec-ing commands.
- **Minimize subprocess roundtrips.** Every process boundary must justify its existence. Pipes and subprocesses are used only when there is no alternative.
- **Language-agnostic commands.** Commands can be shell, Python, Node, Ruby — anything with a shebang. monom doesn't care how your scripts are written.
- **Speed is non-negotiable.** Tab completion must feel instant. monom's own overhead should be imperceptible.
- **Testability by design.** Go logic is unit-tested in Go. CLI surface behavior is tested end-to-end with shUnit2. The two layers never conflate.

## Architecture

```
┌──────────────────────────────────────────────────┐
│  CLI User types: my-tool <Tab>                   │
│                  my-tool category1 sub_command1   │
└──────────────┬───────────────────────┬───────────┘
               │ completion            │ execution
               ▼                       ▼
        _monom_completion()         monom()          ← thin shell functions
               │                       │
     ┌─────────┴──────────┐    ┌───────┴────────┐
     │ _monom_cfg complete │    │ _monom_cfg run  │   ← user's config file
     │ mnmd filter <pfx>   │    │ mnmd pack       │   ← Go binary
     └─────────────────────┘    └───────┬────────┘
                                        │
                                  exec resolved path
```

There are two roles:

- **CLI Author** — writes the config file and command scripts. Works against monom's interface.
- **CLI User** — types commands and hits Tab. Never knows monom exists.

The user's config file (`monom` at the project root) exposes two subcommands: `complete` (list all discoverable command paths) and `run` (resolve args to a file path). This is the seam between monom's engine and the author's project.

## Project Status

monom is in **early planning and development**. This is the second attempt at building it — the first stalled out. This time around, the constraint is very little available time, so the approach leans heavily on AI-assisted development and [OpenSpec](https://openspec.dev/) for structured planning. The core design documents (constitution, architecture, terminology) are written and stable. An initial implementation exists with a working Go binary, shell bindings for bash/zsh, tab completion, and an end-to-end test harness. Active work is focused on solidifying the binary's subcommands and refining the shell integration.
