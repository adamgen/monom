package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adamgen/monom/internal/config"
)

func writeFile(t *testing.T, dir, rel, content string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	return p
}

// hookScript returns a hook script that prints complete on `complete` and
// config on `config`.
func hookScript(t *testing.T, dir, complete, cfg string) string {
	t.Helper()
	return writeFile(t, dir, "monom", "#!/bin/sh\ncase \"$1\" in\n  complete) printf '"+complete+"' ;;\n  config) printf '"+cfg+"' ;;\nesac\n", 0o755)
}

// zeroConfigRoot returns a root holding one registered command and one
// executable the gate rejects.
func zeroConfigRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "deploy", "#!/bin/sh\n", 0o755)
	writeFile(t, root, "notes.TXT", "text\n", 0o755)
	return root
}

func only(t *testing.T, rep Report, check string) []Problem {
	t.Helper()
	var out []Problem
	for _, p := range rep.Problems {
		if p.Check == check {
			out = append(out, p)
		}
	}
	return out
}

func mustCheck(t *testing.T, in Input) Report {
	t.Helper()
	rep, err := Check(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return rep
}

func TestCheck_HookOutputPathsAreValidatedAndDiscoveryDoesNotRun(t *testing.T) {
	root := zeroConfigRoot(t)
	cfg := hookScript(t, root, "category1/sub1\\nmy command/sub\\nanother bad/y\\n", "")

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	if rep.Discovered || len(rep.Commands) != 3 {
		t.Errorf("commands = %v discovered = %v, want the hook's 3 paths", rep.Commands, rep.Discovered)
	}
	if got := only(t, rep, PathSpaces); len(got) != 2 || got[0].Severity != config.SeverityError {
		t.Errorf("path-spaces = %v, want 2 errors", got)
	}
	if got := only(t, rep, RootContents); len(got) != 0 {
		t.Errorf("root contents are not inspected when the complete hook answers: %v", got)
	}
}

func TestCheck_EmptyCompleteFallsBackToDiscovery(t *testing.T) {
	root := zeroConfigRoot(t)
	cfg := hookScript(t, root, "", "")

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	if !rep.Discovered || strings.Join(rep.Commands, ",") != "deploy" {
		t.Errorf("commands = %v discovered = %v, want [deploy] via discovery", rep.Commands, rep.Discovered)
	}
}

func TestCheck_RootContentsSeverityResolution(t *testing.T) {
	cases := []struct {
		name, user, project, want string
	}{
		{"default is warning", "", "", config.SeverityWarning},
		{"user env makes it an error", config.SeverityError, "", config.SeverityError},
		{"project overrides user", config.SeverityError, config.SeverityWarning, config.SeverityWarning},
		{"project opts into strict", "", config.SeverityError, config.SeverityError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := zeroConfigRoot(t)
			content := ""
			if c.project != "" {
				content = "check.root-contents = " + c.project + "\n"
			}
			cfg := writeFile(t, root, "monom", content, 0o644)

			rep := mustCheck(t, Input{Root: root, UserConfig: cfg, UserSeverity: c.user})
			got := only(t, rep, RootContents)
			if len(got) != 1 || got[0].Severity != c.want {
				t.Fatalf("root-contents = %v, want one %s", got, c.want)
			}
			if !strings.Contains(got[0].Message, "notes.TXT") {
				t.Errorf("message should name the skipped executable: %q", got[0].Message)
			}
		})
	}
}

func TestCheck_InvalidUserSeverityIsAConfigErrorAndDefaultStands(t *testing.T) {
	root := zeroConfigRoot(t)
	rep := mustCheck(t, Input{Root: root, UserConfig: filepath.Join(root, "monom"), UserSeverity: "loud"})
	if got := only(t, rep, Config); len(got) != 1 || got[0].Severity != config.SeverityError {
		t.Errorf("config = %v, want one error", got)
	}
	if got := only(t, rep, RootContents); len(got) != 1 || got[0].Severity != config.SeverityWarning {
		t.Errorf("root-contents = %v, want default warning", got)
	}
}

func TestCheck_ConfigHookSettingsApplyToScriptProjects(t *testing.T) {
	root := zeroConfigRoot(t)
	cfg := hookScript(t, root, "", "check.root-contents = error\\n")

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	if got := only(t, rep, RootContents); len(got) != 1 || got[0].Severity != config.SeverityError {
		t.Errorf("root-contents = %v, want one error", got)
	}
}

func TestCheck_FailingConfigHookIsAConfigError(t *testing.T) {
	root := zeroConfigRoot(t)
	cfg := writeFile(t, root, "monom", "#!/bin/sh\n[ \"$1\" = config ] && { echo boom >&2; exit 2; }\necho deploy\n", 0o755)

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	got := only(t, rep, Config)
	if len(got) != 1 || !strings.Contains(got[0].Message, "boom") {
		t.Errorf("config = %v, want one error carrying the hook's stderr", got)
	}
}

func TestCheck_DeclarativeConfigProblemsAndInvalidDeclarationsAreErrors(t *testing.T) {
	root := zeroConfigRoot(t)
	cfg := writeFile(t, root, "monom", "missing/cmd\ncheck.bogus = 1\n", 0o644)

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	if got := only(t, rep, Declared); len(got) != 1 || got[0].Severity != config.SeverityError {
		t.Errorf("declared-commands = %v, want one error", got)
	}
	if got := only(t, rep, Config); len(got) != 1 || got[0].Severity != config.SeverityError {
		t.Errorf("config = %v, want one error", got)
	}
}

func TestCheck_NonExecutableHookScriptIsARootContentsFinding(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "deploy", "#!/bin/sh\n", 0o755)
	cfg := writeFile(t, root, "monom", "#!/bin/sh\necho deploy\n", 0o644)

	rep := mustCheck(t, Input{Root: root, UserConfig: cfg})
	got := only(t, rep, RootContents)
	if len(got) != 1 || !strings.Contains(got[0].Message, "not executable") {
		t.Errorf("root-contents = %v, want the missing execute bit reported", got)
	}
}

func TestCheck_EmptyUserConfigPathReturnsError(t *testing.T) {
	if _, err := Check(Input{Root: t.TempDir()}); err == nil {
		t.Fatal("expected error for empty config path")
	}
}

func TestCheck_DiscoveryWithoutRootReturnsError(t *testing.T) {
	if _, err := Check(Input{UserConfig: filepath.Join(t.TempDir(), "monom")}); err == nil {
		t.Fatal("expected error when discovery has no root")
	}
}

func TestCheck_FailingCompleteHookReturnsError(t *testing.T) {
	root := t.TempDir()
	cfg := writeFile(t, root, "monom", "#!/bin/sh\n[ \"$1\" = complete ] && exit 1\nexit 0\n", 0o755)
	if _, err := Check(Input{Root: root, UserConfig: cfg}); err == nil {
		t.Fatal("expected error when complete fails")
	}
}

func TestReportCount(t *testing.T) {
	rep := Report{Problems: []Problem{
		{Severity: config.SeverityError}, {Severity: config.SeverityWarning}, {Severity: config.SeverityWarning},
	}}
	if rep.Count(config.SeverityError) != 1 || rep.Count(config.SeverityWarning) != 2 {
		t.Errorf("counts wrong: %+v", rep)
	}
}
