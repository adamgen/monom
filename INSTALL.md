# Installing monom — a guide for agents

This is a procedure for a coding agent asked to install monom on a user's machine. Follow it in order. Every step ends with a check; do not move on until the check passes.

When a step says **ask the user**, stop and ask. Don't guess. Everything else you can do on your own.

---

## What you are setting up

monom is a git checkout with a compiled binary inside it. Installing it means three things:

1. The checkout lives somewhere permanent.
2. `bin/mnmd` is built inside that checkout.
3. The user's shell rc file sources `<checkout>/src/monom` on startup. `mnmd install` writes that line.

The rc line is an absolute path into the checkout. The binary finds `src/monom` through its own real location (`<checkout>/bin/mnmd` → `<checkout>/src/monom`). So if the checkout moves, or the binary is copied out of it, the install breaks. See `internal/install/TRADEOFFS.md` for why.

---

## 1. Check prerequisites

```sh
git --version
go version          # must be at least the version in go.mod (currently 1.24)
echo "$SHELL"
uname -s            # Darwin or Linux
```

- **Go missing or too old:** ask the user how they want it installed (Homebrew, their package manager, or go.dev/dl). Don't pick one for them.
- **`make` is optional.** Step 3 has a plain `go build` fallback.

## 2. Confirm the user's shell

monom supports **bash and zsh only**. `mnmd install` picks the rc file based on `$SHELL`, and exits 1 for any other shell.

Your own tool shell may not be the user's login shell (sandboxes, CI images, and IDE terminals often differ). If you're not sure the `$SHELL` you see is the one the user types into, **ask the user**.

- On **fish, nushell, or anything other than bash or zsh**: stop. Tell the user monom doesn't support their shell.

## 3. Clone into a permanent location, then build

**Ask the user** where the checkout should live if they haven't said. A sensible default is `~/.local/share/monom`. Never use a temp directory, a scratch directory, or your own workspace. The rc line will point here for good.

```sh
MONOM_HOME="$HOME/.local/share/monom"     # or the location the user chose
git clone https://github.com/adamgen/monom.git "$MONOM_HOME"
cd "$MONOM_HOME"
make build                                 # or: mkdir -p bin && go build -o bin/mnmd ./cmd/mnmd
```

**Check:** `"$MONOM_HOME/bin/mnmd"` exists and is executable.

If the user already has a checkout, use it and build there. Don't clone a second copy: two checkouts means two `source` lines, and both get loaded.

## 4. Write the rc line

Tell the user which rc file this will change before you run it, unless they've already told you to go ahead:

```sh
"$MONOM_HOME/bin/mnmd" install
```

Run the binary from inside the checkout (or through a symlink to it). Don't copy it anywhere first. A copied binary writes a `source` line to a `src/monom` that doesn't exist, and it still reports success.

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

**Check that commands resolve and run,** using the bundled demo project. Use the same shell and flags as the row above. The double quotes are on purpose: your shell expands `$MONOM_HOME` before the new shell starts.

```sh
bash -ic "cd '$MONOM_HOME/fixtures/demo-project' && monom infra; monom infra local start; mnmd check"
```

Expected:

```
monom: 'infra' is a command group
available: cloud, local
starting local environment...
✔ 7 commands OK
```

`mnmd check` finds the project root and config file itself, so it also works on its own in a fresh shell.

## 6. Hand off to the user

You can't press Tab, so the last check is theirs. Tell them to:

1. Open a new terminal, or run `source <rc file>` in the one they have open.
2. Run `cd <MONOM_HOME>/fixtures/demo-project`, then type `monom ` and press Tab. They should see `db  infra  release`.

If they plan to build their own CLI next, a project needs no config at all: any git repository, or a directory with an empty `monom` file, works. `fixtures/zero-config-project` shows which executables default discovery registers. When they want to customize discovery or execution, point them to `fixtures/demo-project/monom`, a complete minimal hook script. The interface it implements is described under *The User Config Interface* in `architecture.md`.

---

## Troubleshooting

| Symptom | Cause | What to do |
| --- | --- | --- |
| `hint: run 'mnmd install' to activate shell integration` on stderr | `mnmd` ran in a shell that hasn't sourced `src/monom`. Expected in your own tool shell. | Nothing, as long as step 5 passes. Set `MONOM_ACTIVE=1` to silence it in scripts. |
| `mnmd install: unsupported shell` (exit 1) | `$SHELL` isn't bash or zsh. | See step 2. If the user does use bash or zsh and `$SHELL` is wrong, run `SHELL=/bin/zsh "$MONOM_HOME/bin/mnmd" install` (or `/bin/bash`), after confirming with them. |
| zsh: `monom` runs, but Tab does nothing | The `source` line comes before `compinit` in `~/.zshrc`, so the completion was never registered. | Show the user the order. Offer to move the line below `compinit` (or below the framework line that calls it, e.g. `source $ZSH/oh-my-zsh.sh`). |
| Linux bash: works with `bash -lic` but not in a new terminal tab | Install wrote to `~/.bash_profile`, which login shells read. Most Linux terminals start non-login shells, which read `~/.bashrc`. | Check whether `~/.bash_profile` sources `~/.bashrc`. If it doesn't, ask the user whether to add the `source` line to `~/.bashrc` too. |
| Shell startup prints `no such file or directory: .../src/monom` | The checkout was moved or deleted after install, or install ran from a copied binary. | Remove the stale line from the rc file. Re-run step 4 from the real checkout. |
| `monom: no project root found` | The current directory isn't inside a project: no `monom` file here or in any parent, no git repository, and no alias pinning `_MONOM_PROJECT_ROOT`. | Not an install problem. `cd` into a project, or `touch monom` at the directory that should be the root. |
| Something else | — | Set `MONOM_DEBUG_LOG=/tmp/monom.log`, reproduce the problem, and read the log. |

## Updating

```sh
cd "$MONOM_HOME" && git pull && make build
```

You don't need to run install again. The rc line points at the checkout, and the checkout didn't move.

## Uninstalling

Remove the `source ".../src/monom"` line from the rc file, then delete the checkout. Ask the user before deleting anything.
