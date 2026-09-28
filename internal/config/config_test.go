package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "monom")
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestLoad_MissingFileIsAbsentNotError(t *testing.T) {
	f, err := Load(filepath.Join(t.TempDir(), "monom"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Kind != Absent {
		t.Errorf("kind = %v, want Absent", f.Kind)
	}
}

func TestLoad_EmptyFileIsDeclarativeWithNothingDeclared(t *testing.T) {
	f, err := Load(writeFile(t, "", 0o644))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Kind != Declarative || len(f.Declared) != 0 || len(f.Problems) != 0 {
		t.Errorf("got %+v, want empty Declarative", f)
	}
}

func TestLoad_ExecutableShebangFileIsScriptAndNotParsed(t *testing.T) {
	f, err := Load(writeFile(t, "#!/bin/sh\ntools/build\ncheck.root-contents = error\n", 0o755))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Kind != Script {
		t.Errorf("kind = %v, want Script", f.Kind)
	}
	if len(f.Declared) != 0 || !reflect.DeepEqual(f.Settings, Settings{}) {
		t.Errorf("a script must never be parsed, got %+v", f)
	}
}

func TestLoad_NonExecutableShebangFileIsInertScript(t *testing.T) {
	f, err := Load(writeFile(t, "#!/bin/sh\n", 0o644))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Kind != InertScript {
		t.Errorf("kind = %v, want InertScript", f.Kind)
	}
}

func TestLoad_ExecutableFileWithoutShebangIsStillDeclarative(t *testing.T) {
	f, err := Load(writeFile(t, "tools/build\n", 0o755))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Kind != Declarative || !reflect.DeepEqual(f.Declared, []string{"tools/build"}) {
		t.Errorf("got %+v, want Declarative declaring tools/build", f)
	}
}

func TestLoad_DirectoryIsAnError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "monom")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("expected error for a directory")
	}
}

func TestParse_DeclarationsSettingsAndComments(t *testing.T) {
	input := `
# comment
  tools/build
./scripts/release/
check.root-contents = error
`
	declared, settings, problems := Parse(strings.NewReader(input), true)
	if want := []string{"tools/build", "scripts/release"}; !reflect.DeepEqual(declared, want) {
		t.Errorf("declared = %v, want %v", declared, want)
	}
	if settings.RootContents != SeverityError {
		t.Errorf("RootContents = %q, want error", settings.RootContents)
	}
	if len(problems) != 0 {
		t.Errorf("unexpected problems: %v", problems)
	}
}

func TestParse_DeclarationsOutsideTheRootAreProblems(t *testing.T) {
	for _, line := range []string{"/usr/bin/env", "../sibling/cmd", "a/../../b", "./", "a/./b"} {
		declared, _, problems := Parse(strings.NewReader(line), true)
		if len(declared) != 0 || len(problems) != 1 {
			t.Errorf("%q: declared=%v problems=%v, want one problem", line, declared, problems)
		}
	}
}

func TestParse_UnknownKeyAndInvalidValueAreProblemsWithLineNumbers(t *testing.T) {
	_, settings, problems := Parse(strings.NewReader("check.typo = error\ncheck.root-contents = loud\n"), true)
	if len(problems) != 2 {
		t.Fatalf("problems = %v, want 2", problems)
	}
	if !strings.HasPrefix(problems[0], "line 1:") || !strings.HasPrefix(problems[1], "line 2:") {
		t.Errorf("problems lack line numbers: %v", problems)
	}
	if settings.RootContents != "" {
		t.Errorf("invalid value must leave the setting unset, got %q", settings.RootContents)
	}
}

func TestParse_ConfigHookOutputCannotDeclareCommands(t *testing.T) {
	declared, settings, problems := Parse(strings.NewReader("tools/build\ncheck.root-contents = warning\n"), false)
	if len(declared) != 0 || len(problems) != 1 {
		t.Errorf("declared=%v problems=%v, want the path line reported", declared, problems)
	}
	if settings.RootContents != SeverityWarning {
		t.Errorf("settings still apply, got %q", settings.RootContents)
	}
}

func TestSettingsMerge_SetFieldsOverride(t *testing.T) {
	base := Settings{RootContents: SeverityError}
	if got := base.Merge(Settings{}); got.RootContents != SeverityError {
		t.Errorf("unset override must keep base, got %q", got.RootContents)
	}
	if got := base.Merge(Settings{RootContents: SeverityWarning}); got.RootContents != SeverityWarning {
		t.Errorf("set override must win, got %q", got.RootContents)
	}
}

func TestHasShebang(t *testing.T) {
	cases := map[string]bool{"#!/bin/sh\n": true, "#!": true, "# comment\n": false, "": false, "#": false}
	for content, want := range cases {
		if got := HasShebang(writeFile(t, content, 0o644)); got != want {
			t.Errorf("HasShebang(%q) = %v, want %v", content, got, want)
		}
	}
	if HasShebang(filepath.Join(t.TempDir(), "missing")) {
		t.Error("missing file must report false")
	}
}

func TestParse_DiscoverHideIsRepeatableAndNormalized(t *testing.T) {
	_, settings, problems := Parse(strings.NewReader("discover.hide = *.TXT\ndiscover.hide = ./tools/wip-*/\n"), true)
	if len(problems) != 0 {
		t.Fatalf("problems = %v", problems)
	}
	want := []string{"*.TXT", "tools/wip-*"}
	if strings.Join(settings.Hide, "|") != strings.Join(want, "|") {
		t.Errorf("hide = %q, want %q", settings.Hide, want)
	}
}

func TestParse_InvalidDiscoverHidePatternsAreProblems(t *testing.T) {
	for _, v := range []string{"[", "", "/abs", "../up", "a/../b", "./"} {
		_, settings, problems := Parse(strings.NewReader("discover.hide = "+v+"\n"), false)
		if len(problems) != 1 || !strings.Contains(problems[0], KeyDiscoverHide) || len(settings.Hide) != 0 {
			t.Errorf("%q: hide=%q problems=%v, want one discover.hide problem", v, settings.Hide, problems)
		}
	}
}

func TestSettingsMerge_HidePatternsAppendWithoutAliasing(t *testing.T) {
	base := Settings{Hide: make([]string, 1, 4)}
	base.Hide[0] = "a"
	x := base.Merge(Settings{Hide: []string{"b"}})
	y := base.Merge(Settings{Hide: []string{"c"}})
	if strings.Join(x.Hide, ",") != "a,b" || strings.Join(y.Hide, ",") != "a,c" {
		t.Errorf("x=%v y=%v", x.Hide, y.Hide)
	}
}

func TestProjectSettings_ReadsTheConfigHookLeniently(t *testing.T) {
	p := writeFile(t, "#!/bin/sh\n[ \"$1\" = config ] && printf 'discover.hide = scratch\\ndiscover.hide = [\\n'\n", 0o755)
	f, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.ProjectSettings().Hide; strings.Join(got, ",") != "scratch" {
		t.Errorf("hide = %v, want only the valid pattern", got)
	}
	failing := writeFile(t, "#!/bin/sh\nexit 3\n", 0o755)
	if f, _ = Load(failing); len(f.ProjectSettings().Hide) != 0 {
		t.Error("a failing config hook must contribute nothing")
	}
}
