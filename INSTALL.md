# Installing monom — a guide for agents

This is a procedure for a coding agent asked to install monom on a user's machine. Follow it in order. Every step ends with a check; do not move on until the check passes.

When a step says **ask the user**, stop and ask. Don't guess. Everything else you can do on your own.

---

## What you are setting up

monom is a compiled binary next to the shell files that load it. Installing it means three things:

1. An install tree lives somewhere permanent: `<dir>/bin/mnmd` and `<dir>/src/monom` (plus `monom.bash`, `monom.zsh`). `install.sh` puts it in `~/.local/share/monom`.
2. `mnmd` in that tree is a prebuilt release binary. Where there's no release for the platform, the tree is a git checkout with `bin/mnmd` built by `./build.sh`.
3. The user's shell rc file sources `<dir>/src/monom` on startup. `mnmd install` writes that line.

The rc line is an absolute path into the tree. The binary finds `src/monom` through its own real location (`<dir>/bin/mnmd` → `<dir>/src/monom`), so symlinks to it are fine (`install.sh` links `~/.local/bin/mnmd`). If the tree moves, or the binary is copied out of it, the install breaks. See `internal/install/TRADEOFFS.md` for why.

---

## 1. Check prerequisites

```sh
uname -sm           # Linux or Darwin, x86_64/amd64 or arm64/aarch64: these have release tarballs
command -v curl
echo "$SHELL"
```

- **No curl:** ask the user how they want to proceed.
- **Other OS or architecture:** there's no release tarball; use the checkout route in step 3. It needs Go (at least the version in `go.mod`, currently 1.24). If Go is missing, ask the user how they want it installed (Homebrew, their package manager, or go.dev/dl). Don't pick one for them.
- **Windows:** not supported outside WSL.

## 2. Confirm the user's shell

monom supports **bash and zsh only**. `mnmd install` picks the rc file based on `$SHELL`, and exits 1 for any other shell.

Your own tool shell may not be the user's login shell (sandboxes, CI images, and IDE terminals often differ). If you're not sure the `$SHELL` you see is the one the user types into, **ask the user**.

- On **fish, nushell, or anything other than bash or zsh**: stop. Tell the user monom doesn't support their shell.

## 3. Install

The installer adds a `source` line to the rc file `mnmd install` picks: `~/.zshrc` for zsh; `~/.bash_profile` if it exists, else `~/.bashrc`, for bash. Tell the user which file will change before you run it, unless they've already told you to go ahead.

```sh
curl -fsSL https://monom.dev/install.sh | bash
MONOM_HOME="$HOME/.local/share/monom"
```

If monom.dev is unreachable, the same script is at `https://raw.githubusercontent.com/adamgen/monom/main/install.sh`. It owns `~/.local/share/monom/{bin,src}` and replaces them on every run.

Expected output:

```
monom: downloading https://github.com/adamgen/monom/releases/latest/download/monom-darwin-arm64.tar.gz
monom: installed mnmd <version> in /Users/<user>/.local/share/monom
added to /Users/<user>/.zshrc
restart your shell or run: source /Users/<user>/.zshrc
```

or `already installed` in place of the last two lines, which is also a success. Write down the rc file path it printed. You need it in step 4 (Verify).

- **`no monom release for <os>/<arch>`:** build from a checkout instead (needs Go). The rc line will point into it, so **ask the user** where it should live if they haven't said, never a temp or scratch directory or your own workspace:

  ```sh
  MONOM_HOME="$HOME/.local/share/monom-src"   # or the location the user chose
  git clone https://github.com/adamgen/monom "$MONOM_HOME" && cd "$MONOM_HOME" && ./build.sh && bin/mnmd install
  ```

- **The user already has a monom checkout wired up:** don't install a second copy. Update it with `git pull && ./build.sh` there, and skip to step 4.
- **`mnmd install` failed** (for example `unsupported shell`): see step 2 and *Troubleshooting*. The files are installed; only the rc line is missing.

Don't copy `bin/mnmd` anywhere: a copied binary writes a `source` line to a `src/monom` that doesn't exist, and still reports success. Symlinks are fine.

**Check:** `"$MONOM_HOME/bin/mnmd" version` prints a version.

## 4. Verify

Your tool shell won't pick up the change, because each command you run is a fresh non-interactive process. Verify in a new **interactive** shell, which is what the user's terminal starts.

**Check the line was written:**

```sh
grep -n 'src/monom' <rc file from step 3>
```

**Check the shell loads it.** Use the command that matches the rc file:

| rc file from step 3 | Command |
| --- | --- |
| `~/.zshrc` | `zsh -ic 'whence -w monom mnmd; echo "completion: ${_comps[monom]}"'` |
| `~/.bashrc` | `bash -ic 'type -t monom mnmd; complete -p monom'` |
| `~/.bash_profile` | `bash -lic 'type -t monom mnmd; complete -p monom'` |

Expected: `monom` and `mnmd` are both functions. The completion is `_monom` (zsh) or `complete -F _monom_completion monom` (bash). You may also see output from the user's own rc file, or warnings like `no job control in this shell`, because there's no terminal attached. Those are fine.

**Check that commands resolve and run** in a throwaway project. Use the same shell and flags as the row above:

```sh
demo="$(mktemp -d)" && touch "$demo/monom" && mkdir -p "$demo/infra" \
  && printf '#!/bin/sh\necho starting local environment...\n' > "$demo/infra/start" && chmod +x "$demo/infra/start"
bash -ic "cd '$demo' && monom infra; monom infra start; mnmd check"
```

Expected:

```
monom: 'infra' is a command group
available: start
starting local environment...
✔ 1 commands OK (default discovery)
```

`mnmd check` finds the project root and config file itself, so it also works on its own in a fresh shell.

## 5. Hand off to the user

You can't press Tab, so the last check is theirs. Tell them to:

1. Open a new terminal, or run `source <rc file>` in the one they have open.
2. Run `cd` into any git repository with scripts in it, type `monom ` and press Tab. They should see its top-level folders and commands.

If they plan to build their own CLI next, a project needs no config at all: any git repository, or a directory with an empty `monom` file, works. [`fixtures/zero-config-project`](https://github.com/adamgen/monom/tree/main/fixtures/zero-config-project) shows which executables default discovery registers. When they want to customize discovery or execution, point them to [`fixtures/demo-project/monom`](https://github.com/adamgen/monom/blob/main/fixtures/demo-project/monom), a complete minimal hook script. The interface it implements is described under *The User Config Interface* in `architecture.md`.

---

## Troubleshooting

| Symptom | Cause | What to do |
| --- | --- | --- |
| `hint: run 'mnmd install' to activate shell integration` on stderr | `mnmd` ran in a shell that hasn't sourced `src/monom`. Expected in your own tool shell. | Nothing, as long as step 4 passes. Set `MONOM_ACTIVE=1` to silence it in scripts. |
| `mnmd install: unsupported shell` (exit 1) | `$SHELL` isn't bash or zsh. | See step 2. If the user does use bash or zsh and `$SHELL` is wrong, run `SHELL=/bin/zsh "$MONOM_HOME/bin/mnmd" install` (or `/bin/bash`), after confirming with them. |
| zsh: `monom` runs, but Tab does nothing | The `source` line comes before `compinit` in `~/.zshrc`, so the completion was never registered. | Show the user the order. Offer to move the line below `compinit` (or below the framework line that calls it, e.g. `source $ZSH/oh-my-zsh.sh`). |
| Linux bash: works with `bash -lic` but not in a new terminal tab | Install wrote to `~/.bash_profile`, which login shells read. Most Linux terminals start non-login shells, which read `~/.bashrc`. | Check whether `~/.bash_profile` sources `~/.bashrc`. If it doesn't, ask the user whether to add the `source` line to `~/.bashrc` too. |
| Shell startup prints `no such file or directory: .../src/monom` | The install tree was moved or deleted after install, or install ran from a copied binary. | Remove the stale line from the rc file. Re-run step 3. |
| `monom: no project root found` | The current directory isn't inside a project: no `monom` file here or in any parent, no git repository, and no alias pinning `_MONOM_PROJECT_ROOT`. | Not an install problem. `cd` into a project, or `touch monom` at the directory that should be the root. |
| Something else | — | Set `MONOM_DEBUG_LOG=/tmp/monom.log`, reproduce the problem, and read the log. |

## Updating

Re-run the one-liner; it replaces `bin/` and `src/` and leaves the rc line alone. `MONOM_VERSION=v1.2.3` pins a release. For a checkout: `git pull && ./build.sh`.

## Uninstalling

Remove the `source ".../src/monom"` line from the rc file, then delete `~/.local/share/monom` (or the checkout) and the `~/.local/bin/mnmd` link. Ask the user before deleting anything.
