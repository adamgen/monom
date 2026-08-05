# Tradeoffs — the shell layer

Decisions here were contested. The code shows *what* `src/monom` and the bindings do; this file records *why* the losing option lost.

Every entry answers the same standing question the constitution poses: *is there a technical reason this cannot be in Go?* Where the answer turned out to be "no", the logic moved and is not documented here.

---

## An absent `run` hook and a silent one are the same thing

**Chosen:** the `run` hook's exit code selects the behavior, and exit 0 with empty stdout means "fall back to the user's original args" — whether the hook is missing entirely or present and declining.

| hook result | `monom()` does |
| --- | --- |
| exit 0, empty stdout | use `"$@"` unchanged |
| exit 0, non-empty stdout | re-split the output, pass to `mnmd pack` |
| non-zero exit | forward the hook's stderr, return its exit code, do not exec |

**Rejected:** a sentinel exit code (e.g. 127, or a monom-specific 64) meaning "I don't implement this hook."

**Why.** The constitution requires hooks to be free of ceremony: an author adds one by implementing its arm and omits it by not implementing it. A config file that has no `run` case in its `case` statement falls through and exits 0 with no output — that is the natural shape of "did nothing," and it costs the author nothing to produce. Requiring a sentinel would mean every config file has to know about a protocol before it can decline to participate in it, which is precisely the registration step monom promises not to have.

Non-zero is treated as a real failure rather than a decline for the same reason: an author who exits non-zero did so deliberately, so surfacing it imposes no ceremony on anyone who didn't.

**What it costs.** A hook that is genuinely broken in a way that still exits 0 — a typo'd variable expanding to nothing — is indistinguishable from a hook that declined, and `monom` proceeds with the untransformed args. There is no diagnostic for this; `MONOM_DEBUG_LOG` is the only way to see it.

---

## The hook's output is re-split through an array, never passed bare

**Chosen:**

```bash
unpacked_args=($(printf '%s' "$hook_out"))
mnmd pack "${unpacked_args[@]}"
```

**Rejected:** `mnmd pack $hook_out`, relying on the shell to word-split.

**Why.** zsh does not word-split unquoted parameter expansions by default (`SH_WORD_SPLIT` is off). The bare form works in bash and silently fails in zsh: `pack` receives `"custom-folder db migrate"` as one argument, joins it to nothing, and fails to resolve a path that exists. The array makes the split explicit and identical in both shells.

The re-split is necessary at all because of an asymmetry in the hook contract: `run` *receives* argv but can only *emit* a flat stdout stream, and it is allowed to change the argument count — that is the point of aliasing. So `"$@"` cannot be reused after the hook runs.

---

## The group listing comes from `complete | filter`, not from `pack`

**Chosen:** on pack's exit 3, `_monom_list_group` re-runs `_monom_cfg complete | mnmd filter <tokens> ""` to produce the `available:` line.

**Rejected:** having pack print the children it already has, saving a subprocess on a path that has already failed.

**Why.** Detailed in `internal/pack/TRADEOFFS.md` — the short version is that the filesystem's children and the surface tree's children are allowed to differ, and a listing that contradicts `monom <group> <Tab>` is worse than a slower one. The trailing `""` is what tells `filter` to drill into the level rather than prefix-match it.

---

## `exec` runs in a subshell

**Chosen:** `( exec "$pack_out" )`.

**Rejected:** a bare `exec "$pack_out"`.

**Why.** `monom()` is a shell *function* running in the user's interactive session. A bare `exec` replaces that session's process with the command — the terminal exits when the command finishes. The subshell contains the replacement.

**What it costs.** One extra fork per command invocation, and `monom` can no longer be used to replace the current shell deliberately. The exec inside the subshell is still worth keeping: it avoids a second fork for the command itself.

---

## The `debug` hook is queried on every invocation

**Chosen:** `_setup_monom` runs `_monom_cfg debug` unconditionally, on both the Tab path and the execution path, and exports the result as `MONOM_DEBUG_LOG` when it names a writable file.

**Rejected:** querying it only when logging is already enabled, or caching the result for the session.

**Why.** The hook's most valuable use is turning logging *on* for one project while the global switch is off — so gating the query on the global switch would disable the feature's main purpose. Caching would break the moment the user `cd`s to another project, which is the normal way to use a per-project setting.

The precedence — a valid local path overrides the global value, an invalid one falls back to it — exists so that a broken hook degrades to the user's explicit global choice rather than to silence.

**What it costs.** One unconditional subprocess spawn per invocation, including every Tab press, plus one append-probe when the hook prints a path. The config file is opaque, so running it is the only way to discover whether it defines `debug` at all. This is the same attempt-and-fallback cost model as the `run` hook, and it is the largest fixed cost monom imposes on the completion path.

**Why the diagnostics differ by path.** An invalid hook warns on stderr during execution but only reaches `_monom_log` during completion — the completion handlers call `_setup_monom 2>/dev/null` precisely because stderr mid-Tab corrupts the prompt line. Same rule as `internal/filter/TRADEOFFS.md`.

---

## Writability is probed in a subshell

`( : >> "$debug_candidate" ) 2>/dev/null` rather than a direct redirect. `:` is a POSIX special builtin, so a redirection failure on it aborts the *calling* shell — which here is the user's interactive session. The subshell absorbs the abort. This is a correctness requirement, not a style choice.
