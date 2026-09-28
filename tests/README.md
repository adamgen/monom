# Tests

`make check` runs everything: `go vet`, Go unit tests, every shUnit2 suite under `tests/`, and shellcheck. `make help` lists the narrower targets.

There are three kinds of tests. Pick by what the test needs:

| The test is… | Write a… | Where |
| --- | --- | --- |
| a CLI line a user types, then Enter or Tab, then what they see | **declarative case** | `tests/cases/*.yaml` |
| anything else observable from outside: exit codes of `mnmd` subcommands, stdin contracts, files written, env vars, function/registration tables | shUnit2 test | `tests/mnmd_<subcommand>_test`, `tests/monom_<area>_test` |
| Go logic not observable from the binary | Go unit test | `*_test.go` next to the code |

Prefer a case whenever the scenario fits the input → action → expected shape.

---

## Declarative cases

`tests/monom_cases_test` reads every `tests/cases/*.yaml` file and turns each case into one shUnit2 test per shell. Each test starts a fresh `bash --norc --noprofile` or `zsh --no-rcs`, `cd`s into the project root, sources `src/monom` like a user's rc file does, and then either runs the line (**enter**) or triggers the real completion function the bindings registered for its first word (**tab**).

```sh
make test-cases                                                 # all case files
CASES=tests/cases/monom_run.yaml bash tests/monom_cases_test    # one file
```

### A case file

```yaml
# Suite config: the project root (relative to the repo root) and the shells
# to run every case in.
root: fixtures/demo-project
shells: [bash, zsh]

cases:
  - name: trailing space drills into a group
    input: "monom infra "
    action: tab
    expect: [cloud, local]

  - name: a group lists its children instead of running anything
    input: monom infra
    action: enter
    exit: 1
    expect: |
      monom: 'infra' is a command group
      available: cloud, local

# Groups start from the suite config and override it for their cases.
groups:
  - name: hooks
    root: fixtures/hooks-project
    cases:
      - name: run hook expands an alias into multiple tokens
        input: monom dbm
        action: enter
        expect: ran db/migrate
```

### Schema

**File (suite) level**

| Key | Required | Meaning |
| --- | --- | --- |
| `root` | yes, unless every group sets one | Project root the cases run in, relative to the repo root. Must be a directory. Put new projects under `fixtures/`. Commands run inside the fixture, so fixtures must not be written to. |
| `shells` | no, default `[bash, zsh]` | Shells to run every case in. Each case becomes one test per shell. |
| `cases` | `cases` and/or `groups` | List of cases that use the suite config. |
| `groups` | `cases` and/or `groups` | List of groups. |

**Group**

| Key | Required | Meaning |
| --- | --- | --- |
| `name` | yes | Shown in validation errors. |
| `root`, `shells` | no | Override the suite config for this group's cases only. |
| `cases` | yes | The group's cases. |

**Case**

| Key | Required | Meaning |
| --- | --- | --- |
| `name` | yes | What the behavior is. Unique within the file; it becomes the test function name. |
| `input` | yes | The CLI line as typed. Quote it when whitespace matters: for tab, `"monom infra "` (drill into `infra`) and `monom infra` (complete the word `infra`) are different cases. |
| `action` | yes | `enter` runs the line. `tab` presses Tab at the end of it. |
| `expect` | yes | For **enter**, a string with what the user sees. Use a `\|` block scalar for several lines; trailing newlines are dropped, and `""` means no output. For **tab**, a list of candidates, with `[]` for none. Candidates are always read as strings, so `[yes, 10]` means the words `yes` and `10`. |
| `exit` | no | enter only. The expected exit status, 0–255. Defaults to `0`, so a command that fails has to say so. |
| `match` | no | `exact` (the default) or `normalized`. |

Validation is strict, and happens before any test runs. Unknown keys, missing required keys, duplicate case names, a bad `action`/`match`/`exit`/shell, a `root` that is not a directory, and an `expect` of the wrong shape each fail the run with `FATAL: <file>:<line>: [group "<g>": ]case "<name>": <problem>`. Every problem in the file is listed, not just the first.

### How it is parsed

The bash runner does not parse YAML itself. It builds `tools/cases` (Go, using `gopkg.in/yaml.v3`, logic in `internal/testcases`), which validates the files and prints one `case_add` line per case, single-quoted for bash. The runner `eval`s that output, then runs each case in real shells. This is test-only tooling: `mnmd` doesn't import it, so building monom needs neither the tool nor the YAML library.

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
  case:   a group lists its children instead of running anything  (tests/cases/monom_run.yaml:24)
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

When the difference is whitespace only, the diff shows spaces as `·` and tabs as `→`.

---

## shUnit2 tests

Shared helpers live in `tests/helpers`, which is sourced by every test file and never executed directly. See `CLAUDE.md` for the file skeleton and naming conventions.
