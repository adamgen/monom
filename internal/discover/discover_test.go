package discover

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, content string, mode os.FileMode) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

const script = "#!/bin/sh\necho ok\n"

func TestDiscover_RecordsWhichGateRuleRegisteredEachCommand(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deploy.sh", script, 0o755)
	write(t, root, "bin/server", "\x7fELF", 0o755)
	write(t, root, "Weird.BIN", "\x7fELF", 0o755)

	res := Discover(root, []string{"Weird.BIN"}, nil)
	want := []Command{
		{Path: "Weird.BIN", Via: ViaDeclared},
		{Path: "bin/server", Via: ViaPattern},
		{Path: "deploy.sh", Via: ViaShebang},
	}
	if !reflect.DeepEqual(res.Commands, want) {
		t.Errorf("commands = %+v, want %+v", res.Commands, want)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("a declared executable must not also be reported skipped: %+v", res.Skipped)
	}
}

func TestDiscover_NamePatternRejectsExtensionsUppercaseAndLeadingPunctuation(t *testing.T) {
	for _, name := range []string{"tool.exe", "Tool", "-tool", "lib.so.1"} {
		if NamePattern.MatchString(name) {
			t.Errorf("NamePattern must reject %q", name)
		}
	}
	for _, name := range []string{"tool", "db-migrate", "sub_command1", "2fa"} {
		if !NamePattern.MatchString(name) {
			t.Errorf("NamePattern must accept %q", name)
		}
	}
}

func TestDiscover_SkippedExecutablesCarryTheReason(t *testing.T) {
	root := t.TempDir()
	write(t, root, "data.CSV", "a,b\n", 0o755)
	write(t, root, "my cmd", script, 0o755)

	res := Discover(root, nil, nil)
	if len(res.Commands) != 0 {
		t.Fatalf("nothing should be registered, got %+v", res.Commands)
	}
	reasons := map[string]string{}
	for _, s := range res.Skipped {
		reasons[s.Path] = s.Reason
	}
	if !strings.Contains(reasons["data.CSV"], "no shebang") {
		t.Errorf("data.CSV reason = %q", reasons["data.CSV"])
	}
	if !strings.Contains(reasons["my cmd"], "space") {
		t.Errorf("'my cmd' reason = %q", reasons["my cmd"])
	}
}

func TestDiscover_EveryNoiseDirectoryIsSkipped(t *testing.T) {
	root := t.TempDir()
	for dir := range noiseDirs {
		write(t, root, dir+"/cmd", script, 0o755)
	}
	write(t, root, ".git/hooks/pre-commit", script, 0o755)
	write(t, root, ".hidden-file", script, 0o755)
	write(t, root, "_private", script, 0o755)

	res := Discover(root, nil, nil)
	if len(res.Commands) != 0 || len(res.Skipped) != 0 {
		t.Errorf("noise must be invisible, got commands=%+v skipped=%+v", res.Commands, res.Skipped)
	}
}

func TestDiscover_RootConfigFileIsNotACommandButANestedOneIsABoundary(t *testing.T) {
	root := t.TempDir()
	write(t, root, "monom", script, 0o755)
	write(t, root, "tools/monom", "", 0o644)
	write(t, root, "tools/build", script, 0o755)
	write(t, root, "run", script, 0o755)

	res := Discover(root, nil, nil)
	if got := res.Paths(); !reflect.DeepEqual(got, []string{"run"}) {
		t.Errorf("paths = %v, want [run]", got)
	}
}

func TestDiscover_NonExecutableFilesAreNeitherRegisteredNorSkipped(t *testing.T) {
	root := t.TempDir()
	write(t, root, "readme", script, 0o644)
	res := Discover(root, nil, nil)
	if len(res.Commands) != 0 || len(res.Skipped) != 0 {
		t.Errorf("got commands=%+v skipped=%+v", res.Commands, res.Skipped)
	}
}

func TestDiscover_SymlinkToExecutableIsRegistered(t *testing.T) {
	root := t.TempDir()
	write(t, root, "_impl/tool", script, 0o755)
	if err := os.Symlink(filepath.Join(root, "_impl", "tool"), filepath.Join(root, "tool")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if got := Discover(root, nil, nil).Paths(); !reflect.DeepEqual(got, []string{"tool"}) {
		t.Errorf("paths = %v, want [tool]", got)
	}
}

func TestDiscover_InvalidDeclarationsAreReportedAndNotRegistered(t *testing.T) {
	root := t.TempDir()
	write(t, root, "plain", "text\n", 0o644)
	if err := os.MkdirAll(filepath.Join(root, "group"), 0o755); err != nil {
		t.Fatal(err)
	}

	res := Discover(root, []string{"missing", "plain", "group", "has space"}, nil)
	if len(res.Commands) != 0 {
		t.Errorf("nothing should be registered, got %+v", res.Commands)
	}
	if len(res.Invalid) != 4 {
		t.Errorf("invalid = %v, want 4 entries", res.Invalid)
	}
}

func TestDiscover_DuplicateDeclarationRegisteredOnce(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deploy", script, 0o755)
	res := Discover(root, []string{"deploy", "deploy"}, nil)
	if got := res.Paths(); !reflect.DeepEqual(got, []string{"deploy"}) {
		t.Errorf("paths = %v, want [deploy]", got)
	}
}

func TestDiscover_UnreadableDirectoryIsRecordedAndSkipped(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any directory")
	}
	root := t.TempDir()
	write(t, root, "locked/cmd", script, 0o755)
	locked := filepath.Join(root, "locked")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	res := Discover(root, nil, nil)
	if !reflect.DeepEqual(res.Unreadable, []string{"locked"}) {
		t.Errorf("unreadable = %v, want [locked]", res.Unreadable)
	}
}

func TestIsHiddenOrPrivate_MatchesPrefixesAndHidePatterns(t *testing.T) {
	hide := []string{"*.TXT", "scratch", "tools/wip-*", "a/b"}
	cases := []struct {
		rel  string
		want bool
	}{
		{".git", true},               // built-in: dot prefix
		{"lib/_helper", true},        // built-in: underscore prefix
		{"notes.TXT", true},          // name glob at the root
		{"docs/notes.TXT", true},     // name glob at any depth
		{"notes.txt", false},         // path.Match is case-sensitive
		{"scratch", true},            // exact name
		{"lib/scratch", true},        // exact name, nested
		{"scratchpad", false},        // a name pattern matches the whole name
		{"tools/wip-new", true},      // path glob
		{"other/tools/wip-x", false}, // a path pattern is anchored at the root
		{"tools/wip-x/deep", false},  // `*` never crosses a slash (the dir itself is hidden)
		{"a/b", true},                // literal relative path
		{"x/a/b", false},
		{"deploy", false},
	}
	for _, c := range cases {
		if got := isHiddenOrPrivate(c.rel, path.Base(c.rel), hide); got != c.want {
			t.Errorf("isHiddenOrPrivate(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
	if isHiddenOrPrivate("deploy", "deploy", nil) {
		t.Error("no patterns must hide nothing but the built-in prefixes")
	}
	if isHiddenOrPrivate("x", "x", []string{"["}) {
		t.Error("an invalid pattern must match nothing")
	}
}

func TestDiscover_HidePatternsExcludeFilesAndDirectoriesButNotDeclarations(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deploy", script, 0o755)
	write(t, root, "notes.TXT", "x", 0o755)
	write(t, root, "scratch/try", script, 0o755)
	write(t, root, "lib/scratch/try", script, 0o755)
	write(t, root, "tools/build", script, 0o755)
	write(t, root, "tools/wip-new", script, 0o755)
	write(t, root, "tools/wip-keep", script, 0o755)

	res := Discover(root, []string{"tools/wip-keep"}, []string{"*.TXT", "scratch", "tools/wip-*"})
	want := []string{"deploy", "tools/build", "tools/wip-keep"}
	if got := res.Paths(); !reflect.DeepEqual(got, want) {
		t.Errorf("paths = %v, want %v", got, want)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("a hidden executable must not be reported skipped: %+v", res.Skipped)
	}
	for _, c := range res.Commands {
		if c.Path == "tools/wip-keep" && c.Via != ViaDeclared {
			t.Errorf("tools/wip-keep via %q, want declared", c.Via)
		}
	}
}
