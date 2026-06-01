---
name: /opsx-document-feature
id: opsx-document-feature
category: Workflow
description: Document an already-built feature as a standard, validated OpenSpec change
---

Document an **already-implemented** feature as a standard OpenSpec change — the
brownfield inverse of `/opsx-propose`. Use this when a feature already exists in
the code and needs a canonical spec in `openspec/specs/`.

I'll investigate the real code, capture its as-built behavior as validated delta
specs, and leave the change ready for `/opsx-archive` to sync into
`openspec/specs/`.

The common case is messy: the feature was written **on the fly** — no plan, no
proposal, often no tests — and only in retrospect deserves a spec. So I treat
the **code as the source of truth**; tests are a bonus, and there was usually
never a real task plan to reconstruct.

**This is documentation, not implementation.** I will not edit the feature's
code. If you want to change behavior, use `/opsx-propose` instead.

---

**Input**: The argument after `/opsx-document-feature` is the feature to
document — a capability name, subcommand, module, or file path.

**Steps**

1. **If no feature is given, ask which one to document**

   Use the **AskUserQuestion tool**:
   > "Which already-built feature should I document? Point me at the subcommand,
   > module, or file(s)."

   Derive a kebab-case capability name. Do NOT proceed without knowing the
   target.

2. **Investigate the real code FIRST**

   The code is the source of truth. Read the implementing source file(s), and
   their tests **if any exist** (don't block on missing tests). Map observable
   behavior: inputs, outputs, exit codes, env vars, edge cases, errors — reading
   the code, and running it if that's the only way to confirm behavior. Every
   requirement and scenario MUST be grounded in code, a test, or observed
   behavior. Do not invent behavior; note anything unconfirmable as an open
   question.

3. **Check for an existing spec**

   ```bash
   ls openspec/specs/
   ```
   - No existing `openspec/specs/<capability>/` → author an `## ADDED Requirements` delta.
   - Existing spec → use `## MODIFIED Requirements` (copy the full block first) or warn the user.

4. **Create the change scaffold**

   ```bash
   openspec new change "document-<capability>"
   ```

5. **Get the artifact build order**

   ```bash
   openspec status --change "document-<capability>" --json
   ```
   Parse `applyRequires` and `artifacts`. Track progress with the **TodoWrite tool**.

6. **Author artifacts in dependency order**

   For each `ready` artifact:
   ```bash
   openspec instructions <artifact-id> --change "document-<capability>" --json
   ```
   - **proposal.md** — why this existing feature needs a canonical spec.
   - **design.md** — only if warranted.
   - **specs/<capability>/spec.md** — the as-built spec, using the EXACT format:
     `### Requirement: <name>` (SHALL/MUST) and `#### Scenario: <name>` (exactly
     four `#`) with `- **WHEN**` / `- **THEN**`. Derive scenarios from the code
     (and tests, if any).
   - **tasks.md** — there was usually no real plan, so don't fabricate one. If
     required, write one retrospective checked item
     (`- [x] 1.1 Implemented as-built in <path> — documented retrospectively`).
     Capture the real value as unchecked gaps `- [ ]` (most often missing
     tests). Skip the file if the schema doesn't require it.

   Re-run `openspec status ... --json` after each until all `applyRequires` are `done`.

7. **Validate (the whole point)**

   ```bash
   openspec validate "document-<capability>" --strict
   ```
   Fix format errors until it passes.

8. **Hand off to archive**

   Report success and instruct the user to run `/opsx-archive` to sync the spec
   into `openspec/specs/<capability>/` and archive the change.

**Output**

- Capability documented + change name/location
- Files the spec was grounded in
- Confirmation that `openspec validate --strict` passed
- Next step: "Run `/opsx-archive` to sync into openspec/specs/."

**Guardrails**
- NEVER modify the feature's implementation — documentation only.
- NEVER write requirements/scenarios not grounded in code, tests, or observed behavior.
- Treat the code as the source of truth; assume tests may not exist.
- ALWAYS use exact `### Requirement:` / `#### Scenario:` format (four `#`).
- If the capability already has a spec, prefer `MODIFIED` or warn.
- Don't fabricate a retrospective task plan; reconstruct `tasks.md` minimally and capture gaps as `- [ ]`.
- Always end by validating and pointing the user to `/opsx-archive`.
