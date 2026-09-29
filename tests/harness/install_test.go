// install.sh, run the way a user runs it: piped into bash, in an empty HOME,
// against release tarballs that build.sh produced and a local HTTP server
// serves (MONOM_DOWNLOAD_BASE). Most scenarios run with no Go on PATH: PATH
// is a directory of symlinks to the handful of tools the installer and the
// shell integration use.
package harness

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
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
	"bash", "zsh", "sh", "curl", "wget", "tar", "gzip", "mkdir", "mktemp", "rm", "mv",
	"cp", "ln", "chmod", "cat", "sed", "awk", "grep", "uname", "tr", "cut", "find",
	"head", "tail", "dirname", "basename", "sha256sum", "shasum", "date", "wc",
	"sort", "env", "ls", "touch", "sysctl",
}

type installEnv struct {
	t        *testing.T
	repo     string
	script   []byte
	tools    string // PATH without go
	release  string // URL serving the release assets
	empty    string // URL with no assets (a repo with no release)
	source   string // URL of a source tarball of this tree
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
		t.Run("release/no-go/"+sh, func(t *testing.T) { e.testReleaseInstall(t, sh) })
	}
	t.Run("release/wget-only", e.testWgetOnly)
	t.Run("release/upgrades-earlier-install-in-place", e.testUpgradeInPlace)
	t.Run("release/stops-for-a-wired-git-checkout", e.testGitCheckoutStops)
	t.Run("release/checksum-mismatch-aborts", e.testChecksumMismatch)
	t.Run("no-release/no-go-fails-clearly", e.testNoReleaseNoGo)
	t.Run("no-release/go-builds-from-source", e.testSourceFallback)
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

	srcDir := t.TempDir()
	writeSourceTarball(t, repo, filepath.Join(srcDir, "main.tar.gz"))

	release := httptest.NewServer(http.FileServer(http.Dir(dist)))
	empty := httptest.NewServer(http.NotFoundHandler())
	source := httptest.NewServer(http.FileServer(http.Dir(srcDir)))
	t.Cleanup(release.Close)
	t.Cleanup(empty.Close)
	t.Cleanup(source.Close)

	return &installEnv{
		t: t, repo: repo, script: script,
		tools:    toolDir(t, installTools),
		release:  release.URL,
		empty:    empty.URL,
		source:   source.URL + "/main.tar.gz",
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

// writeSourceTarball packs the files a source build needs the way GitHub's
// archive endpoint does: everything under one top-level directory.
func writeSourceTarball(t *testing.T, repo, dest string) {
	f, err := os.Create(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, top := range []string{"go.mod", "build.sh", "README.md", "cmd", "internal", "src"} {
		err := filepath.WalkDir(filepath.Join(repo, top), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(repo, path)
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = "monom-main/" + filepath.ToSlash(rel)
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()
			_, err = io.Copy(tw, src)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
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
	mustContain(t, "first run", first.out, "verified sha256", "installed mnmd "+installTestVersion+" to "+dir, "added to "+rc)
	if strings.Contains(first.out, "compinit") {
		t.Errorf("no compinit note expected when the zshrc runs compinit:\n%s", first.out)
	}

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

func (e *installEnv) testWgetOnly(t *testing.T) {
	if _, err := exec.LookPath("wget"); err != nil {
		t.Skip("wget not available")
	}
	var tools []string
	for _, tool := range installTools {
		if tool != "curl" {
			tools = append(tools, tool)
		}
	}
	home := realTempDir(t)
	r := e.mustInstall(t, home, toolDir(t, tools), "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+e.release)
	mustContain(t, "wget run", r.out, "verified sha256", "installed mnmd "+installTestVersion)
}

func (e *installEnv) testUpgradeInPlace(t *testing.T) {
	home := realTempDir(t)
	old := filepath.Join(home, ".monom")
	for _, d := range []string{"bin", "src"} {
		if err := os.MkdirAll(filepath.Join(old, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	rc := filepath.Join(home, ".bashrc")
	writeFile(t, rc, "# mine\nsource \"$HOME/.monom/src/monom\"\n")

	r := e.mustInstall(t, home, e.tools, "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+e.release, "MONOM_BIN_DIR=")
	mustContain(t, "upgrade run", r.out, "found an earlier install at "+old, "installed mnmd "+installTestVersion+" to "+old)
	if _, err := os.Stat(filepath.Join(home, ".local", "share", "monom")); err == nil {
		t.Error("a second install was created in ~/.local/share/monom")
	}
	if _, err := os.Lstat(filepath.Join(home, ".local", "bin", "mnmd")); err == nil {
		t.Error("MONOM_BIN_DIR= should skip the symlink")
	}
	mustContain(t, "upgrade run", r.out, "already installed (your shell sources "+old+"/src/monom)")
	if got := sourceLines(t, rc); len(got) != 1 {
		t.Errorf("the $HOME/.monom line already sources this install; want it alone, got %q", got)
	}
}

func (e *installEnv) testGitCheckoutStops(t *testing.T) {
	home := realTempDir(t)
	checkout := filepath.Join(home, ".monom")
	if err := os.MkdirAll(filepath.Join(checkout, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	rc := filepath.Join(home, ".bashrc")
	line := "source \"" + checkout + "/src/monom\"\n"
	writeFile(t, rc, line)

	r := e.install(t, home, e.tools, "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+e.release)
	if r.code == 0 {
		t.Fatal("install.sh installed a second copy next to a wired git checkout")
	}
	mustContain(t, "stderr", r.errOut, "already sources a monom git checkout at "+checkout, "git pull && ./build.sh", "MONOM_INSTALL_DIR")
	if data, _ := os.ReadFile(rc); string(data) != line {
		t.Errorf("rc file changed:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(home, ".local")); err == nil {
		t.Error("~/.local was created")
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
	mustContain(t, "stderr", r.errOut, "checksum mismatch for "+asset)
	e.assertNothingInstalled(t, home)
}

func (e *installEnv) testNoReleaseNoGo(t *testing.T) {
	home := realTempDir(t)
	r := e.install(t, home, e.tools, "SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE="+e.empty)
	if r.code == 0 {
		t.Fatal("install.sh succeeded with no release and no Go")
	}
	mustContain(t, "stderr", r.errOut,
		"no monom-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz at "+e.empty,
		"Go is not installed", "https://go.dev/dl/", "MONOM_VERSION=")
	e.assertNothingInstalled(t, home)
}

func (e *installEnv) testSourceFallback(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not available")
	}
	home := realTempDir(t)
	path := e.tools + string(os.PathListSeparator) + filepath.Dir(goBin)
	env := []string{"SHELL=/bin/bash", "MONOM_DOWNLOAD_BASE=" + e.empty, "MONOM_SOURCE_URL=" + e.source}
	// Reuse the build cache so the test doesn't compile the standard library.
	if cache, err := exec.Command("go", "env", "GOCACHE").Output(); err == nil {
		env = append(env, "GOCACHE="+strings.TrimSpace(string(cache)))
	}
	r := e.mustInstall(t, home, path, env...)
	dir := filepath.Join(home, ".local", "share", "monom")
	mustContain(t, "source run", r.out, "falling back to a source build", "installed mnmd source to "+dir, "added to ")
	got := e.interactive(t, home, "bash", "type -t monom; complete -p monom")
	mustContain(t, "bash -ic", got, "function", "complete -F _monom_completion monom")
}

func (e *installEnv) assertNothingInstalled(t *testing.T, home string) {
	t.Helper()
	for _, p := range []string{".local", ".bashrc", ".zshrc", ".bash_profile"} {
		if _, err := os.Lstat(filepath.Join(home, p)); err == nil {
			t.Errorf("%s was created by a failed install", p)
		}
	}
}
