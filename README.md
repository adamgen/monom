# monom

monom is a CLI framework that turns a file tree into a tab-completable command tree. Organize your scripts in folders, and monom gives you a full-featured CLI with tab completion — in any shell, for any scripting language.

**Your file tree is your command tree.** Folders become command categories. Scripts become commands. No boilerplate, no registration, no framework lock-in.

## Philosophy

monom is built on a few hard principles:

- **Go owns logic, shell owns surface.** All decision-making — discovery, filtering, resolution, argument parsing — lives in a compiled Go binary (`mnmd`). Shell code exists only where technically unavoidable: sourcing into the parent process, registering completion hooks, and exec-ing commands.
- **Minimize subprocess roundtrips.** Every process boundary must justify its existence. Pipes and subprocesses are used only when there is no alternative.
- **Language-agnostic commands.** Commands can be shell, Python, Node, Ruby — anything with a shebang. monom doesn't care how your scripts are written.
- **Speed is non-negotiable.** Tab completion must feel instant. monom's own overhead should be imperceptible.
- **Testability by design.** Go logic is unit-tested in Go. CLI surface behavior is tested end-to-end with shUnit2. The two layers never conflate.

## Install

Linux or macOS, bash or zsh:

```sh
curl -fsSL https://raw.githubusercontent.com/adamgen/monom/main/install.sh | bash
```

[`install.sh`](install.sh) is short; read it first if you like. It downloads the prebuilt `mnmd` for your OS and architecture (linux/darwin, amd64/arm64) from the latest [GitHub release](https://github.com/adamgen/monom/releases), verifies its SHA-256 against the release's `checksums.txt`, installs it with the shell integration into `~/.local/share/monom` (linking `~/.local/bin/mnmd`), and runs `mnmd install` to add one `source` line to your rc file. No sudo, no make. Re-running it upgrades in place and never adds the line twice.

If there's no prebuilt binary for your machine, it builds `mnmd` from source instead, which needs Go 1.24+; without Go it stops with instructions. Environment variables override the defaults: `MONOM_VERSION` (a release tag), `MONOM_INSTALL_DIR`, `MONOM_BIN_DIR`, `MONOM_NO_MODIFY_RC=1`, `MONOM_FROM_SOURCE=1`. The full list is at the top of the script.

**From a checkout** (to work on monom): `./build.sh && bin/mnmd install`. `build.sh` needs only Go; `make build` calls it. Releases are cut by pushing a `v*` tag, which runs [`.github/workflows/release.yml`](.github/workflows/release.yml).

To have a coding agent do it, point it at [`INSTALL.md`](INSTALL.md). It covers checking your shell, verifying the install, and the common pitfalls.

## Quick start: zero-config

A project needs no monom config file. From any git repository, or any directory with an empty `monom` file in it (`touch monom`), or one pinned by an alias:

```sh
alias mon='_MONOM_PROJECT_ROOT=~/code/my-project monom'
```

monom scans the project for executables and turns them into the command tree. Only executables that pass the **discovery gate** become commands:

- a file with a shebang (`#!/usr/bin/env bash`, `#!/usr/bin/env python3`, …), or
- a file named like a command, `^[a-z0-9][a-z0-9_-]*$` (lowercase, no extension — this admits compiled binaries), or
- a path declared in the `monom` file.

Hidden and `_`-prefixed files and directories, `node_modules`, `vendor`, `__pycache__`, `venv`, `target`, `dist`, and nested monom projects are never scanned. To hide more, add `discover.hide = <pattern>` lines to the `monom` file (or print them from a hook script's `config` hook). A pattern without a `/` matches a name at any depth; one with a `/` matches a path from the root. Both use `path.Match` globs. A declared path is registered even when a pattern hides it. `mnmd discover` prints the registered commands.

**How the root is found**, first match wins: `$_MONOM_PROJECT_ROOT` (what an alias pins), the nearest directory upward containing a file named `monom`, the nearest git root. The current directory alone is never a root.

**The `monom` file** is optional and comes in two forms. An executable script with a shebang implements [hooks](architecture.md#hooks) (`complete`, `run`, `debug`, `config`) to customize anything. Any other file is declarative and is never executed:

```
# register an executable the gate would skip
tools/Build.EXE
# make mnmd check fail on unregistered executables (default: warning)
check.root-contents = error
# treat more entries as hidden, like dot-files (one pattern per line)
discover.hide = *.TXT
discover.hide = tools/wip-*
```

**Checking a project.** `mnmd check` validates the registered commands. Problems in the root's contents — executables that were found but not registered, for example — are **warnings** by default and do not fail the run. Every other problem is an error and exits 1. To make root-contents problems errors, set `check.root-contents = error` in a declarative `monom` file, or print it from a hook script's `config` hook. Set `MONOM_CHECK_ROOT_CONTENTS=error` to make that your default everywhere. A project's own setting wins over it.

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
     │ _monom_complete     │    │ _monom_cfg run  │   ← hook or mnmd discover
     │ mnmd filter <pfx>   │    │ mnmd pack       │   ← Go binary
     └─────────────────────┘    └───────┬────────┘
                                        │
                                  exec resolved path
```

There are two roles:

- **CLI Author** — writes the config file and command scripts. Works against monom's interface.
- **CLI User** — types commands and hits Tab. Never knows monom exists.

The optional config file (`monom` at the project root) is the seam between monom's engine and the author's project. As a hook script it can expose `complete` (list all command paths, replacing default discovery) and `run` (rewrite args before path resolution), among other hooks. Without it, monom's default discovery (`mnmd discover`) supplies the command tree.

## Project Status

monom is in **early development**. This is the second attempt at building it — the first stalled out. The binding constraint is very little available time, so the approach leans heavily on AI-assisted development.

An implementation exists: a working Go binary (`mnmd`), shell bindings for bash and zsh, tab completion, and an end-to-end test harness. Active work is focused on solidifying the binary's subcommands and refining the shell integration. `BACKLOG.md` lists what's intended next.

## How the project is documented

Deliberately lightly, and in four places:

- **The code and its tests** describe all behavior. Nothing exists only as prose about what a function does.
- **`constitution.md`, `architecture.md`, `terminology.md`** hold the high-level decisions — invariants, the shape of the system, and names. These answer questions no single file can.
- **`TRADEOFFS.md`, beside the code** holds decisions that were genuinely contested: what was chosen, what was rejected, why, and what it costs. Read the one next to a package before changing that package's behavior.
- **`BACKLOG.md`** holds intended work as short statements of intent.

There are no specification documents and no change proposals. The rule is in the constitution under *Documentation Is Load-Bearing or Absent*: if a competent reader could derive it from the code, it doesn't get written.
