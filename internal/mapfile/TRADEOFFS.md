# Tradeoffs — the command map

Decisions here were contested. The code shows *what* the map does; this file records *why* the losing options lost.

---

## JSON, not YAML

**Chosen:** `monom-map.json`, parsed with the standard library. A `"$note"` key, allowed anywhere and ignored, stands in for comments.

**Rejected:** YAML via `gopkg.in/yaml.v3` — nicer to hand-author (real comments, no quote ceremony), and the config-file lingua franca of the target audience. Also rejected: a flat line format (`db/migrate: scripts/x.sh`) parsed by hand, which keeps comments at zero dependencies but loses the nested outline and the extended node form.

**Why.** The map is parsed in exactly one place, behind two subcommands. That is not enough surface to justify monom's first external dependency — and `go.mod` being dependency-free is itself a property worth spending only reluctantly. JSON expresses the full node model (nesting, string shorthand, object form with room for future per-command keys) with nothing added.

**What it costs.** No comments — a real loss, because the map's primary use is annotated migration bookkeeping ("TODO: move this into the tree"). `$note` recovers most of it, at the price of a convention readers must learn. If the format ever needs to change, the single parse site keeps the swap contained.

---

## No match exits 0 with no output, instead of erroring

**Chosen:** `map resolve` prints nothing and exits 0 when the typed tokens match no map node (or continue past a command node). Combined with the run hook's empty-output fallback, unresolved tokens drop through to `mnmd pack` against the file tree.

**Rejected:** exiting non-zero with a "not in map" error, making the map authoritative for every invocation.

**Why.** The silent no-match is what makes the map an *overlay*: legacy scripts live in the map, new commands are written as ordinary tree files, and both coexist in one CLI with no mode switch. It is also the migration story — moving a script into the tree is just deleting its map entry. An authoritative map would force brownfield projects to register every command forever, reproducing the boilerplate monom exists to kill.

**What it costs.** A typo in a *mapped* command name surfaces as pack's generic `command not found` against a tree path that never existed, not as a map-aware message. `mnmd check` cannot catch typos the user types, only broken map entries; the error quality for this case is deliberately sacrificed to the overlay behavior.

---

## Hand-rolled token-stream parser, not `json.Unmarshal`

**Chosen:** `Parse` walks `json.Decoder` tokens and builds the node tree itself, tracking seen keys per object.

**Rejected:** `json.Unmarshal` into `map[string]any` (or `json.RawMessage`) and interpreting the result — a third of the code.

**Why.** `Unmarshal` silently keeps the last occurrence of a duplicate key. In a command map, a duplicate key is a swallowed command: the author defined it, the file looks right, and the CLI silently lacks it. That is precisely the class of bug `mnmd check` exists to make impossible, and it cannot be detected after `Unmarshal` has already collapsed the duplicate. The token walk also yields positioned, node-path-aware error messages for free.

**What it costs.** ~100 lines of parser where 30 would otherwise do, and the maintenance burden of by-hand JSON traversal.

---

## Resolve does not validate targets on disk

**Chosen:** `map resolve` emits the target's tokens without checking that the file exists or is executable. Parse-time validation covers only what would corrupt the pipeline (whitespace, absolute paths, root escapes).

**Rejected:** stat-ing the target inside resolve and failing early with a map-specific error.

**Why.** `mnmd pack` is the sole resolver: it already validates existence and executability for every command, mapped or not, and the shell already renders its errors. A second validation site would duplicate that contract and inevitably drift from it. Development-time diagnosis belongs to `mnmd check`, which validates every target with full map context.

**What it costs.** A broken target at run time surfaces as pack's `command not found: <root>/<target>` — naming the target path rather than the command the user typed, one step removed from the map entry that caused it.

---

## Reserved keys are rejected, not ignored

**Chosen:** `complete`, `pre-run`, and `post-run` as names anywhere in the map are a parse error ("reserved for future per-command hooks"), as is any `$`-prefixed key other than `$note`.

**Rejected:** accepting and ignoring unknown keys, the tolerant-reader convention.

**Why.** Tolerance here is a trap with a delay: an author who writes `"complete": "my_completions.sh"` against today's binary would see it silently do nothing, ship it, and get a behavior change the day the key is implemented. Rejecting now means the future keys can land without a breaking change and without a migration story.

**What it costs.** A map written for a future monom version fails loudly on an old binary instead of degrading gracefully, and authors cannot stash arbitrary metadata in the file beyond `$note`.
