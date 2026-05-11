# CLAUDE.md — AI Working Guide

Read these documents in order before making any changes:

1. `constitution.md` — governing principles. Every task must be validated against it.
2. `architecture.md` — current intended architecture: binary interface, shell files, data flow.
3. `terminology.md` — canonical term definitions. Use these exactly when naming anything.

---

## Canonical Terminology

| Term | Meaning |
|---|---|
| **Discovery** | Finding all possible commands for a project. Implemented by `monom_cfg complete`. |
| **Command packing** | Transforming CLI arguments back into a file path. Implemented by `monom_cfg run`. |
| **monom config file** | The `monom` executable at the project root. Exposes `complete` and `run`. Env var: `$MONOM_USER_CONFIG`. Shell wrapper: `monom_cfg() { "$MONOM_USER_CONFIG" "$@"; }`. |
| **Project root** | The directory containing the `monom` config file. Env var: `$MONOM_PROJECT_ROOT`. |
| **monomd** | The compiled Go binary. The engine of monom. |
| **CLI author** | The developer building a CLI using monom. |
| **CLI user** | The developer using the CLI the author built. |

---

## Testing

### Run all checks

```bash
./check
```

Runs: `build` → `go_e2e_test` → `sh_test_runner` → `shellcheck`.

### Run specific layers

```bash
cd src && go test ./...   # Go unit tests
./sh_test_runner          # shUnit2 e2e tests
./shellcheck              # shellcheck
```

### Conventions

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
- [ ] Terminology from `terminology.md` used consistently
