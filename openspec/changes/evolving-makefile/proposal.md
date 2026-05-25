## Why

monom has no top-level Makefile, so contributors must memorize or discover the right commands to build, test, lint, and maintain the project. A Makefile provides a single, documented entry point for all common tasks and serves as living documentation that evolves alongside the project.

## What Changes

- **New**: `Makefile` at the repo root with targets covering build, test, lint, and maintenance workflows
- The Makefile is intentionally minimal at first and grows incrementally as the project needs more targets
- Each target is documented with a `## help` comment so `make help` self-describes the available commands

## Capabilities

### New Capabilities

- `makefile`: A root-level `Makefile` exposing `build`, `test`, `check`, `lint`, `clean`, and `help` targets for the monom project

### Modified Capabilities

## Impact

- Adds `Makefile` to the repo root — no existing files are modified
- All existing workflows (`go test ./...`, `shellcheck`, etc.) remain valid; the Makefile wraps them, it does not replace them
- Developers unfamiliar with the project get a self-describing entry point via `make help`
