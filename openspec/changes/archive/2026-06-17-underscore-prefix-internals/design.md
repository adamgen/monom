## Context

monom's shell entrypoint (`src/monom`) and completion bindings (`src/monom.zsh`, `src/monom.bash`) define functions and export environment variables, all named with a `monom`/`MONOM` stem. bash and zsh complete any identifier in the namespace by prefix, so `monom<Tab>` (no trailing space, before a project command is even started) offers the user a list polluted with monom's internals alongside the two real commands `monom` and `monomd`.

Current monom-defined identifiers and their status:

| Identifier | Kind | Today | Notes |
| --- | --- | --- | --- |
| `monom` | function | unprefixed | the command — keep (the only name left in the `monom` namespace) |
| `monomd` | binary | unprefixed | rename → `mnmd` (engine binary; machinery, not a user command) |
| `_monom` | zsh completion fn | already `_` | zsh `_<command>` convention |
| `_monom_log` | function | already `_` | |
| `_monom_self` | local var | already `_` | unset after use |
| `setup_monom` | function | unprefixed | rename |
| `monom_cfg` | function | unprefixed | rename |
| `monom_completion` | bash completion fn | unprefixed | rename |
| `MONOM_BIN` | env var | unprefixed | rename; internal |
| `MONOM_LIB_ROOT` | env var | unprefixed | rename; internal-ish |
| `MONOM_PROJECT_ROOT` | env var | unprefixed | rename; internal shell→Go convention (mis-documented as public) |
| `MONOM_USER_CONFIG` | env var | unprefixed | rename; internal shell→Go convention (mis-named as the seam in the constitution) |
| `MONOM_DEBUG_LOG` | env var | unprefixed | **keep** — user-exported opt-in input |

The rename spans several layers that must move together: the shell files that set the vars and resolve the binary, the Go binary that reads `MONOM_PROJECT_ROOT` (`internal/root/root.go`) and `MONOM_USER_CONFIG` (`internal/check/check.go`, `cmd/monomd/main.go`), the binary artifact itself (`cmd/monomd/` → `cmd/mnmd/`, `bin/monomd` → `bin/mnmd`, Makefile, `.gitignore`), the test files (`tests/monomd_*_test` → `tests/mnmd_*_test` and references), and the canonical docs (`constitution.md`, `architecture.md`, `terminology.md`, `CLAUDE.md`, `README.md`).

## Goals / Non-Goals

**Goals:**

- Remove every monom-defined identifier from the bare-`monom` completion candidate list, leaving only `monom` and `monomd`.
- Apply one consistent rule — prefix with a single leading `_` — so the convention is obvious and self-documenting.
- Correct the canonical docs so the public seam is described as the `monom` config file exposing `complete`, not as a named env var — keeping `constitution.md`, `architecture.md`, and `terminology.md` the single source of truth, updated in lockstep with the code.

**Non-Goals:**

- No behavioral change. The bindings, dispatch, and completion logic are byte-for-byte equivalent after the rename; only identifiers differ.
- Not introducing a namespacing scheme beyond a leading `_` (no `__monom_` double-underscore, no associative-array registry).
- Not renaming `MONOM_DEBUG_LOG` (see Decisions).
- Not changing the *required user config interface contract* — the `complete` subcommand on the `monom` config file is untouched. Only an internal env var name changes.
- No constitution amendment — see Decisions for why these vars were never part of the protected contract.

## Decisions

### Decision: Prefix with a single leading underscore

Use `_` as the sole prefix (`_setup_monom`, `_MONOM_BIN`, …). A leading underscore is the established Unix convention for "internal/private" shell identifiers, it is already used by `_monom` and `_monom_log`, and — critically — `monom<Tab>` will no longer surface them because they no longer start with `monom`.

- **Alternative — double underscore `__monom_`:** more visually distinct but heavier, inconsistent with the existing `_monom*` names, and unnecessary since a single `_` already solves the completion problem.
- **Alternative — leave these env vars unprefixed:** rejected per the user's requirement that nothing beginning with `monom`/`MONOM` (other than the real command) appear in completion. Uppercase env vars are completed by the shell too.

### Decision: Rename the engine binary `monomd` → `mnmd` (not a `_` prefix)

The shell functions/vars get a `_` prefix because they live in the session namespace and can be made invisible to `monom<Tab>`. The engine is a **file on disk** that may be on the user's PATH — a filename cannot be hidden from completion the way a shell identifier can. The achievable and sufficient goal is to take it out of the `monom` prefix so `monom<Tab>` shows exactly one command (`monom`). `mnmd` does that; it reads as a terse, daemon-style engine name (`dockerd`, `containerd`). Subcommands and all behavior are unchanged.

This also resolves a live contradiction: `architecture.md`'s Entry Points section already classifies the binary as *machinery* reached through the three entry points, yet names it with the `monom` prefix and (via this change's own earlier draft note) called it one of "two real commands." After the rename there is one real command — `monom` — and the engine is unambiguously machinery.

- **Alternative — `_monomd`:** rejected. Odd to type for a runnable binary, and a PATH binary named `_monomd` is no more hidden than `mnmd` — it just looks like a mistake.
- **Alternative — `mnomd`:** acceptable but a character longer for no benefit; `mnmd` chosen.
- **Alternative — keep `monomd`:** rejected; it is the contradiction the user identified and leaves the binary under `monom<Tab>`.

### Decision: Treat `MONOM_PROJECT_ROOT` / `MONOM_USER_CONFIG` as implementation details, not public API

These two vars were documented in `architecture.md` and `MONOM_USER_CONFIG` was named in `constitution.md` as if they were the interface. They are not. The public seam monom promises the CLI author is the **executable `monom` config file at the project root, exposing `complete`** (plus optional hooks). `MONOM_PROJECT_ROOT` and `MONOM_USER_CONFIG` are an internal calling convention: monom's shell layer resolves them and hands them to the `monomd` binary. The CLI author never sets or reads them. (The lone public-facing behavior — a user MAY pre-set the project root to skip auto-discovery — is a *behavior* monom honors, not a named-API guarantee about the variable's spelling.)

Consequences:

1. **No constitution amendment.** The constitution is reworded to state its stability contract in terms of the config file and its required `complete` subcommand, naming no env var. Renaming a private variable does not touch the contract, so no amendment is triggered.
2. `terminology.md` defines "monom config file" and "project root" by concept; the env var is demoted to an internal-mechanism note.
3. `architecture.md` keeps the Environment Variables table (architecture is the right home for internal mechanism) but introduces it as internal shell↔Go plumbing, with the pre-set-project-root behavior flagged as the one public affordance.

The required *interface* (the `complete` subcommand) is unchanged — only the spelling of an internal pointer changes.

- **Alternative — keep treating them as public and amend the constitution:** rejected. It enshrines an internal variable name as a stability guarantee, which is exactly the leak this change corrects, and forces every future rename through a constitution amendment.
- **Alternative — keep these two unprefixed:** rejected by the user; also leaves the exact completion noise the change exists to remove.

### Decision: Keep `MONOM_DEBUG_LOG` unprefixed

`MONOM_DEBUG_LOG` is the one `MONOM_*` name monom does *not* define — the **user** exports it to opt into debug logging; monom only reads it. It is an input flag, semantically closer to a CLI flag than to an internal. It only appears in a user's `monom<Tab>` completion if that user already exported it, so it does not contribute to the default noise. Renaming it to `_MONOM_DEBUG_LOG` would make a user-facing toggle look private and would be an odd thing to ask users to type. It stays as-is; this exception is called out in `architecture.md`/`debug-log` spec text rather than left implicit.

### Decision: Single atomic rename across shell, Go, docs, and tests

Because the env var names form a contract between the shell (writer) and Go (reader), they must change together or the binary breaks. The implementation does the full rename in one change rather than staging it, and validates with `go vet`, `go test ./...`, the shUnit2 suites, and `shellcheck` before completion.

## Risks / Trade-offs

- **A reader/writer pair drifts (shell sets `_MONOM_USER_CONFIG`, Go still reads `MONOM_USER_CONFIG`)** → Enumerate every `os.Getenv`/export site up front (already located: `root.go`, `check.go`, `main.go`, all three shell files) and cover with the existing Go + shUnit2 + completion tests, which assert end-to-end behavior across the boundary.
- **A stale reference to an old name lingers in docs or a comment** → After the mechanical rename, grep the whole tree for `\bMONOM_(BIN|LIB_ROOT|PROJECT_ROOT|USER_CONFIG)\b` and the bare function names, excluding `MONOM_DEBUG_LOG`, and confirm only intentional historical mentions (e.g. archived changes) remain.
- **Someone was relying on the old env var names despite them being internal** → These were never a documented public API the CLI author writes against (they're set *by* monom, not by the author), so there is no supported contract to break. The one public behavior — pre-setting the project root — is preserved (the variable is simply spelled `_MONOM_PROJECT_ROOT`), and that affordance is documented. No compatibility shim is added, to avoid reintroducing a `MONOM_*` name into the namespace.
- **Archived openspec changes still reference old names** → Left untouched on purpose; archives are a historical record, not live contract.
