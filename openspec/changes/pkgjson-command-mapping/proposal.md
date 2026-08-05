## Why

In monorepo projects, a JS package already outlines its subcommands as `scripts` in its `package.json`. Today a CLI author who wants `monom web build` to run that package's `build` script has no path: `resolve-*` mapping values must be executable file paths, and a package.json script is not a file — it runs as `npm run <script>` in the package directory. monom should read the manifest the author already maintains and expose its scripts as first-class monom commands (discovery, completion, and run) with zero duplication — and it must compose cleanly with the legacy-migration mappings the same config already declares via `resolve-run`/`resolve-complete`.

## What Changes

- Extend the **command mapping** table consumed by `mnmd resolve-run` and `mnmd resolve-complete` with a directive entry kind:
  - `<value> <key words...>` — file mapping, exactly as today (unchanged).
  - `@pkgjson <package.json-path> [prefix]` — manifest mapping: expands in place to one entry per script in that manifest, keyed `<prefix> <script>` (or just `<script>` with no prefix).
  - One table drives both hooks in one `mnmd` call each; entry order — across both kinds — is match priority, so legacy-name vs package-script precedence is the author's line ordering. No new subcommands, no hook chaining.
- `mnmd resolve-complete` prints directive entries as `<prefix>/<script>` discovery paths, in manifest order, alongside the file-mapping keys.
- `mnmd resolve-run`, on a directive hit, prints an exec line that runs the script via the package's package manager (`npm` by default, detected from the `packageManager` field; pnpm/yarn/bun recognized); misses stay passthrough, exactly as today.
- Extend `mnmd pack` to carry arguments: when the **first** token is an absolute path, the remaining tokens are the command's arguments — pack validates the first token as today and prints all tokens **one per line** (single-command output is byte-identical to today, so this is backward compatible). This is what lets a `run` hook resolve to "executable + args" instead of only a bare file.
- Extend the shell execution core (`_monom_run` in `src/monom`): split pack's output on newlines and `exec` the result as an argv array instead of a single quoted path.
- Extend `fixtures/legacy-migration/` with a JS package and an `@pkgjson` line in its existing table, demonstrating legacy file mappings and manifest scripts coexisting in one config.
- Update `architecture.md` (mapping-table entry kinds, pack contract, exec flow) and `terminology.md` (extend **Command mapping** with the manifest-driven directive).

Not in scope: passing user-typed arguments through to scripts (blocked on the future `mnmd args` design — commands receive no args anywhere in monom today), running `npm install`/lifecycle management, and workspace auto-discovery (globbing `workspaces` patterns to find package.json files automatically — a natural follow-up once this contract exists).

No **BREAKING** changes: existing mapping tables parse identically (`@pkgjson` was previously an invalid value line), and the required user-config interface (`complete`) is untouched, so no constitution amendment is needed; hooks and pack evolve in `architecture.md` as allowed.

## Capabilities

### New Capabilities
- `resolve-subcommands`: the `mnmd resolve-run` / `mnmd resolve-complete` pair — mapping-table format (file entries and `@pkgjson` directive entries), hook-contract semantics (passthrough miss, discovery output, entry-order precedence), manifest parsing, package-manager detection, and error posture. The pair shipped without a spec; this capability back-fills the existing behavior and adds the directive.

### Modified Capabilities
- `pack-subcommand`: pack accepts an absolute first token followed by argument tokens; output becomes newline-delimited tokens (first line = executable path, subsequent lines = args). Single-token resolution output is unchanged.
- `shell-binding-core`: `_monom_run` execs pack's output as an argv array (newline-split) instead of a single path.

## Impact

- **New Go package**: `internal/pkgjson` (manifest reading, script-order preservation, package-manager detection) + unit tests — a helper consumed by `internal/resolve`.
- **Modified Go**: `internal/resolve/resolve.go` (directive parsing and expansion), `internal/pack/pack.go` (absolute-first-token args passthrough). `cmd/mnmd/main.go` is untouched apart from doc comments — dispatch gains nothing.
- **Modified shell**: `src/monom` (`_monom_run` exec path). No new logic in shell — only the technically-unavoidable exec surface changes.
- **Tests**: Go unit tests for `internal/pkgjson`, extended `internal/resolve` and `internal/pack` tests; extended shUnit2 e2e suites `tests/mnmd_resolve_run_test`, `tests/mnmd_resolve_complete_test`, `tests/mnmd_pack_test`, `tests/monom_run_test`.
- **Fixtures**: `fixtures/legacy-migration/` gains a JS package and a mixed table.
- **Docs**: `architecture.md`, `terminology.md`.
- **Runtime deps**: none added — package manager binaries are resolved from `$PATH` at run time and only when a mapped script is actually invoked.
