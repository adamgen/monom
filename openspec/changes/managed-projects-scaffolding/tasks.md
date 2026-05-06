## 1. Constitutional & Terminology Updates

- [ ] 1.1 Add "managed project" and "custom project" definitions to `terminology.md`
- [ ] 1.2 Add "Principle: Managed Projects and Scaffolding" section to `constitution.md`
- [ ] 1.3 Update `architecture.md` with managed project data flow and new monomd subcommands
- [ ] 1.4 Update `CLAUDE.md` with managed/custom project awareness and new subcommands table

## 2. Managed Project Detection

- [ ] 2.1 Extend `monomd root` to detect `monom.yaml` as a project marker (in addition to executable `monom`)
- [ ] 2.2 Implement precedence logic: `monom.yaml` wins when both exist
- [ ] 2.3 Add Go unit tests for project detection (yaml-only, executable-only, both-present)

## 3. monom.yaml Parsing

- [ ] 3.1 Add YAML parsing to monomd (add `gopkg.in/yaml.v3` dependency or similar)
- [ ] 3.2 Implement config struct with `commands` (required), `default_language` (optional), `run`/`complete` delegation (optional)
- [ ] 3.3 Validate required `commands` property with clear error on missing
- [ ] 3.4 Add Go unit tests for YAML parsing (valid, missing commands, with optional fields, with delegation)

## 4. Internal Discovery & Resolution (Managed Mode)

- [ ] 4.1 Implement `monomd complete` — walk the commands directory tree, output command paths
- [ ] 4.2 Implement resolution in `monomd run` for managed projects — resolve args to file path directly
- [ ] 4.3 Skip non-executable files during discovery
- [ ] 4.4 Support script delegation — when `run`/`complete` properties exist in YAML, delegate to those scripts
- [ ] 4.5 Add Go unit tests for internal discovery (flat, nested, non-executable excluded)
- [ ] 4.6 Add Go unit tests for internal resolution (simple, nested, not-found)

## 5. Shell Binding Updates

- [ ] 5.1 Update `monom_completion()` to detect managed vs custom and use appropriate data flow
- [ ] 5.2 Update `monom()` function to detect managed vs custom for command execution
- [ ] 5.3 Verify managed mode uses single subprocess call (no pipe through user config)

## 6. Scaffolding: monomd init

- [ ] 6.1 Implement `monomd init` — creates `monom.yaml` and `commands/` directory
- [ ] 6.2 Add `--language` flag (bash|python|node) setting `default_language` in generated YAML
- [ ] 6.3 Add interactive language prompt when stdin is a terminal and no `--language` flag
- [ ] 6.4 Error handling: refuse to init if project already exists (yaml or executable)
- [ ] 6.5 Add Go unit tests for init (happy path, already-exists, language flag)

## 7. Scaffolding: monomd new command

- [ ] 7.1 Implement `monomd new command <path>` — creates executable script in commands directory
- [ ] 7.2 Generate correct shebang based on language (project default or `--language` flag)
- [ ] 7.3 Set executable permission on generated file
- [ ] 7.4 Create intermediate directories for nested commands (e.g., `infra/provision`)
- [ ] 7.5 Error handling: command already exists, not in a project, in a custom project
- [ ] 7.6 Add Go unit tests for new command (bash/python/node, nested, errors)

## 8. End-to-End Tests

- [ ] 8.1 Create test fixture: managed project (monom.yaml + commands directory)
- [ ] 8.2 shUnit2 e2e test: tab completion in managed project
- [ ] 8.3 shUnit2 e2e test: command execution in managed project
- [ ] 8.4 shUnit2 e2e test: `monomd init` creates valid project
- [ ] 8.5 shUnit2 e2e test: `monomd new command` creates runnable command
- [ ] 8.6 shUnit2 e2e test: scaffolding refuses to operate in custom project
