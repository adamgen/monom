## MODIFIED Requirements

### Requirement: monom function dispatches via mnmd pack
`monom()` SHALL call `_setup_monom`, then resolve the command path via `mnmd pack "$@"` (the wrapper), and exec the resolved path. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `mnmd pack` instead of the original args.

The `run` hook's exit code SHALL select the behavior:

- **exit 0 with empty stdout** — hook absent or no transform. `monom()` SHALL fall back to `"$@"`. Absent and empty are merged on purpose: a config that omits the `run` arm exits 0 with no output, and the constitution's zero-ceremony hooks principle forbids requiring a sentinel to disambiguate them.
- **exit 0 with non-empty stdout** — the hook transformed the args. `monom()` SHALL use the hook's output.
- **non-zero exit** — hook present and failed. `monom()` SHALL surface the hook's stderr, abort with its exit code, and SHALL NOT fall back or exec. A non-zero exit is an explicit failure the author raised, so surfacing it imposes no ceremony.

The hook's stderr SHALL be captured and forwarded to the user on failure rather than discarded.

`monom()` SHALL branch on `mnmd pack`'s exit code:

- **exit 0** — a leaf command was resolved. `monom()` SHALL exec the path printed on stdout.
- **exit 3** — the args resolved to a command group (a directory), not a runnable command. `mnmd pack` produced no output (exit 3 is a pure signal). `monom()` SHALL print a concise user-facing message to stderr and SHALL return non-zero without exec'ing anything. The message SHALL identify the group by the last token the user typed and SHALL list the available children, and SHALL NOT include the `mnmd pack:` prefix. Format:

```
monom: 'infra' is a command group
available: cloud, local
```

  `monom()` SHALL obtain the child listing from the canonical discovery source — the same pipeline tab-completion uses: `_monom_cfg complete | mnmd filter <tokens> ""`, where the trailing empty word tells `filter` to drill into the group's level. This makes the listing identical to what `monom <group> <Tab>` shows and honors any `run`-hook surface tree, because `complete` is the single source of truth rather than a direct filesystem read. When that pipeline yields no children (empty group, or a group whose children `complete` does not surface), `monom()` SHALL still report that it is a command group, omitting the `available:` line.
- **any other non-zero exit** — a real error. `monom()` SHALL forward `mnmd pack`'s stderr and return non-zero without exec'ing anything.

The args flow through three parts. Both `_monom_cfg run` and `mnmd pack` **receive** the args as separate CLI arguments — that input format is identical. The asymmetry is on `run`'s **output**: a hook is a separate process, so it can only emit a flat stdout stream, not an argv array. `monom()` therefore re-splits that stream back into separate args before handing them to `pack`.

The hook may also change the *number* of args — that is its purpose (aliasing, namespace remapping). Below, the hook prepends `custom-folder`, turning 2 args into 3:

```
monom db migrate
  → "$@"  = ["db", "migrate"]                       # separate args
  → _monom_cfg run db migrate                        # IN: separate args
        ↳ prints "custom-folder db migrate\n"       # OUT: one flat stream (transformed: 2 → 3 args)
  → (monom re-splits the stream on whitespace)
  → mnmd pack custom-folder db migrate            # IN: separate args
        ↳ joins with "/", resolves custom-folder/db/migrate
```

Because the hook can emit a different arg count than it received, `monom()` cannot reuse `"$@"` — it must parse the hook's actual output. The re-split SHALL be done via an array, never a bare unquoted string handed to `pack`: zsh does not word-split unquoted parameters by default (`SH_WORD_SPLIT` off), so `mnmd pack $string` would pass `"custom-folder db migrate"` as a single argument and fail to resolve `custom-folder/db/migrate`.

#### Scenario: Command execution without run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` exits 0 with no output (hook absent or declined)
- **THEN** `mnmd pack deploy` is called and its output is exec'd

#### Scenario: Command execution with run hook
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` outputs `infra deploy`
- **THEN** `mnmd pack infra deploy` is called and its output is exec'd

#### Scenario: run hook failure aborts and surfaces the error
- **WHEN** `monom deploy` is called and `$_MONOM_USER_CONFIG run deploy` exits non-zero
- **THEN** `monom` forwards the hook's stderr, exits with the hook's exit code, and does not call `mnmd pack` or exec anything

#### Scenario: Multi-word command preserves separate args in both shells
- **WHEN** `monom db migrate` is called (no `run` hook) in either bash or zsh
- **THEN** `mnmd pack` receives `db` and `migrate` as two separate arguments and resolves `db/migrate`, not a single `"db migrate"` argument

#### Scenario: monom exits non-zero when mnmd pack fails
- **WHEN** `mnmd pack` exits non-zero with a real error (command not found)
- **THEN** `monom` forwards `mnmd pack`'s stderr and exits non-zero without exec'ing anything

#### Scenario: Command group invocation lists children from complete instead of failing
- **WHEN** `monom infra` is called, `mnmd pack infra` exits 3, and `_monom_cfg complete` lists `infra/cloud/...` and `infra/local/...`
- **THEN** `monom` runs `_monom_cfg complete | mnmd filter infra ""`, prints `monom: 'infra' is a command group` and `available: cloud, local` to stderr, exits non-zero, and does not exec anything

#### Scenario: Group message uses the last typed token as the group name
- **WHEN** `monom infra cloud` is called, `mnmd pack infra cloud` exits 3, and `complete` lists `infra/cloud/deploy`
- **THEN** `monom` prints `monom: 'cloud' is a command group` and `available: deploy` to stderr and exits non-zero

#### Scenario: Group children reflect the complete hook, not the raw filesystem
- **WHEN** `monom infra` is called, `mnmd pack infra` exits 3, and `complete` surfaces a child set that differs from the directory's raw entries
- **THEN** the `available:` list matches the `complete | mnmd filter` output (the surface tree), not the directory's raw entries

#### Scenario: Empty command group is reported without an available list
- **WHEN** `monom infra` is called, `mnmd pack infra` exits 3, and `complete | mnmd filter infra ""` yields no children
- **THEN** `monom` reports that `infra` is a command group, omits the `available:` line, exits non-zero, and does not exec anything
