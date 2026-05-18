## Why

The project has moved past its prototyping stage and into a clearly defined architecture (documented in `architecture.md` and `constitution.md`). The existing `src/` code, shell utilities, and test projects reflect the old prototype design — which predates the `monomd` binary model, the clean shell/Go separation principle, and the current user config interface. Keeping this code in place creates confusion about what is canonical, makes the codebase hard to navigate, and creates friction when implementing the new architecture.

## What Changes

- Move all prototype-era code files (under `src/`, `test_projects/`, `sh_test_runner`, `go_e2e_test`, `build`, `check`, `shellcheck`, `install.sh`, and `dependencies/`) into a top-level `_archive/` folder.
- No markdown files, OpenSpec artifacts, or governance documents (`constitution.md`, `architecture.md`, `terminology.md`, `CLAUDE.md`, `README.md`, `old_notes.md`) are moved — they remain at the repo root.
- The `_archive/` folder is kept in the repository as a reference and removed in a future cleanup once no longer needed.
- The project root is left containing only governance docs, OpenSpec artifacts, and the Go module file (`go.mod`).

## Capabilities

### New Capabilities

- `archive-prototype-code`: A single atomic move of all prototype-era code into `_archive/`, preserving the files for reference while clearing the working root for clean implementation of the `monomd` architecture.

### Modified Capabilities

<!-- none -->

## Impact

- `src/` — all files moved to `_archive/src/`
- `test_projects/` — moved to `_archive/test_projects/`
- `dependencies/` — moved to `_archive/dependencies/`
- `sh_test_runner`, `go_e2e_test`, `build`, `check`, `shellcheck`, `install.sh` — moved to `_archive/`
- `go.mod` — stays at root (it will be updated to match the new Go source structure when `monomd` is implemented)
- No OpenSpec or markdown files are touched
- Existing `go_e2e_test` and `sh_test_runner` scripts will no longer be at the root; the `./check` script referenced in `CLAUDE.md` will no longer work until the new implementation is built
