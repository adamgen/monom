# Installing monom — a guide for agents

This is a procedure for a coding agent asked to install monom on a user's machine. Follow it in order. Every step ends with a check; do not move on until the check passes.

When a step says **ask the user**, stop and ask. Don't guess. Everything else you can do on your own.

---

## What you are setting up

monom is a compiled binary next to the shell files that load it. Installing it means three things:

1. An install tree lives somewhere permanent: `<dir>/bin/mnmd` and `<dir>/src/monom` (plus `monom.bash`, `monom.zsh`). `install.sh` puts it in `~/.local/share/monom`.
2. `mnmd` in that tree is a prebuilt release binary, or one built from source when there's no release for the platform.
3. The user's shell rc file sources `<dir>/src/monom` on startup. `mnmd install` writes that line.

The rc line is an absolute path into the tree. The binary finds `src/monom` through its own real location (`<dir>/bin/mnmd` → `<dir>/src/monom`), so symlinks to it are fine (`install.sh` links `~/.local/bin/mnmd`). If the tree moves, or the binary is copied out of it, the install breaks. See `internal/install/TRADEOFFS.md` for why.

---

## 1. Check prerequisites

```sh
uname -sm           # Linux or Darwin; x86_64/amd64 or arm64/aarch64 have prebuilt binaries
command -v curl || command -v wget
echo "$SHELL"
go version          # only needed if there's no prebuilt binary for this platform
```

- **No curl and no wget:** ask the user how they want to proceed.
- **Other OS or architecture:** `install.sh` builds from source, which needs Go (at least the version in `go.mod`, currently 1.24). If Go is missing, ask the user how they want it installed (Homebrew, their package manager, or go.dev/dl). Don't pick one for them.
- **Windows:** not supported outside WSL.

## 2. Confirm the user's shell

monom supports **bash and zsh only**. `mnmd install` picks the rc file based on `$SHELL`, and exits 1 for any other shell.

Your own tool shell may not be the user's login shell (sandboxes, CI images, and IDE terminals often differ). If you're not sure the `$SHELL` you see is the one the user types into, **ask the user**.

- On **fish, nushell, or anything other than bash or zsh**: stop. Tell the user monom doesn't support their shell.

## 3. Run the installer, without touching rc files yet

If the user wants monom somewhere other than `~/.local/share/monom`, set `MONOM_INSTALL_DIR`. Never use a temp directory, a scratch directory, or your own workspace. The rc line will point here for good.

```sh
curl -fsSL https://raw.githubusercontent.com/adamgen/monom/main/install.sh | MONOM_NO_MODIFY_RC=1 bash
MONOM_HOME="$HOME/.local/share/monom"      # or the MONOM_INSTALL_DIR you chose
```

It prints what it downloaded, that the SHA-256 matched, and `installed mnmd <version> to <dir>`. If an rc file already sources an earlier monom install, it upgrades that directory instead and says so. Use that directory as `MONOM_HOME`. If the user already uses a monom git checkout, the installer stops and says so: update the checkout with `git pull && ./build.sh` and skip to step 5.

If it stops with `no monom-<os>-<arch>.tar.gz ... and Go is not installed`, go back to step 1.

**Check:** `"$MONOM_HOME/bin/mnmd" version` prints a version.

## 4. Write the rc line

Tell the user which rc file this will change before you run it, unless they've already told you to go ahead:

```sh
"$MONOM_HOME/bin/mnmd" install
```

Run the binary from the install tree (or through a symlink to it, like `~/.local/bin/mnmd`). Don't copy it anywhere first. A copied binary writes a `source` line to a `src/monom` that doesn't exist, and it still reports success.

(Running the one-liner without `MONOM_NO_MODIFY_RC=1` does steps 3 and 4 together.)

Expected output is either:

```
added to /Users/<user>/.zshrc
restart your shell or run: source /Users/<user>/.zshrc
```

or `already installed`, which is also a success.

Write down the rc file path it printed. You need it in step 5.

## 5. Verify

Your tool shell won't pick up the change, because each command you run is a fresh non-interactive process. Verify in a new **interactive** shell, which is what the user's terminal starts.

**Check the line was written:**

```sh
grep -n 'src/monom' <rc file from step 4>
```

**Check the shell loads it.** Use the command that matches the rc file:

| rc file from step 4 | Command |
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

## 6. Hand off to the user

You can't press Tab, so the last check is theirs. Tell them to:

1. Open a new terminal, or run `source <rc file>` in the one they have open.
2. Run `cd` into any git repository with scripts in it, type `monom ` and press Tab. They should see its top-level folders and commands.

If they plan to build their own CLI next, a project needs no config at all: any git repository, or a directory with an empty `monom` file, works. [`fixtures/zero-config-project`](https://github.com/adamgen/monom/tree/main/fixtures/zero-config-project) shows which executables default discovery registers. When they want to customize discovery or execution, point them to [`fixtures/demo-project/monom`](https://github.com/adamgen/monom/blob/main/fixtures/demo-project/monom), a complete minimal hook script. The interface it implements is described under *The User Config Interface* in `architecture.md`.

---

## Troubleshooting

| Symptom | Cause | What to do |
| --- | --- | --- |
| `hint: run 'mnmd install' to activate shell integration` on stderr | `mnmd` ran in a shell that hasn't sourced `src/monom`. Expected in your own tool shell. | Nothing, as long as step 5 passes. Set `MONOM_ACTIVE=1` to silence it in scripts. |
| `mnmd install: unsupported shell` (exit 1) | `$SHELL` isn't bash or zsh. | See step 2. If the user does use bash or zsh and `$SHELL` is wrong, run `SHELL=/bin/zsh "$MONOM_HOME/bin/mnmd" install` (or `/bin/bash`), after confirming with them. |
| zsh: `monom` runs, but Tab does nothing | The `source` line comes before `compinit` in `~/.zshrc`, so the completion was never registered. | Show the user the order. Offer to move the line below `compinit` (or below the framework line that calls it, e.g. `source $ZSH/oh-my-zsh.sh`). |
| Linux bash: works with `bash -lic` but not in a new terminal tab | Install wrote to `~/.bash_profile`, which login shells read. Most Linux terminals start non-login shells, which read `~/.bashrc`. | Check whether `~/.bash_profile` sources `~/.bashrc`. If it doesn't, ask the user whether to add the `source` line to `~/.bashrc` too. |
| Shell startup prints `no such file or directory: .../src/monom` | The install tree was moved or deleted after install, or install ran from a copied binary. | Remove the stale line from the rc file. Re-run step 3. |
| `monom: no project root found` | The current directory isn't inside a project: no `monom` file here or in any parent, no git repository, and no alias pinning `_MONOM_PROJECT_ROOT`. | Not an install problem. `cd` into a project, or `touch monom` at the directory that should be the root. |
| Something else | — | Set `MONOM_DEBUG_LOG=/tmp/monom.log`, reproduce the problem, and read the log. |

## Updating

Re-run the one-liner. It replaces `bin/mnmd` and `src/` in place and leaves the rc line alone. `MONOM_VERSION=<tag>` pins a release. For a git checkout: `git pull && ./build.sh`.

## Uninstalling

Remove the `source ".../src/monom"` line from the rc file, then delete the install tree (`~/.local/share/monom` by default) and the `~/.local/bin/mnmd` link. Ask the user before deleting anything.
