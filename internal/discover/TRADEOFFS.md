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
