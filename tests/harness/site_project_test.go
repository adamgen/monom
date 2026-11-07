package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestSiteProject checks what monom.dev's homepage claims against the real
// mnmd and monom. It rebuilds the site's example project ("acme",
// site/app/lib/project.json) from fixtures/site-project in a temp git repo,
// compiles its two binaries from source (tools/lint in Go, tools/fmt in
// Rust, testdata/site-binaries), and asserts the site's trees, outputs and
// snippets in bash and zsh. Snippets that live as constants in
// site/app/routes/_index.tsx are read from that file, so they can't drift.
//
// Every interpreter the example uses (python3, node, ruby, rustc) is
// required when CI is set; locally a missing one skips only its checks.
//
//	cd tests/harness && go test -count=1 -run TestSiteProject -v .
func TestSiteProject(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", filepath.Join(repo, "bin", "mnmd"), "./cmd/mnmd")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("could not build mnmd: %v\n%s", err, out)
	}

	entries := siteEntries(t, repo)
	index := readFile(t, filepath.Join(repo, "site/app/routes/_index.tsx"))
	have := tools(t, "python3", "node", "ruby", "rustc", "git")
	if !have["git"] {
		t.Skip("git not available")
	}
	root := newSiteProject(t, repo, have)

	var commands, warnings []string
	for _, e := range entries {
		switch e.Outcome {
		case "command":
			commands = append(commands, e.Command)
		case "warn":
			warnings = append(warnings, fmt.Sprintf("warning: [root-contents] executable not registered: %s (no shebang, and the name does not match ^[a-z0-9][a-z0-9_-]*$); add a shebang, rename it, or declare it in the monom config file", e.Path))
		}
	}
	sort.Strings(commands)
	runnable := func(e siteEntry) bool {
		need := map[string]string{"python": "python3", "node": "node", "ruby": "ruby", "rust": "rustc"}[e.Lang]
		return need == "" || have[need]
	}
	all := true
	for _, e := range entries {
		if e.Outcome == "command" && !runnable(e) {
			all = false
		}
	}

	for _, sh := range []string{"bash", "zsh"} {
		t.Run(sh, func(t *testing.T) {
			if _, err := exec.LookPath(sh); err != nil {
				if os.Getenv("CI") != "" {
					t.Fatalf("%s is required in CI", sh)
				}
				t.Skipf("%s not available", sh)
			}
			s := siteShell{t: t, repo: repo, root: root, sh: sh}

			// lib/project.ts DISCOVER and CHECK (hero demo, "keep it honest in CI").
			s.expect("mnmd discover", 0, strings.Join(commands, "\n"))
			s.expect("mnmd check", 0, strings.Join(append(warnings,
				fmt.Sprintf("✔ %d commands OK (default discovery), %d warning(s)", len(commands), len(warnings))), "\n"))

			// Every command runs by name, and by file path ("No lock-in").
			for _, e := range entries {
				if e.Outcome != "command" || !runnable(e) {
					continue
				}
				s.expect("monom "+strings.ReplaceAll(e.Command, "/", " "), 0, e.Output)
				s.expect("./"+e.Path, 0, e.Output)
			}

			// "Folders are nouns": GROUP.
			s.expect("monom infra", 1, strings.TrimPrefix(tsConst(t, index, "GROUP"), "$ monom infra\n"))

			// "The repo's front door": monom <Tab>, monom db <Tab>.
			s.tab("monom ", "db", "infra", "release", "tools")
			s.tab("monom db ", "migrate", "seed")

			// "Absent": mnmd root from a subfolder prints the git root.
			s.expect("cd infra/cloud && mnmd root", 0, root)

			// "Go owns logic, shell owns surface".
			if sh == "bash" {
				s.expect("type -t monom mnmd", 0, "function\nfunction")
			}

			// "Completion never breaks mid-typing".
			s.expect("echo 'bad path/x' | mnmd filter bad; echo $?", 0, "0")

			// CHECK_ERROR: the snippet's own first line writes the monom file.
			s.withMonom("check.root-contents = error\n", func() {
				lines := strings.Split(tsConst(t, index, "CHECK_ERROR"), "\n")
				if lines[0] != "$ echo 'check.root-contents = error' > monom" || lines[1] != "$ mnmd check" {
					t.Fatalf("CHECK_ERROR changed shape: %q", lines[:2])
				}
				s.expect("mnmd check", 1, strings.Join(lines[2:len(lines)-2], "\n"))
			})

			// DECLARATIVE: the declared tools/Format.sh registers (by its
			// path), so check passes even at error severity, and
			// discover.hide drops infra/local.
			s.withMonom(tsConst(t, index, "DECLARATIVE")+"\n", func() {
				var want []string
				for _, c := range append(append([]string{}, commands...), "tools/Format.sh") {
					if !strings.HasPrefix(c, "infra/local/") {
						want = append(want, c)
					}
				}
				sort.Strings(want)
				s.expect("mnmd discover", 0, strings.Join(want, "\n"))
				s.expect("mnmd check >/dev/null", 0, "")
			})

			// HOOK_SCRIPT: `monom ship` aliases infra cloud deploy, and the
			// config hook hides tools/.
			s.withMonom(tsConst(t, index, "HOOK_SCRIPT")+"\n", func() {
				s.expect("monom ship", 0, "deploying to cloud...")
				var kept []string
				for _, c := range commands {
					if !strings.HasPrefix(c, "tools/") {
						kept = append(kept, c)
					}
				}
				s.expect("mnmd discover", 0, strings.Join(kept, "\n"))
			})

			// "Your code can own the list": with the complete hook snippet,
			// Tab offers exactly the tracked executables (file names as
			// tracked, _lib included), not default discovery's list.
			s.withMonom("#!/usr/bin/env bash\ncase \"$1\" in\n"+completeSnippet+"\nesac\n", func() {
				s.expect("git add -A", 0, "")
				s.tab("monom ", "_lib", "db", "infra", "release.rb", "tools")
				s.tab("monom db ", "migrate.py", "seed.js")
			})
		})
	}
	if !all {
		t.Log("some interpreters are missing, so their commands were not run; CI runs all of them")
	}
}

// completeSnippet is the "Your code can own the list" snippet in _index.tsx,
// verbatim.
const completeSnippet = `  complete)
    cd "$(dirname "$0")" || exit
    # tracked executables, minus this file
    git ls-files -s | awk '$1 == 100755 { print $4 }' | grep -vx monom
    ;;`

type siteEntry struct {
	Path, Command, Outcome, Lang, Why, Output string
}

func siteEntries(t *testing.T, repo string) []siteEntry {
	var entries []siteEntry
	if err := json.Unmarshal([]byte(readFile(t, filepath.Join(repo, "site/app/lib/project.json"))), &entries); err != nil {
		t.Fatalf("site/app/lib/project.json: %v", err)
	}
	return entries
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// tsConst returns a backtick string constant from _index.tsx; it must not
// interpolate anything.
func tsConst(t *testing.T, src, name string) string {
	t.Helper()
	m := regexp.MustCompile("(?s)\nconst " + name + " = `(.*?)`;").FindStringSubmatch(src)
	if m == nil || strings.Contains(m[1], "${") {
		t.Fatalf("_index.tsx: no plain constant %s", name)
	}
	return m[1]
}

func tools(t *testing.T, names ...string) map[string]bool {
	have := map[string]bool{}
	for _, n := range names {
		if _, err := exec.LookPath(n); err == nil {
			have[n] = true
		} else if os.Getenv("CI") != "" {
			t.Fatalf("%s is required in CI: the site's example project uses it", n)
		} else {
			t.Logf("%s not available: skipping what needs it", n)
		}
	}
	return have
}

// newSiteProject copies the fixture into a temp git repo, adds what git
// can't carry (node_modules/), and compiles the Go and Rust binaries.
func newSiteProject(t *testing.T, repo string, have map[string]bool) string {
	root := realTempDir(t)
	src := filepath.Join(repo, "fixtures/site-project")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		dst := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, info.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "node_modules/pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "node_modules/pkg/cli"), "#!/usr/bin/env bash\necho never\n")
	if err := os.Chmod(filepath.Join(root, "node_modules/pkg/cli"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	bins := filepath.Join(repo, "tests/harness/testdata/site-binaries")
	run(bins, "go", "build", "-o", filepath.Join(root, "tools/lint"), "lint.go")
	if have["rustc"] {
		run(bins, "rustc", "-O", "-o", filepath.Join(root, "tools/fmt"), "fmt.rs")
	}
	writeFile(t, filepath.Join(root, ".gitignore"), "node_modules/\n")
	run(root, "git", "init", "-q")
	return root
}

type siteShell struct {
	t              *testing.T
	repo, root, sh string
}

func (s siteShell) run(script string) (string, int) {
	s.t.Helper()
	var b strings.Builder
	fmt.Fprintf(&b, "cd %s || exit 97\n", shQuote(s.root))
	if s.sh == "zsh" {
		b.WriteString("autoload -Uz compinit && compinit -D -u\n")
	}
	fmt.Fprintf(&b, "source %s\n%s\n", shQuote(filepath.Join(s.repo, "src", "monom")), script)
	args := []string{"--norc", "--noprofile", "-c", b.String()}
	if s.sh == "zsh" {
		args = []string{"--no-rcs", "-c", b.String()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.sh, args...)
	cmd.Env = append(cleanEnv(s.repo),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	status := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			s.t.Fatalf("could not run %s: %v", s.sh, err)
		}
		status = exitErr.ExitCode()
	}
	return strings.TrimRight(out.String(), "\n"), status
}

func (s siteShell) expect(line string, exit int, want string) {
	s.t.Helper()
	got, status := s.run(line)
	if got != want || status != exit {
		s.t.Errorf("%s: %s\n  exit %d, want %d\n%s", s.sh, line, status, exit, strings.Join(unifiedDiff(want, got), "\n"))
	}
}

func (s siteShell) tab(line string, want ...string) {
	s.t.Helper()
	tab := bashTab
	if s.sh == "zsh" {
		tab = zshTab
	}
	got, status := s.run("_case_line=" + shQuote(line) + "\n" + tab)
	lines := strings.Split(got, "\n")
	sort.Strings(lines)
	if strings.Join(lines, " ") != strings.Join(want, " ") || status != 0 {
		s.t.Errorf("%s: %q <Tab>: got %q (exit %d), want %q", s.sh, line, lines, status, want)
	}
}

// withMonom runs fn with a monom file at the project root, then removes it.
func (s siteShell) withMonom(content string, fn func()) {
	s.t.Helper()
	p := filepath.Join(s.root, "monom")
	mode := os.FileMode(0o644)
	if strings.HasPrefix(content, "#!") {
		mode = 0o755
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		s.t.Fatal(err)
	}
	if err := os.Chmod(p, mode); err != nil {
		s.t.Fatal(err)
	}
	defer os.Remove(p)
	fn()
}
