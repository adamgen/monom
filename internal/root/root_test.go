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

func TestFindProjectRoot_EnvVarHonoredWithoutConfigFile(t *testing.T) {
	// An alias pins the root to a directory that has no monom config file at
	// all — zero-config. The pin must win over the walk.
	aliasTarget := realPath(t, t.TempDir())
	withEnv(t, "_MONOM_PROJECT_ROOT", aliasTarget)
	withWd(t, makeProject(t))

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != aliasTarget {
		t.Errorf("got %q, want %q", got, aliasTarget)
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

func TestFindProjectRoot_NonExecutableMonomFileMarksRoot(t *testing.T) {
	// `touch monom` produces an empty, non-executable file; it is a valid
	// (empty) monom config file and marks the nearest root.
	outer := makeProject(t)
	inner := filepath.Join(outer, "child")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inner, "monom"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, inner)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != inner {
		t.Errorf("got %q, want %q", got, inner)
	}
}

func TestFindProjectRoot_DirectoryNamedMonomIsNotAMarker(t *testing.T) {
	dir := realPath(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(dir, "monom"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, dir)

	if got, err := FindProjectRoot(); err == nil {
		t.Fatalf("expected error, got root %q", got)
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

func TestFindProjectRoot_GitRootIsFallbackWithoutConfigFile(t *testing.T) {
	repo := realPath(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	subdir := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, subdir)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != repo {
		t.Errorf("got %q, want %q", got, repo)
	}
}

func TestFindProjectRoot_GitFileMarksWorktreeRoot(t *testing.T) {
	worktree := realPath(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, worktree)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != worktree {
		t.Errorf("got %q, want %q", got, worktree)
	}
}

func TestFindProjectRoot_ConfigFileAboveGitRootWins(t *testing.T) {
	// A monom file anywhere up the chain is an explicit marker and beats the
	// implicit git fallback, even when the git root is nearer.
	project := makeProject(t)
	repo := filepath.Join(project, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, repo)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != project {
		t.Errorf("got %q, want %q", got, project)
	}
}

func TestFindProjectRoot_ConfigFileInsideGitRepoWins(t *testing.T) {
	repo := realPath(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	sub := filepath.Join(repo, "tools")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "monom"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	withEnv(t, "_MONOM_PROJECT_ROOT", "")
	withWd(t, sub)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != sub {
		t.Errorf("got %q, want %q", got, sub)
	}
}
