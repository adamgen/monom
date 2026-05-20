## 1. Move All Prototype Code

- [x] 1.1 Run: `mkdir _archive && mv src test_projects dependencies build check go_e2e_test sh_test_runner shellcheck install.sh _archive/`

## 2. Add Archive README

- [x] 2.1 Create `_archive/README.md` noting that this folder is historical prototype code kept for reference only — it is not functional, not maintained, and should not be used as a basis for new work. Include a note that once the `monomd` implementation is complete and stable, `_archive/` should be deleted.

## 3. Update CLAUDE.md

- [x] 3.1 Remove or stub out the Testing section commands (`./check`, `./sh_test_runner`, `./shellcheck`, `cd src && go test ./...`) with a note that these are pending the `monomd-binary` implementation
- [x] 3.2 Remove or stub out the Common Tasks section (`src/main.go`, `src/go_utils/`, `src/monom`) references with the same note

## 4. Verify

- [x] 4.1 Confirm `_archive/` contains all moved items with correct structure: `ls _archive/`
- [x] 4.2 Confirm none of the moved items exist at repo root and governance files are untouched: `ls` at repo root

## 5. Archive This OpenSpec Change

- [ ] 5.1 Once the full migration is complete — including `monomd`, shell bindings, tests, and everything else — delete `_archive/` and run `/opsx:archive` on this change
