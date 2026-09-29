// install.sh, run the way a user runs it: piped into bash, in an empty HOME,
// against release tarballs that build.sh produced and a local HTTP server
// serves (MONOM_DOWNLOAD_BASE). PATH is a directory of symlinks to the handful
// of tools the installer and the shell integration use, and no Go.
package harness

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const installTestVersion = "v0.0.0-installtest"

// installTools is everything install.sh, src/monom and the checks below call.
// go is deliberately absent.
var installTools = []string{
	"bash", "zsh", "sh", "curl", "tar", "gzip", "mkdir", "mktemp", "rm", "mv",
	"cp", "ln", "chmod", "cat", "sed", "awk", "grep", "uname", "tr", "cut", "find",
	"head", "tail", "dirname", "basename", "sha256sum", "shasum", "date", "wc",
	"sort", "env", "ls", "touch",
}

type installEnv struct {
	t        *testing.T
	repo     string
	script   []byte
	tools    string // PATH without go
	release  string // URL serving the release assets
	empty    string // URL with no assets (a repo with no release)
	compdump string
}

func TestInstallScript(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("install.sh supports linux and darwin")
	}
	for _, tool := range []string{"bash", "zsh", "curl"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available", tool)
		}
	}
	e := newInstallEnv(t)

	for _, sh := range []string{"bash", "zsh"} {
		t.Run("release/"+sh, func(t *testing.T) { e.testReleaseInstall(t, sh) })
	}
	t.Run("release/checksum-mismatch-aborts", e.testChecksumMismatch)
	t.Run("no-release/fails-clearly", e.testNoRelease)
}

func newInstallEnv(t *testing.T) *installEnv {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join(repo, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}

	dist := t.TempDir()
	build := exec.Command("bash", "build.sh", "dist", dist)
	build.Dir = repo
	build.Env = append(os.Environ(), "VERSION="+installTestVersion, "TARGETS="+runtime.GOOS+"/"+runtime.GOARCH)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build.sh dist: %v\n%s", err, out)
	}

	release := httptest.NewServer(http.FileServer(http.Dir(dist)))
	empty := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(release.Close)
	t.Cleanup(empty.Close)

	return &installEnv{
		t: t, repo: repo, script: script,
		tools:    toolDir(t, installTools),
		release:  release.URL,
		empty:    empty.URL,
		compdump: buildCompdump(t, t.TempDir()),
	}
}

// toolDir symlinks each available tool into a fresh directory, for a PATH that
// has exactly those. bash is the system one when there is one: it is what
// `curl ... | bash` runs on macOS (3.2).
func toolDir(t *testing.T, tools []string) string {
	dir := t.TempDir()
	for _, tool := range tools {
		p, err := exec.LookPath(tool)
		if tool == "bash" {
			if _, statErr := os.Stat("/bin/bash"); statErr == nil {
				p, err = "/bin/bash", nil
			}
		}
		if err != nil {
			continue
		}
		if err := os.Symlink(p, filepath.Join(dir, tool)); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

type installRun struct {
	out, errOut string
	code        int
}

// install pipes install.sh into bash with exactly env (plus HOME and PATH).
func (e *installEnv) install(t *testing.T, home, path string, env ...string) installRun {
	t.Helper()
	cmd := exec.Command(filepath.Join(e.tools, "bash"))
	cmd.Stdin = strings.NewReader(string(e.script))
	cmd.Dir = home
	cmd.Env = append([]string{"HOME=" + home, "PATH=" + path, "LANG=C"}, env...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := installRun{out: stdout.String(), errOut: stderr.String()}
	if exit, ok := err.(*exec.ExitError); ok {
		r.code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("running install.sh: %v", err)
	}
	t.Logf("install.sh exit %d\nstdout:\n%s\nstderr:\n%s", r.code, r.out, r.errOut)
	return r
}

func (e *installEnv) mustInstall(t *testing.T, home, path string, env ...string) installRun {
	t.Helper()
	r := e.install(t, home, path, env...)
	if r.code != 0 {
		t.Fatalf("install.sh exited %d", r.code)
	}
	return r
}

// interactive runs cmd in an interactive shell that reads home's rc files,
// the way a new terminal does.
func (e *installEnv) interactive(t *testing.T, home, sh, script string) string {
	t.Helper()
	cmd := exec.Command(filepath.Join(e.tools, sh), "-ic", script)
	cmd.Dir = home
	cmd.Env = []string{"HOME=" + home, "PATH=" + e.tools, "SHELL=" + filepath.Join(e.tools, sh), "TERM=dumb"}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s -ic %q: %v\nstdout:\n%s\nstderr:\n%s", sh, script, err, out, stderr.String())
	}
	return string(out)
}

// realTempDir is t.TempDir with symlinks resolved (macOS: /var -> /private/var),
// matching the paths `mnmd install` writes.
func realTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func sourceLines(t *testing.T, rc string) []string {
	t.Helper()
	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("reading %s: %v", rc, err)
	}
	return regexp.MustCompile(`(?m)^source ".*/src/monom"$`).FindAllString(string(data), -1)
}

func mustContain(t *testing.T, what, got string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("%s: missing %q in:\n%s", what, w, got)
		}
	}
}

func (e *installEnv) testReleaseInstall(t *testing.T, sh string) {
	home := realTempDir(t)
	rc := filepath.Join(home, ".bashrc")
	if sh == "zsh" {
		rc = filepath.Join(home, ".zshrc")
		// A user's zshrc runs compinit; monom's completion registers after it.
		writeFile(t, rc, "autoload -Uz compinit && compinit -u -d "+shQuote(e.compdump)+"\n")
	}
	env := []string{"SHELL=" + filepath.Join(e.tools, sh), "MONOM_DOWNLOAD_BASE=" + e.release}
	dir := filepath.Join(home, ".local", "share", "monom")

	first := e.mustInstall(t, home, e.tools, env...)
	mustContain(t, "first run", first.out, "installed mnmd "+installTestVersion+" in "+dir, "added to "+rc)

	second := e.mustInstall(t, home, e.tools, env...)
	mustContain(t, "second run", second.out, "installed mnmd "+installTestVersion, "already installed")
	if got := sourceLines(t, rc); len(got) != 1 || got[0] != `source "`+dir+`/src/monom"` {
		t.Errorf("want exactly one source line for %s in %s, got %q", dir, rc, got)
	}

	link := filepath.Join(home, ".local", "bin", "mnmd")
	if target, err := os.Readlink(link); err != nil || target != filepath.Join(dir, "bin", "mnmd") {
		t.Errorf("%s -> %q (%v), want a link to %s/bin/mnmd", link, target, err, dir)
	}
	out, err := exec.Command(link, "version").Output()
	if err != nil || strings.TrimSpace(string(out)) != installTestVersion {
		t.Errorf("mnmd version = %q (%v), want %s", out, err, installTestVersion)
	}

	// A new terminal: the rc line loads monom, mnmd and the completion.
	project := filepath.Join(home, "proj")
	if err := os.MkdirAll(filepath.Join(project, "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(project, "monom"), "")
	writeFile(t, filepath.Join(project, "tools", "deploy"), "#!/bin/sh\necho deployed \"$@\"\n")
	if err := os.Chmod(filepath.Join(project, "tools", "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	check := "cd proj && monom tools deploy && mnmd version"
	switch sh {
	case "bash":
		// The registered function called the way readline calls it on Tab,
		// in the same bash the installer ran in (/bin/bash: 3.2 on macOS).
		tab := `COMP_WORDS=(monom tools dep); COMP_CWORD=2; COMP_LINE="monom tools dep"; COMP_POINT=${#COMP_LINE}; ` +
			`_monom_completion monom dep tools; echo "tab: ${COMPREPLY[*]}"; `
		got := e.interactive(t, home, "bash", "type -t monom mnmd; complete -p monom; cd proj && "+tab+"cd ~ && "+check)
		mustContain(t, "bash -ic", got, "function\nfunction\n", "complete -F _monom_completion monom", "tab: deploy\n", "deployed", installTestVersion)
	case "zsh":
		got := e.interactive(t, home, "zsh", `whence -w monom mnmd; print -r -- "completion: ${_comps[monom]}"; `+check)
		mustContain(t, "zsh -ic", got, "monom: function", "mnmd: function", "completion: _monom", "deployed", installTestVersion)
	}

	// A real Tab press in a pseudo-terminal, against the installed tree. The
	// pty driver reads the edit buffer through bind -x and READLINE_LINE,
	// which bash only has from 4.0 (macOS /bin/bash is 3.2).
	if sh == "bash" {
		if out, err := exec.Command("bash", "-c", "echo ${BASH_VERSINFO[0]}").Output(); err != nil || strings.TrimSpace(string(out)) < "4" {
			t.Logf("bash %s: skipping the pty Tab press (the call above covered the completion)", strings.TrimSpace(string(out)))
			return
		}
	}
	rel, err := filepath.Rel(dir, project)
	if err != nil {
		t.Fatal(err)
	}
	s := startPTY(t, dir, rel, sh, e.compdump, nil)
	if line := s.press("monom tools dep", 1); line != "monom tools deploy " {
		s.fail("Tab after 'monom tools dep' gave " + shQuote(line))
	}
}

func (e *installEnv) testChecksumMismatch(t *testing.T) {
	dist := t.TempDir()
	asset := "monom-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	writeFile(t, filepath.Join(dist, asset), "not the real tarball")
	resp, err := http.Get(e.release + "/checksums.txt")
	if err != nil {
		t.Fatal(err)
	}
	sums, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	writeFile(t, filepath.Join(dist, "checksums.txt"), string(sums))
	srv := httptest.NewServer(http.FileServer(http.Dir(dist)))
	defer srv.Close()

	home := realTempDir(t)
	r := e.install(t, home, e.tools, "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+srv.URL)
	if r.code == 0 {
		t.Fatal("install.sh succeeded with a corrupt tarball")
	}
	mustContain(t, "stderr", r.errOut, "sha256 mismatch for "+asset)
	e.assertNothingInstalled(t, home)
}

func (e *installEnv) testNoRelease(t *testing.T) {
	home := realTempDir(t)
	r := e.install(t, home, e.tools, "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+e.empty)
	if r.code == 0 {
		t.Fatal("install.sh succeeded with no release")
	}
	mustContain(t, "stderr", r.errOut,
		"no monom release for "+runtime.GOOS+"/"+runtime.GOARCH, "Build it from source", "https://github.com/adamgen/monom#install")
	e.assertNothingInstalled(t, home)
}

func (e *installEnv) assertNothingInstalled(t *testing.T, home string) {
	t.Helper()
	for _, p := range []string{".local", ".bashrc", ".zshrc", ".bash_profile"} {
		if _, err := os.Lstat(filepath.Join(home, p)); err == nil {
			t.Errorf("%s was created by a failed install", p)
		}
	}
}
