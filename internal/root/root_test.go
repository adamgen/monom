package root

import (
	"os"
	"path/filepath"
	"testing"
)

// makeProject creates a temp directory with an executable "monom" file and
// returns the real (symlink-resolved) directory path.
func makeProject(t *testing.T) string {
	t.Helper()
	dir := realPath(t, t.TempDir())
	monomFile := filepath.Join(dir, "monom")
	if err := os.WriteFile(monomFile, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("makeProject: %v", err)
	}
	return dir
}

// makeNonExecProject creates a temp directory with a non-executable "monom" file.
func makeNonExecProject(t *testing.T) string {
	t.Helper()
	dir := realPath(t, t.TempDir())
	monomFile := filepath.Join(dir, "monom")
	if err := os.WriteFile(monomFile, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("makeNonExecProject: %v", err)
	}
	return dir
}

// realPath resolves symlinks in p so path comparisons work on macOS.
func realPath(t *testing.T, p string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("realPath(%q): %v", p, err)
	}
	return resolved
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

func TestFindProjectRoot_EnvVarHonoredWhenValid(t *testing.T) {
	project := makeProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", project)
	withWd(t, t.TempDir()) // cwd has no monom — env var must win

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_EnvVarIgnoredWhenMissingFile(t *testing.T) {
	emptyDir := realPath(t, t.TempDir()) // no monom file inside
	withEnv(t, "_MONOM_PROJECT_ROOT", emptyDir)

	project := makeProject(t)
	withWd(t, project)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_EnvVarIgnoredWhenMissingDir(t *testing.T) {
	nonExistentDir := filepath.Join(t.TempDir(), "does_not_exist")
	withEnv(t, "_MONOM_PROJECT_ROOT", nonExistentDir)

	project := makeProject(t)
	withWd(t, project)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_FoundInCurrentPWD(t *testing.T) {
	project := makeProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, project)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_FoundInParent(t *testing.T) {
	project := makeProject(t)
	subdir := filepath.Join(project, "deep", "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, subdir)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_NotFoundAnywhere(t *testing.T) {
	emptyDir := t.TempDir()
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, emptyDir)

	_, err := FindProjectRoot()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindProjectRoot_NonExecutableMonom_SkippedDuringWalk(t *testing.T) {
	validProject := makeProject(t)
	subdir := filepath.Join(validProject, "child")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Non-executable monom in subdir — walk must skip it and find validProject.
	monomInChild := filepath.Join(subdir, "monom")
	if err := os.WriteFile(monomInChild, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, subdir)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != validProject {
		t.Errorf("got %q, want %q", got, validProject)
	}
}

func TestFindProjectRoot_WalkStopsAtFilesystemRoot(t *testing.T) {
	emptyDir := t.TempDir()
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, emptyDir)

	_, err := FindProjectRoot()
	if err == nil {
		t.Fatal("expected error when no monom found, got nil")
	}
}

func TestFindProjectRoot_NonExecProject_Unused(t *testing.T) {
	// Ensure makeNonExecProject compiles and is valid.
	dir := makeNonExecProject(t)
	withEnv(t, "_MONOM_PROJECT_ROOT", dir)
	withWd(t, t.TempDir())

	_, err := FindProjectRoot()
	if err == nil {
		t.Fatal("expected error: non-exec project root should be ignored and walk should fail")
	}
}
