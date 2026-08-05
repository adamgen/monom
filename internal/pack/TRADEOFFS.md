# Tradeoffs — `mnmd pack`

Decisions here were contested. The code shows *what* pack does; this file records *why* the losing option lost.

---

## Exit 3 is a pure signal — pack does not enumerate the group

**Chosen:** when the tokens resolve to a directory, pack writes nothing to stdout or stderr and exits 3. The shell turns that into a user-facing listing by re-running the discovery pipeline (`_monom_cfg complete | mnmd filter <tokens> ""`).

**Rejected:** having pack print the directory's children on exit 3, since it already has the path in hand and a `ReadDir` is one call away.

**Why.** The children pack can see are the *filesystem's* children. What the user is entitled to see are the *surface tree's* children — the set produced by the author's `complete` hook and reshaped by any `run` hook. Those two sets are allowed to differ; that divergence is the entire point of the hooks. A directory read would produce a listing that contradicts what `monom <group> <Tab>` offers a keystroke later, and the user would have no way to tell which one is lying.

Routing the listing through `complete | filter` costs an extra subprocess on a path that has already failed, and buys the guarantee that the listing and the completions cannot disagree — they are literally the same pipeline.

**What it costs.** Exit 3 is meaningless on its own: `mnmd pack infra` from a bare terminal prints nothing and returns 3, which reads as a silent failure. That is deliberate — pack is a resolver, and the human-facing rendering belongs to the shell layer that has the user's context. See `src/TRADEOFFS.md` for the message format.

---

## A command group is never runnable by default

**Chosen:** invoking a group is always the exit-3 signal. There is no built-in "run the default leaf" behavior and no flag to enable one.

**Rejected:** auto-dispatching a group to a conventionally-named child (`index`, `main`, `default`, or the directory's own name).

**Why.** This is clig.dev's "don't have a catch-all subcommand," and the failure mode is a time bomb rather than an inconvenience. Suppose `monom infra` silently runs `infra/default`. The day an author adds a leaf that the convention prefers — or renames the one it was resolving to — every existing invocation of `monom infra` changes meaning with no error, no warning, and no diff that looks related. Scripts and muscle memory both break quietly.

An author who wants `monom infra` to *do* something says so explicitly in their own config via the `run` hook, mapping `infra` → `infra cloud deploy`. The override is then visible in their repo, versioned with their project, and impossible to trigger by accident.

**What it costs.** The common case — a group with one obvious entry point — needs three lines of `run` hook instead of zero. Accepted: the cost is paid once per project, by the author, in a place they can read; the alternative charges every user, forever, at a moment when nothing appears wrong.
