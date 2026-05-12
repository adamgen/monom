## Context

monom currently has one project model: the user writes an executable `monom` file implementing `complete` and `run`. This gives full flexibility but creates friction for new users who just want a conventional file-tree-based CLI.

The proposal introduces a second model — "managed projects" — where a declarative `monom.yaml` replaces the executable config. For managed projects, `monomd` handles discovery and resolution internally, and scaffolding commands help users create projects and commands.

Key constraints:
- The existing custom project model must remain fully supported and first-class
- The constitution's "Go owns logic" and "minimize subprocess roundtrips" principles apply
- Scaffolding must feel like an on-ramp, not a prescription

## Goals / Non-Goals

**Goals:**
- Define the managed project model and its `monom.yaml` config
- Define scaffolding commands and their behavior
- Establish how managed and custom projects coexist
- Add constitutional principle governing this area
- Support bash, python, and node as scaffolding language choices

**Non-Goals:**
- Building a plugin system or extensibility framework
- Replacing the custom project model
- Supporting every possible directory layout (managed projects follow ONE convention)
- Auto-migration from custom to managed or vice versa

## Decisions

### Decision 1: Two project modes — managed and custom

**Choice**: monom recognizes two project types based on what exists at the project root.

| Mode | Marker file | Who handles discovery/resolution |
|------|-------------|----------------------------------|
| Managed | `monom.yaml` | `monomd` internally |
| Custom | `monom` (executable) | User's script via `complete`/`run` interface |

**Why not just one mode**: The executable interface is powerful but requires writing discovery logic. Most projects just want "files in a folder = commands." The managed mode serves the 80% case with zero logic.

**Why not deprecate custom**: Custom projects are the escape hatch. They're needed for non-standard layouts, dynamic discovery, or any case where the file tree doesn't directly map to commands.

**Coexistence**: If both `monom.yaml` and `monom` (executable) exist, `monom.yaml` takes precedence. The YAML can optionally delegate to the executable for specific operations (see Decision 3).

### Decision 2: monom.yaml — minimal declarative config

**Choice**: The YAML file is deliberately minimal. Core properties only:

```yaml
# monom.yaml — minimum viable
commands: ./commands
```

```yaml
# monom.yaml — with optional properties
commands: ./commands
default_language: python
```

Core properties (protected, changes require constitution review):
- `commands` — path to the commands directory (required)

Optional properties (can evolve without constitutional amendment):
- `default_language` — language for scaffolded command scripts (bash|python|node)

**Why minimal**: Every property added is a maintenance burden and a potential source of confusion. Start with the minimum and add only when there's proven need.

**Why `commands` is the only required property**: It's the one thing monomd can't guess. Everything else has sensible defaults.

### Decision 3: Hybrid mode — YAML with script delegation

**Choice**: `monom.yaml` can optionally point to an executable for operations that need custom logic:

```yaml
commands: ./commands
run: ./monom-run    # delegate 'run' resolution to this script
```

This allows a project to be mostly managed (monomd handles discovery) but with custom resolution logic. The script interface for delegated operations is the same as the custom project interface — `$script run <args>`.

**Why**: Some projects need conventional discovery (simple tree walk) but non-trivial resolution (e.g., checking permissions, selecting variants). This avoids forcing them fully into custom mode.

**Open question**: Should `complete` also be delegatable? Or just `run`? Leaning toward both for symmetry, but it might be YAGNI.

### Decision 4: Scaffolding lives in monomd

**Choice**: All scaffolding commands are subcommands of `monomd`:
- `monomd init` — creates `monom.yaml` + `commands/` directory in current project
- `monomd new command <path>` — creates an executable command script

**Why in monomd**: Single binary, Go owns logic, one tool to learn. No reason for a separate binary.

**Why not `monomd new project`**: It's just `mkdir <name> && cd <name> && monomd init`. Adding a command for two trivial steps adds surface area without value.

### Decision 5: Scaffolding language selection

**Choice**: `monomd init` and `monomd new command` ask for a language (or accept a flag):
- bash (default)
- python
- node

The generated command script includes:
- Appropriate shebang (`#!/bin/bash`, `#!/usr/bin/env python3`, `#!/usr/bin/env node`)
- Executable permission set
- Minimal runnable body (e.g., `echo "TODO: implement"`)

**Why these three**: They're the most common scripting languages. More can be added later. The generated script is so minimal that language choice mainly means "which shebang."

### Decision 6: monomd root discovers both markers

**Choice**: `monomd root` walks up from `$PWD` looking for either `monom.yaml` or `monom` (executable). First match wins. If both exist in the same directory, `monom.yaml` takes precedence.

**Why YAML takes precedence**: If you have both, you're likely migrating from custom to managed. The YAML represents your intent.

### Decision 7: Data flow for managed projects

**Choice**: Managed projects use a simplified data flow with fewer subprocess roundtrips:

```
Tab press (managed):
  shell → monomd complete <prefix>    (one call — Go walks tree + filters)

Command run (managed):
  shell → monomd run <args>           (one call — Go resolves directly) → exec

Tab press (custom — unchanged):
  shell → monom_cfg complete | monomd filter <prefix>

Command run (custom — unchanged):
  shell → monomd run → monom_cfg run → print path → exec
```

**Why**: The "minimize subprocess roundtrips" principle. For managed projects, monomd has all the information it needs — no reason to spawn a user script.

### Decision 8: Constitutional principle text

**Choice**: Add to `constitution.md`:

A new "Principle: Managed Projects and Scaffolding" section covering:
- The managed/custom dichotomy as a first-class concept
- Constraints on scaffolding behavior (transparent, editable, minimal, on-ramp not prescription)
- `monom.yaml` `commands` property as protected (changes need constitutional review)
- The default convention: file tree = command tree, literally

Also add "managed project" and "custom project" to `terminology.md`.

### Decision 9: Scaffolding hints for non-managed projects

**Choice**: When a user runs `monomd new command` in a custom project, monomd prints a helpful message:

```
This is a custom project (uses executable monom config).
monomd can't scaffold commands because your config defines the structure.
Create your command manually according to your project's conventions.
```

**Why**: Better than silently failing or guessing wrong.

## Risks / Trade-offs

**[Two project modes add cognitive load]** → Mitigated by making managed the default for new users. Custom is documented as "advanced mode" for when you need it.

**[monom.yaml schema becomes a compatibility surface]** → Mitigated by keeping it ultra-minimal (one required property). Only `commands` is constitutionally protected.

**[Scaffolding generates code that becomes stale]** → Mitigated by generating minimal code. The `monom.yaml` has almost nothing to become stale. Command scripts are just a shebang + placeholder.

**[Shell bindings must branch on project type]** → One `if` statement (check for monom.yaml vs monom executable). Adds minimal complexity.

## Open Questions

1. **Should the user's monom executable be able to expose additional functions beyond `complete`/`run`?** For example, a `scaffold` function that tells monomd where to put new commands. This would let custom projects opt into scaffolding. Deferred — solve only if users actually ask for it.

2. **Should `monom.yaml` support a `name` property for alias setup?** e.g., `name: acme` would make `monomd init` also set up `acme` as an alias. Deferred — alias story is a separate concern.

3. **Exact behavior when both `monom.yaml` and executable `monom` exist**: Is this an error, a warning, or silently YAML-wins? Current decision is YAML-wins silently, but could revisit.
