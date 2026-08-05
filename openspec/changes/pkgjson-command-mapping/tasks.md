## 1. Pack args passthrough (prerequisite for the run half)

- [ ] 1.1 Extend `internal/pack/pack.go`: when the first token is absolute, validate only that token (exists, executable, not a directory) and return the full token list; keep relative-token joining unchanged. Add unit tests for absolute+args, absolute-not-executable, and the unchanged relative multi-token join.
- [ ] 1.2 Update `runPack` in `cmd/mnmd/main.go` to print the result as newline-delimited tokens (single-token output stays byte-identical to today).
- [ ] 1.3 Extend `tests/mnmd_pack_test` with e2e scenarios from the `pack-subcommand` delta spec: absolute single token as-is, absolute+args five-line output, non-executable absolute token error; assert existing single-token scenarios' output is unchanged.

## 2. Shell exec of argv

- [ ] 2.1 Update `_monom_run` in `src/monom`: split pack's stdout on newlines only (portable bash 3.2 / zsh idiom per design D3) and `( exec "${tokens[@]}" )`; verify shellcheck passes.
- [ ] 2.2 Extend `tests/monom_run_test`: multi-line pack output execs executable + args in order (both shells); single-line path containing spaces stays one argv element.

## 3. internal/pkgjson helper package

- [ ] 3.1 Create `internal/pkgjson` with manifest reading: parse `scripts` preserving file order via `json.Decoder` token walking; parse `packageManager` into a PM name. Unit-test order preservation, missing/invalid JSON, absent `scripts`.
- [ ] 3.2 Implement the PM invocation table (npm `--prefix`, pnpm `-C`, yarn `--cwd`, bun `--cwd`) with `exec.LookPath` resolution and `CodedError`s for unrecognized PM and PM-not-on-PATH. Unit tests stub `$PATH`.
- [ ] 3.3 Verify the exact cwd flags against real npm/pnpm/yarn/bun installs (design open question) and correct the PM table + spec if any differ.

## 4. @pkgjson directive in internal/resolve

- [ ] 4.1 Extend `parse` in `internal/resolve/resolve.go`: recognize `@pkgjson <path> [prefix]` lines as directive entries alongside file entries, preserving table order; resolve relative manifest paths against the project root via `internal/root`. Unit tests: mixed tables, comments/blanks, value-only skip unchanged.
- [ ] 4.2 Extend `Complete`: directive entries emit `<prefix>/<script>` lines in manifest order; skip script names containing whitespace or `/`; unreadable manifest → stderr warning, entry skipped, still exit 0; file entries stay filesystem-free. Unit tests per spec scenarios.
- [ ] 4.3 Extend `Run`: lazy shape-match for directives (manifest opened only when prefix words + one word match); hit → PM exec line via `internal/pkgjson`; first match wins across both entry kinds; loud `CodedError`s for unreadable shape-matched manifest and whitespace in package dir or script name; miss stays passthrough. Unit tests per spec scenarios, including file-entry-shadows-script precedence.
- [ ] 4.4 Confirm `cmd/mnmd/main.go` needs no dispatch change; update the `runResolveRun`/`runResolveComplete` doc comments to mention directive entries.

## 5. Fixture and e2e tests

- [ ] 5.1 Extend `fixtures/legacy-migration/`: add a JS package (e.g. `apps/web/` with a `package.json` defining ordered scripts, one manifest using `packageManager`), add `@pkgjson` lines to the existing `monom` config's table, and include stub PM executables the tests can put on `$PATH`.
- [ ] 5.2 Extend `tests/mnmd_resolve_complete_test`: mixed-table discovery output (file keys + prefixed scripts, table order), skipped script names, unreadable-manifest warning + exit 0.
- [ ] 5.3 Extend `tests/mnmd_resolve_run_test`: directive hit (npm default and `packageManager` pnpm), legacy-entry shadowing, miss passthrough, shape-matched loud errors, PM-not-on-PATH.
- [ ] 5.4 Add an end-to-end scenario through the full `monom` path in the fixture project: `monom web build` runs the stub PM with the expected argv, `monom db migrate` still runs the legacy script, and the `complete | mnmd filter` pipeline lists both `db` and `web` at the top level.

## 6. Docs and validation

- [ ] 6.1 Update `architecture.md`: document the two mapping-table entry kinds in the `resolve-run`/`resolve-complete` section (with a mid-migration mixed example), pack's newline-delimited output and absolute-first-token args rule, and the `_monom_run` exec change.
- [ ] 6.2 Update `terminology.md`: extend **Command mapping** with the `@pkgjson` directive (manifest-driven entries) and the entry-order precedence rule.
- [ ] 6.3 Run the validation checklist: `go vet`, `go test ./...`, shellcheck on all shell files, full shUnit2 suites; confirm no logic landed in shell and no new subprocess roundtrips on the completion or run path.
