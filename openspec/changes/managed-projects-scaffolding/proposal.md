## Why

New users adopting monom face a cold-start problem: they know they want a CLI, but the path from zero to a working `monom` config file with tab completion requires understanding the `complete`/`run` interface, writing discovery logic, and setting up a directory structure. This friction contradicts the mission ("no boilerplate, no registration, no framework lock-in").

monom needs authoring-time tooling — commands like `monomd init` and `monomd new command` — backed by strong defaults that get a user from nothing to a working CLI in one command.

## What Changes

- Introduce the concept of **managed projects** vs **custom projects** as a constitutional term. A managed project uses a declarative `monom.yaml` config; a custom project uses an executable `monom` file implementing `complete`/`run`.
- Add `monom.yaml` as an alternative project marker. For managed projects, `monomd` handles discovery and resolution internally (fewer subprocess roundtrips).
- Add scaffolding subcommands to `monomd`: `init`, `new command`, potentially `new project`.
- Scaffolding supports multiple languages (bash, python, node) for generated command scripts.
- Add a constitutional principle governing scaffolding behavior and managed projects.
- `monom.yaml` can optionally point to an executable for `complete`/`run` — allowing a hybrid where you get the managed config but delegate specific operations to a script.
- The user's `monom` executable could optionally expose additional functions beyond `complete`/`run` to support scaffolding awareness (open question — needs design).

## Capabilities

### New Capabilities
- `managed-projects`: The managed project model — `monom.yaml` as declarative config, monomd-internal discovery and resolution, terminology, and the constitutional principle.
- `scaffolding`: Authoring-time commands (`monomd init`, `monomd new command`) that generate project structure and command files with strong defaults.

### Modified Capabilities
<!-- No existing specs to modify yet -->

## Impact

- `constitution.md` — new principle section, new terminology (managed/custom)
- `terminology.md` — new terms: managed project, custom project
- `architecture.md` — new monomd subcommands, monom.yaml format, alternative discovery path
- `monomd` binary — new subcommands (init, new), YAML parsing, internal discovery/resolution for managed projects
- `monomd root` — needs to discover both `monom` (executable) and `monom.yaml` as project markers
- Shell bindings — need to detect managed vs custom and adjust data flow (managed: single monomd call; custom: pipe through user config)
