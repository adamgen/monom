# Tradeoffs — project root resolution

Decisions here were contested. The code shows *what* `FindProjectRoot` does; this file records *why* the losing option lost.

---

## Without a monom file, fall back to the git root — never to the working directory

**Chosen:** pin (`$_MONOM_PROJECT_ROOT`) → nearest `monom` file → nearest git root → error.

**Rejected:** treating `$PWD` as the root when nothing else matches.

**Why.** Zero-config projects need a root without a marker file, and the working directory is the only answer that always exists. But `monom <Tab>` typed in `$HOME` would then run default discovery over the whole home directory on every keystroke, and offer every executable in it as a command. A root has to be something the user chose: pinned by an alias, marked by a file, or bounded by version control. The git root is the one boundary nearly every project already has.

**What it costs.** A zero-config project that is not in a git repository needs either `touch monom` or an alias pin. And every git repository is now a monom project: `monom <Tab>` inside any repo offers its gated executables.

---

## Any regular file named `monom` marks a root, not only an executable one

**Chosen:** `touch monom` is enough.

**Rejected:** keeping the executable-bit requirement.

**Why.** An empty or declarative config is never executed, so requiring the execute bit would demand a meaningless `chmod +x` before the most common onboarding step works — and would make `touch monom` silently fail to mark anything.

**What it costs.** A stray non-executable file named `monom` (the sourced library `src/monom` in the monom checkout is one) now marks its directory as a root. The walk stops there instead of continuing upward.
