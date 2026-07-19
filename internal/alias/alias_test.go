package alias

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeProject creates a temp dir with an executable monom config and returns it.
func makeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	monom := filepath.Join(dir, "monom")
	if err := os.WriteFile(monom, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// --- BindLine ---

func TestBindLine_format(t *testing.T) {
	got := BindLine("myapp", "/path/to/project")
	want := `_monom_bind_alias myapp "/path/to/project"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- name validation ---

func TestSave_rejects_invalid_name(t *testing.T) {
	proj := makeProject(t)
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", t.TempDir())
	for _, bad := range []string{"bad name", "1abc", "has/slash", "", "dot.name"} {
		if err := Save(bad, proj); err == nil {
			t.Errorf("expected error for invalid name %q, got nil", bad)
		}
	}
}

func TestSave_accepts_valid_names(t *testing.T) {
	proj := makeProject(t)
	home := t.TempDir()
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", home)
	for _, ok := range []string{"myapp", "my-app", "my_app", "_x", "App2"} {
		if err := Save(ok, proj); err != nil {
			t.Errorf("expected valid name %q to save, got %v", ok, err)
		}
	}
}

// --- root validation ---

func TestSave_rejects_non_project_root(t *testing.T) {
	empty := t.TempDir() // no executable monom file
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", t.TempDir())
	err := Save("myapp", empty)
	if err == nil {
		t.Fatal("expected error for non-project root, got nil")
	}
	if !strings.Contains(err.Error(), "not a monom project root") {
		t.Errorf("error should mention project root, got: %v", err)
	}
}

// --- persistence & idempotency ---

func TestSave_appends_absolute_bind_line(t *testing.T) {
	proj := makeProject(t)
	home := t.TempDir()
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", home)

	if err := Save("myapp", proj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("rc file not written: %v", err)
	}
	// proj is already absolute (t.TempDir), so the line should contain it verbatim.
	want := BindLine("myapp", proj)
	if !strings.Contains(string(data), want) {
		t.Errorf("rc file missing bind line %q; got:\n%s", want, string(data))
	}
}

func TestSave_resolves_relative_path_to_absolute(t *testing.T) {
	proj := makeProject(t)
	home := t.TempDir()
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", home)

	// Run from the project's parent and pass the basename as a relative path.
	parent := filepath.Dir(proj)
	base := filepath.Base(proj)
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd) //nolint:errcheck
	if err := os.Chdir(parent); err != nil {
		t.Fatal(err)
	}

	if err := Save("myapp", base); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(home, ".zshrc"))
	// The persisted path must be absolute, not the relative "base".
	if strings.Contains(string(data), `"`+base+`"`) {
		t.Errorf("bind line should use an absolute path, not %q; got:\n%s", base, string(data))
	}
	if !strings.Contains(string(data), filepath.Join(parent, base)) {
		t.Errorf("bind line missing absolute path; got:\n%s", string(data))
	}
}

func TestSave_idempotent(t *testing.T) {
	proj := makeProject(t)
	home := t.TempDir()
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", home)

	if err := Save("myapp", proj); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	if err := Save("myapp", proj); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(home, ".zshrc"))
	if n := strings.Count(string(data), BindLine("myapp", proj)); n != 1 {
		t.Errorf("expected exactly 1 bind line after idempotent re-save, got %d; content:\n%s", n, string(data))
	}
}

func TestSave_unsupported_shell(t *testing.T) {
	proj := makeProject(t)
	t.Setenv("SHELL", "/usr/bin/fish")
	t.Setenv("HOME", t.TempDir())
	err := Save("myapp", proj)
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error should mention unsupported shell, got: %v", err)
	}
}
