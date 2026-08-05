## MODIFIED Requirements

### Requirement: monom function dispatches via mnmd pack
`monom()` SHALL call `_setup_monom`, then resolve the command via `mnmd pack "$@"`, and exec the result. If the optional `run` hook is present and returns usable output, its output SHALL be passed to `mnmd pack` instead of the original args.

The `run` hook's exit code SHALL select the behavior:

- **exit 0 with empty stdout** — hook absent or no transform. `monom()` SHALL fall back to `"$@"`. Absent and empty are merged on purpose: a config that omits the `run` arm exits 0 with no output, and the constitution's zero-ceremony hooks principle forbids requiring a sentinel to disambiguate them.
- **exit 0 with non-empty stdout** — the hook transformed the args. `monom()` SHALL use the hook's output.
- **non-zero exit** — hook present and failed. `monom()` SHALL surface the hook's stderr, abort with its exit code, and SHALL NOT fall back or exec. A non-zero exit is an explicit failure the author raised, so surfacing it imposes no ceremony.

The hook's stderr SHALL be captured and forwarded to the user on failure rather than discarded.

The args flow through three parts. Both `_monom_cfg run` and `mnmd pack` **receive** the args as separate CLI arguments — that input format is identical. The asymmetry is on `run`'s **output**: a hook is a separate process, so it can only emit a flat stdout stream, not an argv array. `monom()` therefore re-splits that stream back into separate args before handing them to `pack`.

The hook may also change the *number* of args — that is its purpose (aliasing, namespace remapping). Below, the hook prepends `custom-folder`, turning 2 args into 3:

```
monom db migrate
  → "$@"  = ["db", "migrate"]                       # separate args
  → _monom_cfg run db migrate                       # IN: separate args
        ↳ prints "custom-folder db migrate\n"       # OUT: one flat stream (transformed: 2 → 3 args)
  → (monom re-splits the stream on whitespace)
  → mnmd pack custom-folder db migrate              # IN: separate args
        ↳ joins with "/", resolves custom-folder/db/migrate
```

Because the hook can emit a different arg count than it received, `monom()` cannot reuse `"$@"` — it must parse the hook's actual output. The re-split SHALL be done via an array, never a bare unquoted string handed to `pack`: zsh does not word-split unquoted parameters by default (`SH_WORD_SPLIT` off), so `mnmd pack $string` would pass `"custom-folder db migrate"` as a single argument and fail to resolve `custom-folder/db/migrate`.

`mnmd pack`'s output is newline-delimited tokens: the executable's absolute path on the first line, followed by one argument per line (a plain tree command resolves to exactly one line). On pack success, `monom()` SHALL split pack's stdout on newlines **only** — never on spaces or tabs — into an array and exec that array (`exec "${tokens[@]}"` in a subshell). Newline-only splitting is what keeps a resolved path containing spaces (e.g. a project root under `"My Projects"`) working as a single argv element, while still delivering hook-produced arguments (e.g. a `resolve-run` `@pkgjson` exec line) as separate argv elements. The split SHALL work in both bash and zsh.

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

#### Scenario: Multi-line pack output is exec'd as executable plus arguments
- **WHEN** `mnmd pack` succeeds and prints `/usr/local/bin/npm`, `--prefix`, `/repo/apps/web`, `run`, `build` on five lines, in either bash or zsh
- **THEN** `monom` execs `/usr/local/bin/npm` with the four following lines as its four arguments, in order

#### Scenario: Resolved path containing spaces stays one argv element
- **WHEN** `mnmd pack` succeeds and prints a single line `/home/user/My Projects/tool/deploy`
- **THEN** `monom` execs that path as a single argv element with no arguments

#### Scenario: monom exits non-zero when mnmd pack fails
- **WHEN** `mnmd pack` exits non-zero (command not found)
- **THEN** `monom` exits non-zero without exec'ing anything
