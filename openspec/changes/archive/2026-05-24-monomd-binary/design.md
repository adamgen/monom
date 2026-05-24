## Context

monom's architecture document specifies a compiled Go binary (`monomd`) as the engine for all internal logic, with shell files as thin wrappers for what is technically impossible in Go. The prototype codebase has been archived to `_archive/`, leaving the repo root clean. This change writes the `monomd` binary from scratch into that clean repo.

This change also includes a constitution amendment that restructures the user config interface: `complete` becomes the only required subcommand, and `run` becomes a documented optional hook. The amendment introduces a named "Pluggability via Hooks" principle and reframes the existing interface-amendment principle to apply only to the *required* surface, so future optional hooks can be added without amendment.

## Goals / Non-Goals

**Goals:**
- A working `monomd` Go binary with four subcommands: `filter`, `pack`, `root`, `check`
- Go unit tests covering each subcommand's logic in isolation
- shUnit2 e2e tests covering the binary's invocation surface (stdin, args, stdout, exit codes)

**Non-Goals:**
- `monomd args` subcommand (marked TBD in architecture.md, out of scope)
- Shell files (`src/monom`, `src/monom.bash`) — separate change
- Implementation of the `run` hook lookup in the shell `monom()` function — separate change
- Build infrastructure scripts (`build`, `check`, `sh_test_runner`, etc.) — separate change
- Test fixtures beyond what's needed for `monomd` e2e tests — separate change
- Adding new hooks beyond `run` to the architecture

## Decisions

### Decision: `monomd filter` is written from scratch

**Chosen:** Write `src/go_utils/filter.go` from scratch implementing `Filter(commands []string, prefix string) []string` cleanly.

**Rationale:** The prototype had a naming typo (`remvoe_prefix.go`), logic entangled with test fixture concerns, and function signatures that don't cleanly match the architecture. Writing fresh is faster and safer than auditing and salvaging the prototype.

---

### Decision: `monomd pack` takes args (not stdin) and discovers root internally

**Chosen:** `monomd pack <word...>` accepts space-separated tokens as CLI arguments. Internally, it discovers the project root using the same algorithm as `monomd root`, joins the args with `/`, and resolves to an absolute executable path.

**Rationale:**
- **Symmetry with filter:** Both `filter` and `pack` take space-separated user-typed tokens as CLI args. No reason for one to use stdin and the other args.
- **Self-sufficient:** Pack does not depend on any env vars being pre-set. The shell `monom()` function reduces to roughly `exec "$(monomd pack "$@")"`. No `setup_monom()` needed on the execution path.
- **Single source of truth for root discovery:** The same Go helper (`findProjectRoot`) backs both `monomd root` and `monomd pack`. One place to test, one set of edge cases.

**Alternatives considered:**
- *Pack reads stdin (space-separated path).* Rejected — inconsistent with filter's args interface; requires unnecessary pipe machinery in the shell.
- *Shell sets `$MONOM_PROJECT_ROOT` before calling pack.* Rejected — pushes coordination logic into shell that should live in Go; couples pack to a shell setup step.

---

### Decision: `monomd root` honors `$MONOM_PROJECT_ROOT` when valid, otherwise walks

**Chosen:** `monomd root` first checks `$MONOM_PROJECT_ROOT`. If set and pointing to a directory containing an executable `monom` file, return it. Otherwise, walk up from `$PWD` looking for one.

**Rationale:** Users can override discovery by setting the env var (useful for aliases, multi-project setups, testing). Without an override, discovery happens automatically. This is a single algorithm with two entry conditions, not two separate code paths.

**Rejected:** A separate `monomd root --honor-env` flag. Unnecessary — the env-aware behavior is always what callers want; there's no use case for "ignore the env var and force a walk."

---

### Decision: Eliminate `monom_cfg run` from the required execution path

**Chosen:** `monom_cfg run` is no longer required. The shell `monom()` function calls `monomd pack` directly with the user's args. If the user config exposes a `run` hook (now documented in `architecture.md` as an optional hook), the shell tries it and falls back to direct invocation on absence or failure.

**Rationale:**
- The only thing `monom_cfg run` did in the original design was print the user's args back (as a space-separated string for pack to consume). That's a pure pass-through — no value added by default.
- Authors who need to remap or transform args can still do so by exposing the `run` hook. The fallback ensures projects without `run` work seamlessly.
- Simplifies the required interface (one subcommand: `complete`).
- Simplifies the shell `monom()` function dramatically.

**Process:** This decision requires a constitution amendment because the required user config interface is constitution-protected. The amendment is included in this change (see Context).

**Alternatives considered:**
- *Keep `run` as required, pass-through by default.* Rejected — forces every project to implement a pass-through subcommand. Boilerplate with no value.
- *Move the `run` hook detection into `monomd pack` itself.* Rejected — couples pack to the user config; pack should be a pure path-resolution function.

---

### Decision: Remove shebang detection from shell; exec the resolved path directly

**Chosen:** `monomd pack` resolves to an absolute path. The shell `exec`s that path directly. The command file must be executable with a valid shebang. Shebang parsing does not exist in shell or Go.

**Rationale:** The architecture doc says shell "exec's the resolved command file." The file must be directly executable — a shebang + execute bit is the standard Unix contract. Shebang detection in shell is fragile and belongs nowhere.

**Alternative considered:** Embed shebang detection in `monomd pack`. Rejected — overcomplicates the pack contract.

---

### Decision: `monomd filter` receives raw words and builds the prefix internally

**Chosen:** The shell passes `$COMP_WORDS` (the raw space-separated tokens the user has typed) directly to `monomd filter` as arguments. `monomd filter` joins them with `/` internally to build the slash-delimited prefix and determine drill-down vs. partial-match behavior.

**Rationale:** Joining words with `/` and detecting a trailing empty word is logic — it belongs in Go, not shell. The shell's only job is to pass `$COMP_WORDS` through. This keeps all transformation decisions in Go where they are testable in isolation.

**Alternative considered:** Shell builds the slash-joined prefix and passes it as a single string. Rejected — that is logic in shell, which violates "Go owns logic, shell owns surface."

## Risks / Trade-offs

- **Risk: `monomd pack` validates execute permission** → files without execute bit fail at pack time, not run time. This is a clearer error than the silent-interpreter-guess behavior of the prototype. Mitigation: clear, actionable error message from `monomd pack`.

## Open Questions

- Should `monomd root` accept an optional `--from <dir>` flag to walk from a directory other than `$PWD`? Useful for testing. Deferred to implementation — the spec describes the default behavior.
