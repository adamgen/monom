# Tradeoffs — default discovery

Decisions here were contested. The code shows *what* `Discover` does; this file records *why* the losing option lost.

---

## A broad scan narrowed by a gate, not "every executable is a command"

**Chosen:** scan every executable, then register only those with a shebang, a command-shaped name (`^[a-z0-9][a-z0-9_-]*$`), or an explicit declaration.

**Rejected:** registering every executable file the scan finds, as the demo project's hand-written `complete` hook does with `find -perm -u+x`.

**Why.** The execute bit is an unreliable signal in real repositories: files checked out from Windows or FAT volumes, `chmod -R +x` accidents, shared libraries (`libfoo.so.1`), and build artifacts all carry it. Registering them makes completion noisy and trains users to ignore it. A shebang is a deliberate statement that a file is meant to be run; a lowercase extensionless name is the shape every command in the file-tree-as-command-tree model already has, and it is what admits compiled binaries, which have no shebang. Anything else can be declared.

Scanning broadly anyway (rather than only looking for shebangs) is what lets `mnmd check` tell the author which executables were left out and why.

**What it costs.** A shebang-less executable with an extension or capital letters (`Deploy`, `tool.exe`) is not a command until the author renames or declares it. `mnmd check` reports each one, but only as a warning by default.

---

## Default discovery walks the tree on every Tab, uncached

**Chosen:** `mnmd discover` scans the root each time it runs.

**Rejected:** caching the registered set (in a file or an environment variable) and invalidating it on change.

**Why.** Invalidation needs either a filesystem watcher or a tree walk to compare mtimes — and the walk is most of the cost of discovery itself. A stale cache shows the user commands that no longer exist, or hides the one they just wrote, which is exactly the moment they reach for Tab. The noise rules (`node_modules`, `vendor`, hidden directories, nested projects) exclude the trees that make walks expensive.

**What it costs.** Latency grows with the size of the tree. For a very large repository, the fix is a `complete` hook, which replaces default discovery entirely.

---

## `pack` does not enforce the gate

**Chosen:** `mnmd pack` resolves any executable path the user types, registered or not.

**Rejected:** refusing to run executables that default discovery would not register.

**Why.** Pack cannot know whether the gate applies without running the `complete` hook to learn whether default discovery is even in effect — a subprocess on every execution, which `internal/pack/TRADEOFFS.md` already rejected for the same reason. Custom projects rely on pack running binaries the gate would reject. The gate decides what the CLI *offers*: completion, group listings, and `mnmd check`.

**What it costs.** In a zero-config project, typing the exact path of an unregistered executable (`monom notes.TXT`) still runs it. Nothing advertises it, so it is only reachable deliberately.

---

## Declarations win over `discover.hide`

**Chosen:** a path declared in the monom config is registered even when a `discover.hide` pattern matches it (or a directory above it).

**Rejected:** letting a hide pattern veto a declaration.

**Why.** `discover.hide` extends the scan's noise rules, and declarations already bypass every noise rule: `_lib/helper` and `node_modules/x/cli` can be declared. A hide pattern that also beat declarations would be the only noise rule that did, and it would make `discover.hide = tools/wip-*` plus `tools/wip-keep` a silent contradiction instead of a useful idiom: hide a family, keep one. A declaration names one exact path the author typed on purpose; a glob is broad and approximate, so the specific statement wins.

**What it costs.** Hiding cannot be used to switch off a declaration; the author deletes the declaration line instead.

---

## `mnmd discover` runs the `config` hook

**Chosen:** in a hook-script project, `mnmd discover` runs the `config` hook to read `discover.hide`, ignoring invalid lines and a failing hook.

**Rejected:** honoring `discover.hide` only in declarative files, or adding a separate hook for it.

**Why.** A hook script's settings live in its `config` hook, and `mnmd check` already reads them there. A setting that `check` honored but discovery ignored would make the two disagree about the registered set. The shell calls `mnmd discover` only when the `complete` hook printed nothing, so the extra subprocess is paid only by hook projects that rely on default discovery; declarative and zero-config projects pay nothing. Errors are swallowed because discovery runs on Tab, which must never be noisy; `mnmd check` reports them.

**What it costs.** One more process per Tab in a hook script without a `complete` hook. And a `config` hook must not call `mnmd discover`, or the two would recurse.

---

## Two scripts sharing a stem keep their suffixes instead of picking a winner

**Chosen:** when two scripts in one directory share a stem (`build.sh` and `build.py`), both are registered under their full names, `mnmd pack build` is an error naming both, and `mnmd check` reports a `command-names` error.

**Rejected:** a precedence rule that strips one of them, such as preferring `.sh`, or the first in sort order, so `monom build` still runs something.

**Why.** It is the time bomb that `internal/pack/TRADEOFFS.md` turns down for command groups. If `build` resolved to `build.py` by some rule, then adding `build.sh` beside it, or renaming either file, would silently change what `monom build` runs, with no error and no diff that looks related. Showing both full names surfaces the clash to the author in completion, and the check error makes CI catch it.

**What it costs.** Adding a second script with the same stem renames the first command: `monom build` stops working and has to become `monom build.sh`. It fails loudly, but it does break scripts and muscle memory that used the short name until the author renames one.

---

## An existing entry owns its name over a script with that stem

**Chosen:** a script whose stem already names an entry in its directory is not registered: `db/` beside `db.sh` offers only `db` (the group), and `release` beside `release.sh` offers only `release`. `mnmd check` reports the script under `command-names`; it still runs by its full name.

**Rejected:** registering the script under its full name next to the entry, so completion offers `db` and `db.sh` side by side.

**Why.** The entry already resolves the name: pack tries the exact path first, so `monom db` is the group whatever the script is. Offering `db.sh` beside it puts a file suffix back into the CLI's surface, which stripping exists to remove, as a second command that reads like a sibling of the group. Dropping it keeps the surface clean, and the check error tells the author instead of letting the script disappear quietly.

**What it costs.** The script is reachable only by typing its full name, and `mnmd check` fails until the author renames one of the two. A project that keeps a `release` binary next to its `release.sh` build script has to rename one, or turn stripping off.

---

## Only scripts lose their suffix

**Chosen:** suffix stripping, and pack's stem resolution, apply only to scripts: executable files with a shebang. A declared file without a shebang keeps its name (`tools/Build.EXE`), and other executables (a stray `deploy.log`, a `pkg.tar.gz`) never compete for a stem.

**Rejected:** stripping every registered file, declarations included, and counting every executable toward a stem's ambiguity.

**Why.** Pack must resolve exactly the names discovery registers, and it cannot read the config (see `internal/pack/TRADEOFFS.md`), so it cannot know what is declared. A shebang can be checked from the file alone. A suffixed name never matches the naming pattern, so a script is the only suffixed file the scan registers anyway. Counting every executable would let a file that is not a command make a real one ambiguous: `chmod -R +x` would put a `deploy.log` next to `deploy.sh`, and `mnmd check` would fail on a command-names error for a file that was never a command.

**What it costs.** A declared file keeps its suffix in its command name. Adding a shebang strips it, which renames the command.
