## 1. Makefile scaffold

- [x] 1.1 Create `Makefile` at the repo root with `.PHONY` declaration for all targets
- [x] 1.2 Add `help` target as the first (default) target, using `awk` to extract `## comments`

## 2. Core targets

- [x] 2.1 Add `build` target: `mkdir -p bin && go build -o bin/monomd ./cmd/monomd`
- [x] 2.2 Add `test` target: `go test ./...`
- [x] 2.3 Add `lint` target: `shellcheck src/monom* tests/*` (respects `.shellcheckrc`)
- [x] 2.4 Add `clean` target: `rm -f bin/monomd`
- [x] 2.5 Add `check` target with `build` as prerequisite: runs `test`, all shUnit2 suites under `tests/`, and `lint` in sequence

## 3. Validation

- [x] 3.1 Run `make help` and verify all targets are listed with descriptions
- [x] 3.2 Run `make build` and verify `bin/monomd` is produced
- [x] 3.3 Run `make test` and verify Go tests pass
- [x] 3.4 Run `make lint` and verify shellcheck passes with no suppressions
- [x] 3.5 Run `make clean` and verify `bin/monomd` is removed without error; re-run on clean state and verify no error
- [x] 3.6 Run `make check` end-to-end and verify it builds, tests, runs e2e suites, and lints in order
