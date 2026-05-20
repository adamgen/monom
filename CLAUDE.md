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

> **Note:** The test runner scripts (`./check`, `./sh_test_runner`, `./shellcheck`, `cd src && go test ./...`) are pending the `monomd-binary` implementation. They will be created as part of that change. See `openspec/changes/monomd-binary/`.

### Conventions (once implemented)

- Go unit tests: `*_test.go`, colocated with the file they test.
- shUnit2 tests: `${script_name}_test`, colocated, no extension.
- shUnit2 test functions: `test_descriptive_name()`.

### What to test where

| What | Tool |
|---|---|
| A Go function | Go unit test |
| Full CLI behavior | shUnit2 e2e test |
| Shell binding files | shUnit2 test |

---

## Common Tasks

> **Note:** The file paths below (`src/main.go`, `src/go_utils/`, `src/monom`) are pending the `monomd-binary` implementation. See `openspec/changes/monomd-binary/` for the planned structure.

### Add a monomd subcommand

1. Add the handler in `src/main.go`
2. Add logic in `src/go_utils/` with a matching `*_test.go`
3. Add an e2e test
4. Run `./check`

### Add a new shell binding

1. Create `src/monom.<shell>` with only the completion registration
2. Update `src/monom` to source the right file based on shell detection
3. Verify shellcheck passes

---

## Validation Checklist

Before completing any task:

- [ ] No logic added to shell files
- [ ] No new subprocess roundtrips on the completion or run path
- [ ] Go unit tests added for any new Go logic
- [ ] shUnit2 e2e test added or updated if CLI behavior changed
- [ ] `./check` passes
- [ ] No new required subcommands added to the user config interface
- [ ] Terminology from `terminology.md` used consistently (do not redefine terms inline)
