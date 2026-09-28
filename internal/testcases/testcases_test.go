package testcases

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// load writes content to <repo>/cases.yaml, with <repo>/fixture as a valid
// root, and loads it.
func load(t *testing.T, content string) ([]Case, error) {
	t.Helper()
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, "fixture"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(repo, "cases.yaml")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(file, repo)
}

// wantErrors asserts that loading fails and every fragment appears in some
// error line.
func wantErrors(t *testing.T, content string, fragments ...string) {
	t.Helper()
	_, err := load(t, content)
	var errs Errors
	if !errors.As(err, &errs) {
		t.Fatalf("want Errors, got %v", err)
	}
	all := errs.Error()
	for _, f := range fragments {
		if !strings.Contains(all, f) {
			t.Errorf("missing %q in errors:\n%s", f, all)
		}
	}
}

func TestLoad_ResolvesSuiteAndGroupConfig(t *testing.T) {
	cases, err := load(t, `
root: fixture
cases:
  - name: tab case
    input: "monom "
    action: tab
    expect: [b, a]
groups:
  - name: g
    shells: [zsh]
    cases:
      - name: enter case
        input: monom x
        action: enter
        exit: 3
        match: normalized
        expect: |
          line one
          line two
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("got %d cases", len(cases))
	}
	tab, enter := cases[0], cases[1]
	if tab.Input != "monom " || tab.Expect != "b\na" || tab.Match != MatchExact || !reflect.DeepEqual(tab.Shells, []string{"bash", "zsh"}) || tab.Line != 4 {
		t.Errorf("tab case = %+v", tab)
	}
	if enter.Group != "g" || enter.Root != "fixture" || enter.Exit != 3 || enter.Match != MatchNormalized ||
		enter.Expect != "line one\nline two" || !reflect.DeepEqual(enter.Shells, []string{"zsh"}) {
		t.Errorf("enter case = %+v", enter)
	}
}

func TestLoad_EmptyTabListAndEmptyEnterStringMeanNoOutput(t *testing.T) {
	cases, err := load(t, `
root: fixture
cases:
  - {name: a, input: x, action: tab, expect: []}
  - {name: b, input: x, action: enter, expect: ""}
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cases[0].Expect != "" || cases[1].Expect != "" {
		t.Errorf("got %+v", cases)
	}
}

func TestLoad_CandidatesAreStringsEvenWhenYAMLWouldCoerceThem(t *testing.T) {
	cases, err := load(t, `
root: fixture
cases:
  - {name: a, input: x, action: tab, expect: [yes, 2fa, 10, null]}
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cases[0].Expect != "yes\n2fa\n10\nnull" {
		t.Errorf("expect = %q", cases[0].Expect)
	}
}

func TestLoad_UnknownKeysAreErrorsWithContext(t *testing.T) {
	wantErrors(t, `
root: fixture
color: blue
cases:
  - name: typo
    input: x
    action: tab
    expct: []
groups:
  - name: g
    rooot: fixture
    cases: []
`,
		`cases.yaml:3: unknown key "color"`,
		`cases.yaml:8: case "typo": unknown key "expct"`,
		`cases.yaml:11: group "g": unknown key "rooot"`,
		`case "typo": missing required key "expect"`,
	)
}

func TestLoad_MissingFieldsAreErrors(t *testing.T) {
	wantErrors(t, `
root: fixture
cases:
  - name: bare
  - input: x
    action: enter
    expect: ""
`,
		`case "bare": missing required key "input"`,
		`case "bare": missing required key "action"`,
		`cases.yaml:5: case: missing required key "name"`,
	)
}

func TestLoad_BadValuesAreErrors(t *testing.T) {
	wantErrors(t, `
root: fixture
shells: [bash, fish]
cases:
  - {name: a, input: x, action: run, expect: ""}
  - {name: b, input: x, action: tab, exit: 1, expect: []}
  - {name: c, input: x, action: enter, exit: one, expect: ""}
  - {name: d, input: x, action: enter, match: fuzzy, expect: ""}
  - {name: e, input: x, action: tab, expect: "not a list"}
  - {name: f, input: x, action: enter, expect: [not, a, string]}
  - {name: a, input: x, action: enter, expect: ""}
`,
		`unknown shell "fish"`,
		`case "a": action must be "enter", "tab" or "keys", got "run"`,
		`case "b": exit applies only to action: enter`,
		`case "c": exit must be an integer`,
		`case "d": match must be "exact" or "normalized"`,
		`case "e": expect for action: tab must be a list`,
		`case "f": expect for action: enter must be a string`,
		`case "a": duplicate case name (first defined on line 5)`,
	)
}

func TestLoad_RootMustBeADirectoryAndIsRequired(t *testing.T) {
	wantErrors(t, `
cases:
  - {name: a, input: x, action: enter, expect: ""}
groups:
  - name: g
    root: missing
    cases:
      - {name: b, input: x, action: enter, expect: ""}
`,
		`case "a": no root`,
		`group "g": root "missing" is not a directory`,
	)
}

func TestLoad_InvalidYAMLAndEmptyFiles(t *testing.T) {
	wantErrors(t, "root: [unclosed\n", "invalid YAML")
	wantErrors(t, "", "empty file")
	wantErrors(t, "root: fixture\n", "no cases")
}

func TestLoad_KeysCase(t *testing.T) {
	cases, err := load(t, `
root: fixture
cases:
  - name: with listing
    input: "monom infra "
    action: keys
    tabs: 2
    line: "monom infra "
    candidates: [local, cloud]
  - {name: line only, input: monom in, action: keys, tabs: 1, line: "monom infra "}
  - {name: no listing, input: monom zz, action: keys, tabs: 1, line: monom zz, candidates: []}
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a, b, c := cases[0], cases[1], cases[2]
	if a.Tabs != 2 || a.WantLine != "monom infra " || !a.CheckCandidates || !reflect.DeepEqual(a.Candidates, []string{"local", "cloud"}) {
		t.Errorf("with listing = %+v", a)
	}
	if b.Tabs != 1 || b.WantLine != "monom infra " || b.CheckCandidates {
		t.Errorf("line only = %+v", b)
	}
	if !c.CheckCandidates || len(c.Candidates) != 0 {
		t.Errorf("no listing = %+v", c)
	}
}

func TestLoad_KeysFieldsAreValidated(t *testing.T) {
	wantErrors(t, `
root: fixture
cases:
  - {name: a, input: x, action: keys, line: x}
  - {name: b, input: x, action: keys, tabs: 4, line: x}
  - {name: c, input: x, action: keys, tabs: 1}
  - {name: d, input: x, action: keys, tabs: 1, line: [x]}
  - {name: e, input: x, action: keys, tabs: 1, line: x, candidates: x}
  - {name: f, input: x, action: keys, tabs: 1, line: x, expect: [], exit: 1, match: exact}
  - {name: g, input: x, action: tab, expect: [], tabs: 1, line: x, candidates: []}
  - {name: h, input: "x\ty", action: keys, tabs: 1, line: x}
`,
		`case "a": missing required key "tabs"`,
		`case "b": tabs must be an integer from 1 to 3, got "4"`,
		`case "c": missing required key "line"`,
		`case "d": line must be a string`,
		`case "e": candidates must be a list`,
		`case "f": expect does not apply to action: keys`,
		`case "f": exit does not apply to action: keys`,
		`case "f": match does not apply to action: keys`,
		`case "g": tabs applies only to action: keys`,
		`case "g": line applies only to action: keys`,
		`case "g": candidates applies only to action: keys`,
		`case "h": input for action: keys must be one line without tabs`,
	)
}
