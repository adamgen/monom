package pack

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adamgen/monom/internal/cli"
)

// makeProject creates a temp project root with an executable "monom" file and
// returns its real (symlink-resolved) path.
func makeProject(t *testing.T) string {
	t.Helper()
	dir := realPath(t, t.TempDir())
	writeExec(t, filepath.Join(dir, "monom"), "#!/bin/sh\n")
	return dir
}

func realPath(t *testing.T, p string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("realPath(%q): %v", p, err)
	}
	return resolved
}

func writeExec(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("writeExec: %v", err)
	}
}

func writeNonExec(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeNonExec: %v", err)
	}
}

func withEnv(t *testing.T, key, val string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if val == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, val)
	}
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func withWd(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestPack_SingleToken(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "command1"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	got, err := Pack([]string{"command1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(project, "command1")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPack_TwoTokensJoinedWithSlash(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "category1", "sub_command1"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	got, err := Pack([]string{"category1", "sub_command1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(project, "category1", "sub_command1")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPack_NestedPath(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "infra", "cloud", "deploy"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	got, err := Pack([]string{"infra", "cloud", "deploy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(project, "infra", "cloud", "deploy")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPack_NoProjectRoot(t *testing.T) {
	emptyDir := t.TempDir()
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, emptyDir)

	_, err := Pack([]string{"command1"})
	if err == nil {
		t.Fatal("expected error when no project root, got nil")
	}
}

func TestPack_FileNotFound(t *testing.T) {
	project := makeProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestPack_FileExistsButNotExecutable(t *testing.T) {
	project := makeProject(t)
	writeNonExec(t, filepath.Join(project, "command1"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"command1"})
	if err == nil {
		t.Fatal("expected error for non-executable file, got nil")
	}
}

func TestPack_EmptyWordsSlice(t *testing.T) {
	project := makeProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{})
	if err == nil {
		t.Fatal("expected error for empty words, got nil")
	}
}

func TestPack_DirectoryReturnsGroupErrorWithPath(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "infra", "cloud"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	// Pack signals "this is a group" via *GroupError and does NOT enumerate
	// children — discovery is the `complete` hook's job, not pack's.
	_, err := Pack([]string{"infra"})
	var ge *GroupError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GroupError, got %v", err)
	}
	if ge.Path != filepath.Join(project, "infra") {
		t.Errorf("group path: got %q, want %q", ge.Path, filepath.Join(project, "infra"))
	}
}

func TestPack_NestedDirectoryReturnsGroupError(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "infra", "cloud", "deploy"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"infra", "cloud"})
	var ge *GroupError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GroupError, got %v", err)
	}
	if ge.Path != filepath.Join(project, "infra", "cloud") {
		t.Errorf("group path: got %q, want %q", ge.Path, filepath.Join(project, "infra", "cloud"))
	}
}

func TestPack_EmptyDirectoryReturnsGroupError(t *testing.T) {
	project := makeProject(t)
	if err := os.MkdirAll(filepath.Join(project, "empty"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"empty"})
	var ge *GroupError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GroupError, got %v", err)
	}
}

func TestPack_LeafIsNotAGroupError(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "command1"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"command1"})
	var ge *GroupError
	if errors.As(err, &ge) {
		t.Fatalf("leaf command must not be a GroupError, got %v", err)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupError_SatisfiesCodedError(t *testing.T) {
	project := makeProject(t)
	writeExec(t, filepath.Join(project, "infra", "cloud"), "#!/bin/sh\n")
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"infra"})
	var ce cli.CodedError
	if !errors.As(err, &ce) {
		t.Fatalf("expected CodedError, got %T: %v", err, err)
	}
	if ce.ExitCode() != cli.ExitCodes.GroupError {
		t.Errorf("ExitCode() = %d, want %d", ce.ExitCode(), cli.ExitCodes.GroupError)
	}
}

func TestPack_NotFoundIsNotAGroupError(t *testing.T) {
	project := makeProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", project)

	_, err := Pack([]string{"nonexistent"})
	var ge *GroupError
	if errors.As(err, &ge) {
		t.Fatalf("missing path must not be a GroupError, got %v", err)
	}
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
