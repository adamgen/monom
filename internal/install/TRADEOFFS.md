# Tradeoffs — `mnmd install` and the activation nudge

Decisions here were contested. The code shows *what* install does; this file records *why* the losing option lost.

---

## The nudge is unconditional, and it goes to stderr

**Chosen:** every subcommand except `install` prints `hint: run 'mnmd install' to activate shell integration` to stderr when `MONOM_ACTIVE` is unset.

**Rejected:** printing it only for interactive terminals (`isatty`), or once per day via a state file.

**Why.** The failure this catches is a user who built the binary, ran `mnmd filter`, saw plausible output, and concluded monom is installed — when in fact no shell is sourcing `src/monom` and Tab does nothing. The hint has to fire in exactly the situation where the user is poking at the binary directly, which is the situation a tty check would also flag, but a tty check adds a way to be wrong (piped-but-interactive sessions) for no gain. A state file adds a file.

Stderr is what makes the unconditional version tolerable: `mnmd pack a b` still emits nothing but the resolved path on stdout, so `$(mnmd pack ...)` in the shell binding is unaffected. The nudge is visible to a human and invisible to a pipe.

**What it costs.** Scripts that capture combined output (`2>&1`) from `mnmd` without setting `MONOM_ACTIVE` will see the hint interleaved. Accepted — `MONOM_ACTIVE=1` is a one-word opt-out for anyone who means it.

---

## `src/monom` is located from the binary, not from `$PATH` or a build-time constant

**Chosen:** `os.Executable()` → `filepath.EvalSymlinks` → `../src/monom`.

**Rejected:** baking the install prefix in at build time with `-ldflags`, or searching `$PATH` for a sibling `src/`.

**Why.** The symlink resolution is the whole point. A user who installs via Homebrew, or via `ln -s ~/dev/monom/bin/mnmd ~/bin/mnmd`, gets an `os.Executable()` that points at the symlink; joining `../src/monom` to *that* yields a path in `~/` that does not exist. Resolving to the real binary first makes the layout assumption — binary at `<root>/bin/mnmd`, sources at `<root>/src/` — hold under every installation method that preserves the tree.

A build-time constant would be more robust still, but it requires the build to know its final install location, which is false for `make build` followed by `mnmd install` — the primary path today.

**What it costs.** The `<root>/bin` + `<root>/src` layout is now load-bearing and cannot change without changing this function. Relocating either directory breaks install silently for anyone using a symlink.

---

## install.sh downloads a release tarball; it doesn't clone and build

**Chosen:** `install.sh` fetches a prebuilt `monom-<os>-<arch>.tar.gz` from the GitHub release, verifies its SHA-256, and unpacks the `bin/` + `src/` tree into `~/.local/share/monom`. It builds from source (with `build.sh`, never make) only when the release has no tarball for the platform and Go is present.

**Rejected:** cloning the repo and running `make build`, which was the documented install before.

**Why.** A clone-and-build install requires git, make and a Go toolchain on every machine that wants tab completion, and it failed halfway for anyone without Go: the clone was left behind, no binary, no rc line. The release tarball needs only curl (or wget) and tar, and the checksum makes a truncated or tampered download fail loudly instead of installing. Asset names carry no version, so `releases/latest/download/<name>` works without an API call or `jq`.

**What it costs.** The one-liner only installs a prebuilt binary once a release exists, so releases now have to be cut (push a `v*` tag). An install tree is not a checkout: it has no fixtures or tests, and updating means re-running the installer rather than `git pull`.

---

## install.sh delegates the rc line to `mnmd install`

**Chosen:** after placing the tree, `install.sh` runs `<dir>/bin/mnmd install`. It adds one check of its own: an rc line that already sources the same directory under another spelling (`~`, `$HOME`, a symlinked path) counts as installed, because `mnmd install` only recognises the exact path it writes.

**Rejected:** reimplementing shell detection and rc editing in the shell script.

**Why.** Two implementations of "which rc file, and is the line there" would drift. The Go one is unit-tested and is what users of a checkout run anyway.

**What it costs.** The installer depends on the binary it just downloaded running on the machine, which it checks first (`mnmd version`).

---

## Idempotency is a substring match, not a marker comment

**Chosen:** `alreadyInstalled` scans the rc file for any non-comment line containing the resolved `src/monom` path.

**Rejected:** writing a `# >>> monom >>>` … `# <<< monom <<<` block and detecting that, the way conda and rbenv do.

**Why.** A managed block is the right answer when the tool needs to *rewrite* or *remove* its own lines later. install only ever appends one line and never modifies it, so the block buys nothing and costs a parser plus a failure mode when a user edits inside the markers. Skipping commented lines means a user can disable the integration by commenting it out and a re-run will restore it — the behavior a comment implies.

**What it costs.** Two monom checkouts installed from different paths both append their own line, and both will be sourced. That is arguably correct (they are different installs), but it is not deduplicated and nothing warns about it.
