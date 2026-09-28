# Tests

`make check` runs everything: `go vet`, Go unit tests, every shUnit2 suite under `tests/`, the declarative cases, and shellcheck. `make help` lists the narrower targets.

There are three kinds of tests. Pick by what the test needs:

| The test is… | Write a… | Where |
| --- | --- | --- |
| a CLI line a user types, then Enter or Tab, then what they see | **declarative case** | `tests/cases/*.yaml` |
| anything else observable from outside: exit codes of `mnmd` subcommands, stdin contracts, files written, env vars, function/registration tables | shUnit2 test | `tests/mnmd_<subcommand>_test`, `tests/monom_<area>_test` |
| Go logic not observable from the binary | Go unit test | `*_test.go` next to the code |

Prefer a case whenever the scenario fits the input → action → expected shape.

---

## Declarative cases

`TestCases` (`tests/harness/cases_test.go`) loads every `tests/cases/*.yaml` file and runs each case as a Go subtest per shell, named `TestCases/<file>/<case>/<shell>`. Each subtest starts a fresh `bash --norc --noprofile` or `zsh --no-rcs`, `cd`s into the project root, sources `src/monom` like a user's rc file does, and then either runs the line (**enter**) or triggers the real completion function the bindings registered for its first word (**tab**). **keys** cases go further: they type into a real interactive shell in a pseudo-terminal and press Tab (see [Real key presses](#real-key-presses-action-keys)).

```sh
make test-cases                                          # all case files
CASES=tests/cases/monom_run.yaml make test-cases         # one file (space-separate several)

# Or use go test inside the harness module. -run matches subtest names, with
# spaces as _. CASES paths are relative to the repo root.
cd tests/harness
go test -count=1 -run 'TestCases/monom_run/' .
go test -count=1 -run 'TestCases/monom_run/single-token_command_runs/zsh' -v .
```

A shell that isn't installed is skipped (`t.Skip`), so the cases still run on a machine without zsh.

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

A **keys** case (from `tests/cases/monom_keys.yaml`):

```yaml
  - name: tab completes a unique prefix and adds a space
    input: monom in
    action: keys
    tabs: 1
    line: "monom infra "     # the edit buffer after the Tab presses
    candidates: []           # optional: nothing listed on screen
```

### Schema

**File (suite) level**

| Key | Required | Meaning |
| --- | --- | --- |
| `root` | yes, unless every group sets one | Project root the cases run in, relative to the repo root. Must be a directory. Put new projects under `fixtures/`. Commands run inside the fixture, so fixtures must not be written to. |
| `shells` | no, default `[bash, zsh]` | Shells to run every case in. Each case becomes one test per shell. |
| `env` | no | Environment variables to set in the case's shell, as `NAME: value`. `{root}` in a value is replaced with the absolute project root. See [Environment](#environment). |
| `cases` | `cases` and/or `groups` | List of cases that use the suite config. |
| `groups` | `cases` and/or `groups` | List of groups. |

**Group**

| Key | Required | Meaning |
| --- | --- | --- |
| `name` | yes | Shown in validation errors. |
| `root`, `shells` | no | Override the suite config for this group's cases only. |
| `env` | no | Merged over the suite's `env`, name by name. |
| `cases` | yes | The group's cases. |

**Case**

| Key | Required | Meaning |
| --- | --- | --- |
| `name` | yes | What the behavior is. Unique within the file; it becomes the test function name. |
| `input` | yes | The CLI line as typed. Quote it when whitespace matters: for tab, `"monom infra "` (drill into `infra`) and `monom infra` (complete the word `infra`) are different cases. |
| `action` | yes | `enter` runs the line. `tab` calls the completion function for the end of the line. `keys` types the line into an interactive shell and presses Tab. |
| `expect` | enter, tab | For **enter**, a string with what the user sees. Use a `\|` block scalar for several lines; trailing newlines are dropped, and `""` means no output. For **tab**, a list of candidates, with `[]` for none. Candidates are always read as strings, so `[yes, 10]` means the words `yes` and `10`. |
| `exit` | no | enter only. The expected exit status, 0–255. Defaults to `0`, so a command that fails has to say so. |
| `match` | no | enter and tab. `exact` (the default) or `normalized`. |
| `env` | no | Merged over the group's (or suite's) `env`, name by name. |
| `tabs` | keys | How many times to press Tab after the input: 1–3. |
| `line` | keys | The line editor's buffer afterwards, compared exactly. Quote it to keep a trailing space. |
| `candidates` | no | keys only. The completion listing on screen, compared as a set; `[]` means nothing is listed. Omit it to check only `line`. |

Validation is strict, and happens before any test runs. Unknown keys, missing required keys, duplicate case names, a bad `action`/`match`/`exit`/`tabs`/shell, an `env` that isn't a `NAME: string` mapping, a key that doesn't apply to the action (e.g. `expect` on a keys case), a `root` that is not a directory, and an `expect` of the wrong shape each fail `TestCases` under `invalid case files:`, one `<file>:<line>: [group "<g>": ]case "<name>": <problem>` line per problem. Every problem in every selected file is listed, not just the first.

### Environment

Every case starts from your environment with the monom variables removed (`_MONOM_PROJECT_ROOT`, `_MONOM_USER_CONFIG`, `MONOM_DEBUG_LOG`, `MONOM_ACTIVE`, `MONOM_CHECK_ROOT_CONTENTS`); keys cases start from an almost empty one. `env:` adds variables on top, at the suite, group or case level:

```yaml
groups:
  - name: no monom file
    root: fixtures/zero-config-project
    env:
      _MONOM_PROJECT_ROOT: "{root}"   # pin the root, as an alias would
    cases:
      - name: a fresh project without a monom file only warns and exits 0
        input: mnmd check
        action: enter
        expect: ...
      - name: MONOM_CHECK_ROOT_CONTENTS=error makes root-contents an error
        env:
          MONOM_CHECK_ROOT_CONTENTS: error
        ...
```

Two uses come up:

- **Pinning the root.** Root resolution walks up to the nearest `monom` file and falls back to the git root. A fixture without a `monom` file therefore resolves to *this repository* unless the case pins `_MONOM_PROJECT_ROOT`. Fixtures that have a `monom` file need no pin.
- **User-level settings**, such as `MONOM_CHECK_ROOT_CONTENTS`.

Values are strings and are set literally; `{root}` is the only substitution.

### How it runs

- The harness is its own Go module, `tests/harness` (`github.com/adamgen/monom/tests/harness`), with its own `go.mod`/`go.sum` for its test-only dependencies: `gopkg.in/yaml.v3`, `github.com/creack/pty` and `github.com/hinshun/vt10x`. The root `go.mod` has no dependencies. It doesn't import the root module either; it only builds and runs `bin/mnmd` and `src/monom`.
- `tests/harness/testcases.go` loads and validates the YAML with `gopkg.in/yaml.v3`. If any selected file is invalid, `TestCases` fails before running a single case.
- `tests/harness/cases_test.go` is the runner. It first builds `bin/mnmd`, as the shUnit2 suites do, because `src/monom` calls it. Then it runs each case with `os/exec` in `bash --norc --noprofile -c` or `zsh --no-rcs -c`. It strips `_MONOM_PROJECT_ROOT`, `_MONOM_USER_CONFIG`, `MONOM_DEBUG_LOG`, `MONOM_ACTIVE` and `MONOM_CHECK_ROOT_CONTENTS` from the environment, and each shell run times out after 30 seconds.
- The Tab press is performed by two embedded shell snippets: `tests/harness/testdata/tab.bash`, which shellcheck lints, and `testdata/tab.zsh`. The runner assigns the typed line to `_case_line` before either one runs.
- `keys` cases use the session helper in `tests/harness/pty_test.go` (see above).
- Go never crosses a module boundary with `./...`, so `go build`, `go vet` and `go test ./...` at the root don't touch the harness. `make test-cases` runs `go test -count=1 ./...` inside `tests/harness`: the loader's unit tests and `TestCases`, once each. `-count=1` matters because the shells, `src/monom` and `bin/mnmd` are invisible to Go's test cache; if you run `go test` in the harness yourself, pass it too.
- The Makefile exports `GOWORK=off`. No `go.work` is committed: gopls handles a nested module without one, and a workspace would merge the two modules' dependency resolution. `go.work` is gitignored for local use.
- None of this reaches `mnmd`: building monom needs no module downloads at all.

### Real key presses (action: keys)

A keys case starts a real interactive shell in a 100×24 pseudo-terminal (`github.com/creack/pty`), types `input`, presses Tab `tabs` times, and reads back two things:

- **`line`**: the edit buffer, dumped by a key bound to Ctrl-] (a `bind -x` function in bash, a zle widget in zsh) that doesn't change it.
- **the screen**: rendered by a terminal emulator (`github.com/hinshun/vt10x`), so cursor movement and redraws don't matter. `candidates` are the words in the rows below the prompt line, up to the next prompt; bash redraws the prompt under its listing, zsh keeps the cursor on the line. Column layout is ignored.

The session is set up like a user's shell, with everything that could vary pinned:

- A temporary `HOME` and `TERM=xterm`, `LANG=C.UTF-8`, plus `PATH` and the case's `env`; nothing else from your environment.
- bash: `bash --noprofile --rcfile <tmp>/.bashrc -i` and a private inputrc: no bracketed paste or bell, `show-all-if-ambiguous off` (so the first Tab completes the common prefix and the second lists), no pager or colours.
- zsh: `zsh --no-globalrcs -i` with `ZDOTDIR=<tmp>`: `emulate -R zsh`, `auto_menu auto_list list_ambiguous no_menu_complete`, and bracketed paste off. On an ambiguous word the first Tab lists and the second inserts the first candidate. `compinit -C` loads a compdump that is built once per run.
- The rc then `cd`s into `root` and sources `src/monom`, and the prompt is `PROMPT$ `.

Nothing waits a fixed time. The shells write invisible terminal title sequences as markers: one before every prompt, one inside the prompt so it appears only once the line editor is ready for keys, and one with the line dump. Each wait gives up after 5 seconds and prints the screen and the raw terminal output.

Behavior differs between the shells here, so put shell-specific expectations in a group with `shells: [bash]` or `shells: [zsh]`, as `monom_keys.yaml` does.

### What is compared

- **enter:** stdout and stderr interleaved, which is what a terminal shows, plus the exit status.
- **keys:** `line`, exactly, and `candidates` when given, as a set.
- **tab:** the candidates the completion function offers, one per line. Both sides are sorted before comparing, because display order belongs to the shell. A tab case also fails if completion writes anything to stderr or exits non-zero, since completion must never be noisy mid-typing.
  - In bash, the candidates are `COMPREPLY` after calling the function that `complete -p <word>` names, with `COMP_WORDS`, `COMP_CWORD`, and `COMP_LINE` set the way readline sets them.
  - In zsh, `compinit` runs before sourcing, so `compdef` registers the real function in `$_comps`. `compadd` is replaced by a function that prints the candidates matching the word under the cursor, which is the prefix match real `compadd` applies.

**Matching:**

- **`exact`** compares line by line, character for character. Trailing newlines at the very end of the output are ignored. Nothing else is.
- **`normalized`** trims each line, collapses runs of whitespace to one space, and drops blank lines on both sides first. Use it only when spacing is genuinely not part of the behavior.

### Failures

A failing case prints everything needed to reproduce it, and a unified diff:

```
--- FAIL: TestCases/monom_run/a_group_lists_its_children_instead_of_running_anything/zsh (0.03s)
    cases_test.go:172:
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
