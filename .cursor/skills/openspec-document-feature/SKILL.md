---
name: openspec-document-feature
description: Document an already-built feature as a standard OpenSpec change. Use when a feature has ALREADY been implemented (by hand, in a previous session, or before adopting OpenSpec) and you need a canonical, validated spec in openspec/specs/ without pretending the work hasn't happened. This is the brownfield inverse of /opsx-propose.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: monom
  version: "1.0"
---

Document an **already-implemented** feature by driving it through the OpenSpec
change lifecycle in reverse: investigate the real code, capture its as-built
behavior as validated delta specs, and hand off to `/opsx-archive` to sync the
spec into `openspec/specs/`.

The common case is messy: the feature was written **on the fly** — no plan, no
proposal, often no tests — and only *in retrospect* does it deserve a spec. So
treat the **code as the source of truth**. Tests, if they exist, are a bonus.
There was usually never a real task breakdown, so do not fabricate one — see the
tasks guidance below.

**This is documentation, not implementation.** The feature already exists. You
MUST NOT edit, refactor, or "improve" the implementation. Your only outputs are
OpenSpec artifacts under the change directory. If the user wants to *change*
behavior, that's a different workflow — use `/opsx-propose`.

---

**Input**: The name of the feature to document and/or a description of where it
lives in the codebase (a subcommand, a module, a file). If unclear, ask.

**Steps**

1. **Identify the feature (ask if unclear)**

   Use the **AskUserQuestion tool** if you don't know which feature to document:
   > "Which already-built feature should I document? Point me at the subcommand,
   > module, or file(s)."

   Derive a kebab-case capability name (e.g. "the pack subcommand" →
   `pack-subcommand`). **IMPORTANT**: Do NOT proceed without knowing what to
   document.

2. **Investigate the real code FIRST — do not skip this**

   The code is the source of truth. Before writing any spec:
   - Read the implementing source file(s). This is the authoritative input —
     the feature may have been written on the fly with nothing else around it.
   - Read the feature's tests **if they exist** — they're the richest source of
     concrete scenarios (each test is a candidate `#### Scenario:`). But assume
     there may be **none**; do not block on missing tests.
   - Map the observable behavior: inputs, outputs, exit codes, env vars, edge
     cases, error handling — by reading the code, and by running it if that's
     the only way to confirm behavior.

   Every requirement and scenario you later write MUST be grounded in something
   you actually saw — in the code, in a test, or in observed runtime behavior.
   Do not invent behavior. If behavior is ambiguous and you can't confirm it,
   note it as an open question and ask rather than guess.

3. **Check for an existing spec**

   ```bash
   ls openspec/specs/
   ```
   - If `openspec/specs/<capability>/` does **not** exist → the delta will be a
     `## ADDED Requirements` section (new spec for the existing feature).
   - If it **does** exist → use `## MODIFIED Requirements` (copy the full
     existing requirement block before editing) or warn the user that this
     capability is already documented.

4. **Create the change scaffold**

   ```bash
   openspec new change "document-<capability>"
   ```

5. **Get the artifact build order**

   ```bash
   openspec status --change "document-<capability>" --json
   ```
   Parse `applyRequires` and the `artifacts` list. Use the **TodoWrite tool** to
   track progress through the artifacts.

6. **Author artifacts in dependency order**

   For each `ready` artifact, get its instructions and follow them:
   ```bash
   openspec instructions <artifact-id> --change "document-<capability>" --json
   ```

   - **proposal.md** — Why this feature deserves a canonical spec and what
     capability it covers. Frame it as "documenting existing behavior," not new
     work.
   - **design.md** — Only if warranted (cross-cutting, non-obvious decisions).
     For a simple feature, you may skip it if the schema doesn't require it.
   - **specs/<capability>/spec.md** — The as-built spec. Use the EXACT format:
     - `### Requirement: <name>` with SHALL/MUST normative language
     - `#### Scenario: <name>` (exactly four `#`) with `- **WHEN**` / `- **THEN**`
     - Every requirement needs at least one scenario.
     - Derive scenarios directly from the tests and observed behavior.
   - **tasks.md** — A real task breakdown usually never existed (the feature was
     written on the fly), so do NOT fabricate a detailed implementation plan in
     retrospect. Reconstruct it minimally and honestly:
     - If `tasks.md` is required by the schema, write a single retrospective
       checked item, e.g.
       `- [x] 1.1 Implemented as-built in <path> — documented retrospectively`.
     - Do NOT reverse-engineer a fictional multi-step plan to make it look
       planned. The implementation tasks are done by definition; that's not the
       point of this artifact here.
     - The genuine value is the **gaps** the investigation surfaced. Record each
       as an unchecked follow-up `- [ ]` so it isn't lost — most commonly
       missing test coverage (likely, since the code was unplanned), but also
       undocumented edge cases, or behavior you couldn't confirm.
     - If the schema does NOT require `tasks.md`, you may skip it entirely.

   After each artifact, re-run `openspec status ... --json` until all
   `applyRequires` artifacts are `done`.

7. **Validate — this is the whole point**

   ```bash
   openspec validate "document-<capability>" --strict
   ```
   Fix any format errors until validation passes. The value of this workflow is
   producing output the OpenSpec CLI accepts with zero manual edits.

8. **Hand off to archive**

   Do NOT sync or archive yourself with custom logic. Tell the user:
   > "The as-built spec for `<capability>` is authored and validated. Run
   > `/opsx-archive` to sync it into `openspec/specs/<capability>/` and archive
   > the change."

**Output**

Summarize:
- Capability documented and the change name/location
- Where the feature lives (files you grounded the spec in)
- Confirmation that `openspec validate --strict` passed
- Next step: "Run `/opsx-archive` to sync into openspec/specs/."

**Guardrails**
- NEVER modify the feature's implementation. Documentation only.
- NEVER write requirements/scenarios not grounded in real code, tests, or
  observed runtime behavior.
- Treat the code as the source of truth; assume tests may not exist and don't
  block on them.
- ALWAYS use the exact `### Requirement:` / `#### Scenario:` format (four `#`
  for scenarios) — wrong heading depth fails silently.
- If the capability already has a spec, prefer `MODIFIED` or warn — don't create
  duplicates.
- Don't fabricate a retrospective task plan. Reconstruct `tasks.md` minimally
  (one checked "implemented as-built" item) and capture real gaps as `- [ ]`.
- Always end by validating and pointing the user at `/opsx-archive`.
