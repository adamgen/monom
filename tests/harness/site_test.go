package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestSiteExamples keeps monom.dev's example data honest without running
// anything: site/app/lib/project.json and examples.json must match
// fixtures/site-project and the expectations in tests/cases/site_homepage.yaml,
// which TestCases runs for real. It runs no shell.
func TestSiteExamples(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := Load(filepath.Join(repo, "tests/cases/site_homepage.yaml"), repo)
	if err != nil {
		t.Fatal(err)
	}
	// Enter cases at the suite level, by input; at most one per input.
	byInput := map[string]Case{}
	byName := map[string]Case{}
	for _, c := range cases {
		byName[c.Name] = c
		if c.Action == ActionEnter && c.Group == "" {
			byInput[c.Input] = c
		}
	}

	var entries []struct{ Path, Command, Outcome, Output string }
	readJSON(t, filepath.Join(repo, "site/app/lib/project.json"), &entries)
	var examples struct{ Group, CheckError string }
	readJSON(t, filepath.Join(repo, "site/app/lib/examples.json"), &examples)

	var commands, warnings []string
	for _, e := range entries {
		file := filepath.Join(repo, "fixtures/site-project", e.Path)
		info, err := os.Stat(file)
		built := e.Path == "tools/lint" || e.Path == "tools/fmt" // compiled by TestCases
		switch {
		case built:
		case err != nil:
			t.Errorf("project.json lists %s, missing from fixtures/site-project", e.Path)
			continue
		case !info.IsDir() && (info.Mode()&0o111 != 0) != (e.Outcome != "ignored"):
			t.Errorf("%s: executable bit does not match outcome %q", e.Path, e.Outcome)
		}
		switch e.Outcome {
		case "command":
			commands = append(commands, e.Command)
			input := "monom " + strings.ReplaceAll(e.Command, "/", " ")
			if c, ok := byInput[input]; !ok || c.Expect != e.Output || c.Exit != 0 {
				t.Errorf("site_homepage.yaml needs an enter case %q expecting %q", input, e.Output)
			}
		case "warn":
			warnings = append(warnings, fmt.Sprintf("warning: [root-contents] executable not registered: %s (no shebang, and the name does not match ^[a-z0-9][a-z0-9_-]*$); add a shebang, rename it, or declare it in the monom config file", e.Path))
		}
	}
	sort.Strings(commands)

	// The same strings site/app/lib/project.ts builds for DISCOVER and CHECK.
	expect := map[string]string{
		"mnmd discover": strings.Join(commands, "\n"),
		"mnmd check": strings.Join(append(warnings,
			fmt.Sprintf("✔ %d commands OK (default discovery), %d warning(s)", len(commands), len(warnings))), "\n"),
		"monom infra": examples.Group,
	}
	for input, want := range expect {
		if c := byInput[input]; c.Expect != want {
			t.Errorf("site_homepage.yaml %q expects:\n%s\nthe site shows:\n%s", input, c.Expect, want)
		}
	}
	if c := byName["one config line makes check fail"]; c.Expect != examples.CheckError || c.Exit != 1 {
		t.Errorf("site_homepage.yaml check-error case expects:\n%s\nthe site shows:\n%s", c.Expect, examples.CheckError)
	}
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
