# CLAUDE.md — AI Working Guide

Read these documents in order before making any changes:

1. `constitution.md` — governing principles. Every task must be validated against it.
2. `architecture.md` — current intended architecture: binary interface, shell files, data flow.
3. `terminology.md` — canonical term definitions. Use these exactly when naming anything.

---

## Spelling & Casing

The project name is **monom** — always lowercase, even at the start of a sentence.

---

## Testing

> **Note:** The test runner (`make check`) is pending build infrastructure — a separate change after `monomd-binary`. To run Go tests directly in the interim, use `go test ./...` from the repo root. See `openspec/changes/monomd-binary/`.

### Conventions (once implemented)

- Go unit tests: `*_test.go`, colocated with the file they test.
- shUnit2 tests: `${script_name}_test`, colocated, no extension.
- shUnit2 test functions: `test_descriptive_name()`.

### What to test where

| What | Tool |
|---|---|
| A Go function | Go unit test |
| Full CLI binary surface (stdin, args, stdout, exit codes) | shUnit2 e2e test |
| Completion behavior in a real shell environment | shUnit2 completion e2e test |
| Shell binding files | shUnit2 test |

---

## Project Layout

> **Note:** This section is temporary guidance during the `monomd-binary` change. Once all Go utilities and the binary entry point are implemented, this section will be removed — the layout will be self-evident from the repo.

```
monom/
├── go.mod
├── go.sum
├── Makefile
├── cmd/
│   └── monomd/
│       └── main.go          ← binary entry point (package main)
├── internal/
│   ├── filter/
│   │   ├── filter.go
│   │   └── filter_test.go
│   ├── pack/
│   │   ├── pack.go
│   │   └── pack_test.go
│   ├── root/
│   │   ├── root.go
│   │   └── root_test.go
│   └── check/
│       ├── check.go
│       └── check_test.go
└── bin/                     ← compiled output (.gitignore'd)
    └── monomd
```

Shell files (`src/monom`, `src/monom.bash`, `src/monom.zsh`) are a separate subsequent change.

---

## Common Tasks

### Build

```
make build    # compiles bin/monomd
make test     # go test ./...
make check    # test + shUnit2 e2e + shellcheck (pending build infrastructure change)
```

### Add a monomd subcommand

The four subcommands in scope for the initial binary are `filter`, `root`, `pack`, and `check`. `monomd args` is documented in `architecture.md` as TBD and is out of scope until a separate change defines it.

To add or extend a subcommand:

1. Add or update the logic package under `internal/<subcommand>/` with a matching `*_test.go`
2. Wire the dispatch in `cmd/monomd/main.go`
3. Add shUnit2 e2e tests covering stdin, args, stdout, and exit codes
4. Run `make check`

### Add a new shell binding

Shell files are a separate change after the binary. When that work begins:

1. Create `src/monom.<shell>` with only the completion registration
2. Update `src/monom` to source the right file based on shell detection
3. Verify shellcheck passes

---

## Validation Checklist

Before completing any task:

- [ ] No logic added to shell files
- [ ] No new subprocess roundtrips on the completion or run path
- [ ] Go unit tests added for any new Go logic
- [ ] `go vet` passes with no errors
- [ ] All shell files pass `shellcheck` with no suppressions (except those documented inline with an explanation)
- [ ] shUnit2 e2e test added or updated if CLI behavior changed
- [ ] `make check` passes (once build infrastructure exists)
- [ ] No new required subcommands added to the user config interface (amendment required — see `constitution.md`)
- [ ] Terminology from `terminology.md` used consistently (do not redefine terms inline)
