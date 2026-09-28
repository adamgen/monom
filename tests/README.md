# Tests

`make check` runs everything: `go vet`, Go unit tests, every shUnit2 suite under `tests/`, and shellcheck. `make help` lists the narrower targets.

There are three kinds of tests. Pick by what the test needs:

| The test is… | Write a… | Where |
| --- | --- | --- |
| a CLI line a user types, then Enter or Tab, then what they see | **declarative case** | `tests/cases/*.cases` |
| anything else observable from outside: exit codes of `mnmd` subcommands, stdin contracts, files written, env vars, function/registration tables | shUnit2 test | `tests/mnmd_<subcommand>_test`, `tests/monom_<area>_test` |
| Go logic not observable from the binary | Go unit test | `*_test.go` next to the code |

Prefer a case whenever the scenario fits the input → action → expected shape.

---

## Declarative cases

`tests/monom_cases_test` reads every `tests/cases/*.cases` file and turns each case into one shUnit2 test per shell. Each test starts a fresh `bash --norc --noprofile` or `zsh --no-rcs`, `cd`s into the project root, sources `src/monom` like a user's rc file does, and then either runs the line (**enter**) or triggers the real completion function the bindings registered for its first word (**tab**).

```sh
make test-cases                                             # all case files
CASES=tests/cases/monom_run.cases bash tests/monom_cases_test  # one file
```

### A case file

```
# Comments are whole lines starting with "#", allowed anywhere, including
# inside an expect block. There are no end-of-line comments.

# Suite config: the project root, relative to the repo root, and the shells
# to run in (optional; the default is "bash zsh").
@root fixtures/demo-project
@shells bash zsh

-- trailing space drills into a group
input:  "monom infra "
action: tab
expect:
cloud
local

-- a group lists its children instead of running anything
input:  monom infra
action: enter
exit:   1
expect:
monom: 'infra' is a command group
available: cloud, local

# A group starts from the suite config and overrides it for its cases.
== hooks
@root fixtures/hooks-project

-- run hook expands an alias into multiple tokens
input:  monom dbm
action: enter
expect: ran db/migrate
```

**Config.** `@root` and `@shells` lines at the top of the file are the suite config. A `== <name>` line starts a group. `@` lines right after it override the suite config for that group's cases, and each group starts again from the suite config. `@root` must point to a directory; put new projects under `fixtures/`. Commands run inside the fixture, so fixtures must not be written to.

**Cases.** `-- <name>` starts a case. The name becomes the test function name, so it must be unique in the file and should say what the behavior is. Every case has three fields:

| Field | Meaning |
| --- | --- |
| `input:` | The CLI line as typed. Surrounding whitespace is trimmed. Wrap the value in double quotes when whitespace matters: for tab, `"monom infra "` (drill into `infra`) and `monom infra` (complete the word `infra`) are different cases. |
| `action:` | `enter` runs the line. `tab` presses Tab at the end of the line. |
| `expect:` | What the user sees. Write it on the same line for one line of output, or on the lines below for several (a quoted value keeps its whitespace). The block ends at the next `--` or `==` line. Trailing blank lines are dropped, and `expect:` with nothing after it means empty output. To expect a line that starts with `#` or `\`, prefix it with `\`. `expect` must be the case's last field. |

Two optional fields go before `expect:`:

| Field | Meaning |
| --- | --- |
| `exit:` | enter only. The expected exit status. Defaults to `0`, so a command that fails has to say so. |
| `match:` | `exact` (the default) or `normalized`. |

### What is compared

- **enter:** stdout and stderr interleaved, which is what a terminal shows, plus the exit status.
- **tab:** the candidates the completion function offers, one per line. Both sides are sorted before comparing, because display order belongs to the shell. A tab case also fails if completion writes anything to stderr or exits non-zero, since completion must never be noisy mid-typing.
  - In bash, the candidates are `COMPREPLY` after calling the function that `complete -p <word>` names, with `COMP_WORDS`, `COMP_CWORD`, and `COMP_LINE` set the way readline sets them.
  - In zsh, `compinit` runs before sourcing, so `compdef` registers the real function in `$_comps`. `compadd` is replaced by a function that prints the candidates matching the word under the cursor, which is the prefix match real `compadd` applies.

**Matching:**

- **`exact`** compares line by line, character for character. Trailing newlines at the very end of the output are ignored. Nothing else is.
- **`normalized`** trims each line, collapses runs of whitespace to one space, and drops blank lines on both sides first. Use it only when spacing is genuinely not part of the behavior.

### Failures

A failing case prints everything needed to reproduce it, and a unified diff:

```
ASSERT:
  case:   a group lists its children instead of running anything  (tests/cases/monom_run.cases:23)
  shell:  zsh    root: fixtures/demo-project
  input:  "monom infra"
  action: enter    match: exact
  result: output differs; exit code 1, expected 0
    --- expected
    +++ actual
    @@ -1 +1,2 @@
    -monom: 'infra' is a group
    +monom: 'infra' is a command group
    +available: cloud, local
```

When the difference is whitespace only, the diff shows spaces as `·` and tabs as `→`. Malformed case files fail the whole run with `FATAL: <file>:<line>: <reason>` before any test runs.

---

## shUnit2 tests

Shared helpers live in `tests/helpers`, which is sourced by every test file and never executed directly. See `CLAUDE.md` for the file skeleton and naming conventions.
