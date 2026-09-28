// TestCases runs the declarative CLI cases in tests/cases/*.yaml through the
// real shell integration: one subtest per file/case/shell, each in a fresh
// bash or zsh that cds into the case root and sources src/monom like a user's
// rc file does. action: keys cases run in an interactive shell in a
// pseudo-terminal instead (pty_test.go).
//
// This is its own Go module, so the root module's `go test ./...` never
// reaches it. From the repo root:
//
//	make test-cases                                        # every file
//	CASES=tests/cases/monom_run.yaml make test-cases       # one file
//	cd tests/harness && go test -count=1 -run 'TestCases/monom_run/' .
package harness

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

var (
	//go:embed testdata/tab.bash
	bashTab string
	//go:embed testdata/tab.zsh
	zshTab string
)

// caseTimeout bounds one shell run, so a hanging case fails instead of
// stalling the suite.
const caseTimeout = 30 * time.Second

// scrubbedEnv lists variables a user's session may set that would change
// monom's behavior; every case runs without them.
var scrubbedEnv = []string{"_MONOM_PROJECT_ROOT", "_MONOM_USER_CONFIG", "MONOM_DEBUG_LOG", "MONOM_ACTIVE", "MONOM_CHECK_ROOT_CONTENTS"}

func TestCases(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	// Same as the shUnit2 suites: compile bin/mnmd, which src/monom calls.
	build := exec.Command("go", "build", "-o", filepath.Join(repo, "bin", "mnmd"), "./cmd/mnmd")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("could not build mnmd: %v\n%s", err, out)
	}

	files := caseFiles(t, repo)
	loaded := make([][]Case, len(files))
	var problems []string
	for i, f := range files {
		cases, err := Load(f, repo)
		var errs Errors
		switch {
		case errors.As(err, &errs):
			problems = append(problems, errs...)
		case err != nil:
			problems = append(problems, err.Error())
		}
		loaded[i] = cases
	}
	// Any schema problem stops the run before a single case executes.
	if len(problems) > 0 {
		t.Fatalf("invalid case files:\n  %s", strings.Join(problems, "\n  "))
	}

	// action: keys cases in zsh share one pre-built compdump.
	compdump := ""
	if _, err := exec.LookPath("zsh"); err == nil && needsCompdump(loaded) {
		compdump = buildCompdump(t, t.TempDir())
	}

	for i, f := range files {
		t.Run(strings.TrimSuffix(filepath.Base(f), ".yaml"), func(t *testing.T) {
			for _, c := range loaded[i] {
				t.Run(c.Name, func(t *testing.T) {
					for _, sh := range c.Shells {
						t.Run(sh, func(t *testing.T) {
							if c.Action == ActionKeys {
								if _, err := exec.LookPath(sh); err != nil {
									t.Skipf("%s not available", sh)
								}
								runKeysCase(t, repo, compdump, c, sh)
								return
							}
							runCase(t, repo, c, sh)
						})
					}
				})
			}
		})
	}
}

func needsCompdump(loaded [][]Case) bool {
	for _, cases := range loaded {
		for _, c := range cases {
			if c.Action == ActionKeys && contains(c.Shells, "zsh") {
				return true
			}
		}
	}
	return false
}

// caseFiles returns the files to run, relative to repo: $CASES (space
// separated) when set, otherwise every tests/cases/*.yaml.
func caseFiles(t *testing.T, repo string) []string {
	t.Helper()
	if env := os.Getenv("CASES"); env != "" {
		var files []string
		for _, f := range strings.Fields(env) {
			if rel, err := filepath.Rel(repo, f); filepath.IsAbs(f) && err == nil {
				f = rel
			}
			files = append(files, f)
		}
		return files
	}
	matches, err := filepath.Glob(filepath.Join(repo, "tests", "cases", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no case files in tests/cases")
	}
	files := make([]string, len(matches))
	for i, m := range matches {
		files[i], _ = filepath.Rel(repo, m)
	}
	return files
}

func runCase(t *testing.T, repo string, c Case, sh string) {
	shellPath, err := exec.LookPath(sh)
	if err != nil {
		t.Skipf("%s not available", sh)
	}

	ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
	defer cancel()
	args := []string{"--norc", "--noprofile", "-c", caseScript(repo, c, sh)}
	if sh == "zsh" {
		args = []string{"--no-rcs", "-c", caseScript(repo, c, sh)}
	}
	cmd := exec.CommandContext(ctx, shellPath, args...)
	cmd.Env = append(cleanEnv(), caseEnv(repo, c)...)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	if c.Action == ActionEnter {
		cmd.Stderr = &out // what the user sees: stdout and stderr interleaved
	} else {
		cmd.Stderr = &stderr
	}
	status := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("could not run %s: %v", sh, err)
		}
		status = exitErr.ExitCode()
	}

	want := normalize(c.Expect, c.Match, c.Action)
	got := normalize(out.String(), c.Match, c.Action)
	errText := strings.TrimRight(stderr.String(), "\n")

	var problems []string
	if want != got {
		problems = append(problems, "output differs")
	}
	if ctx.Err() != nil {
		problems = append(problems, fmt.Sprintf("timed out after %s", caseTimeout))
	}
	if c.Action == ActionEnter && status != c.Exit {
		problems = append(problems, fmt.Sprintf("exit code %d, expected %d", status, c.Exit))
	}
	if c.Action == ActionTab {
		if status != 0 {
			problems = append(problems, fmt.Sprintf("completion exited %d", status))
		}
		if errText != "" {
			problems = append(problems, "completion wrote to stderr")
		}
	}
	if len(problems) > 0 {
		t.Error(report(c, sh, strings.Join(problems, "; "), want, got, errText))
	}
}

// caseScript builds what a case runs inside a fresh shell: cd into the
// project root, source src/monom exactly as a user's rc file does, then act.
func caseScript(repo string, c Case, sh string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "cd %s || exit 97\n", shQuote(filepath.Join(repo, c.Root)))
	if sh == "zsh" && c.Action == ActionTab {
		// compinit before sourcing, the standard .zshrc order, so compdef
		// registers the real completion functions in $_comps.
		b.WriteString("autoload -Uz compinit && compinit -D -u\n")
	}
	fmt.Fprintf(&b, "source %s\n", shQuote(filepath.Join(repo, "src", "monom")))
	if c.Action == ActionEnter {
		b.WriteString(c.Input + "\n")
		return b.String()
	}
	fmt.Fprintf(&b, "_case_line=%s\n", shQuote(c.Input))
	if sh == "zsh" {
		b.WriteString(zshTab)
	} else {
		b.WriteString(bashTab)
	}
	return b.String()
}

// caseEnv returns the case's env as NAME=value, with "{root}" replaced by
// the absolute project root, sorted for stable reports.
func caseEnv(repo string, c Case) []string {
	abs := filepath.Join(repo, c.Root)
	var env []string
	for k, v := range c.Env {
		env = append(env, k+"="+strings.ReplaceAll(v, "{root}", abs))
	}
	sort.Strings(env)
	return env
}

// envLine renders the case's env for failure reports, unexpanded.
func envLine(c Case) string {
	if len(c.Env) == 0 {
		return ""
	}
	var kv []string
	for k, v := range c.Env {
		kv = append(kv, k+"="+v)
	}
	sort.Strings(kv)
	return "  env:    " + strings.Join(kv, " ") + "\n"
}

func cleanEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if !contains(scrubbedEnv, name) {
			env = append(env, kv)
		}
	}
	return env
}

var spaceRun = regexp.MustCompile(`[[:space:]]+`)

// normalize returns the form of an output that is compared. Trailing
// newlines are always dropped. normalized: whitespace runs collapse to one
// space, lines are trimmed, blank lines dropped. tab: candidates are sorted,
// because display order belongs to the shell.
func normalize(s, match, action string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if match == MatchNormalized {
		kept := lines[:0]
		for _, l := range lines {
			if l = strings.TrimSpace(spaceRun.ReplaceAllString(l, " ")); l != "" {
				kept = append(kept, l)
			}
		}
		lines = kept
	}
	if action == ActionTab {
		sort.Strings(lines) // byte order, like LC_ALL=C sort
	}
	return strings.Join(lines, "\n")
}

// report renders a failing case with everything needed to reproduce it.
func report(c Case, sh, problems, want, got, stderr string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n  case:   %s  (%s:%d)\n", c.Name, c.File, c.Line)
	fmt.Fprintf(&b, "  shell:  %s    root: %s\n", sh, c.Root)
	b.WriteString(envLine(c))
	fmt.Fprintf(&b, "  input:  %q\n", c.Input)
	sorted := ""
	if c.Action == ActionTab {
		sorted = " (candidates sorted)"
	}
	fmt.Fprintf(&b, "  action: %s    match: %s%s\n", c.Action, c.Match, sorted)
	fmt.Fprintf(&b, "  result: %s\n", problems)
	if want != got && normalize(want, MatchNormalized, c.Action) == normalize(got, MatchNormalized, c.Action) {
		b.WriteString("  note:   differs only in whitespace (shown as · for space, → for tab)\n")
		mark := strings.NewReplacer(" ", "·", "\t", "→")
		want, got = mark.Replace(want), mark.Replace(got)
	}
	if want != got {
		for _, l := range unifiedDiff(want, got) {
			b.WriteString("    " + l + "\n")
		}
	}
	if stderr != "" {
		b.WriteString("  stderr:\n")
		for _, l := range strings.Split(stderr, "\n") {
			b.WriteString("    " + l + "\n")
		}
	}
	return b.String()
}

// unifiedDiff returns a unified diff of want against got as a single hunk
// with full context; case outputs are short.
func unifiedDiff(want, got string) []string {
	a, b := splitLines(want), splitLines(got)
	// lcs[i][j] is the length of the longest common subsequence of a[i:], b[j:].
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	out := []string{"--- expected", "+++ actual", fmt.Sprintf("@@ -%s +%s @@", hunkRange(len(a)), hunkRange(len(b)))}
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b) && a[i] == b[j]:
			out = append(out, " "+a[i])
			i, j = i+1, j+1
		case i < len(a) && (j == len(b) || lcs[i+1][j] >= lcs[i][j+1]):
			out = append(out, "-"+a[i])
			i++
		default:
			out = append(out, "+"+b[j])
			j++
		}
	}
	return out
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// hunkRange formats a hunk side the way diff -u does.
func hunkRange(n int) string {
	switch n {
	case 0:
		return "0,0"
	case 1:
		return "1"
	}
	return fmt.Sprintf("1,%d", n)
}

// shQuote single-quotes s for bash and zsh.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
