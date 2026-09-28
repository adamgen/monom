# Tradeoffs — `mnmd check` severities

Decisions here were contested. The code shows *what* check does; this file records *why* the losing option lost.

---

## Only `root-contents` has a configurable severity

**Chosen:** `root-contents` is a warning by default and can be raised to an error per project (`check.root-contents`) or per user (`MONOM_CHECK_ROOT_CONTENTS`). Every other check is always an error.

**Rejected:** a severity setting for every check.

**Why.** The other checks report commands that are broken: a registered path with a space can never be completed, and a declaration that names nothing, or a config key that does not exist, is a mistake the author made in a file they wrote. Letting a project downgrade those would let CI pass over a CLI that does not work. `root-contents` is different in kind: it reports files that are *not* part of the CLI — which is normal on a fresh zero-config project, where the tree predates monom. Making that fatal would fail the first `mnmd check` a new user ever runs.

**What it costs.** An author who wants to silence a specific error-level finding has to fix it; there is no escape hatch.

---

## Warnings never change the exit code

**Chosen:** only error-severity problems make check exit non-zero.

**Rejected:** a distinct exit code for "warnings only".

**Why.** CI treats any non-zero exit as failure, so a warnings exit code would be an error with extra steps. Projects that want warnings to fail CI say so with `check.root-contents = error`, which is visible in their repository.

**What it costs.** A script cannot tell from the exit code alone whether there were warnings; it has to read stdout.

---

## The project setting overrides the user's

**Chosen:** `check.root-contents` in the project beats `MONOM_CHECK_ROOT_CONTENTS` in the user's environment.

**Rejected:** user-over-project, so an individual can always be stricter.

**Why.** Check runs in CI, where the result has to be the same for everyone. A project that states a severity has made a decision about its own tree; a developer's global preference fills the gap only where the project said nothing. This matches the `debug` hook's local-overrides-global precedence.

**What it costs.** A developer who wants strictness everywhere cannot get it in projects that explicitly chose `warning`.
