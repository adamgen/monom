package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- rcFileForShell ---

func TestRcFileForShell_zsh(t *testing.T) {
	home, _ := os.UserHomeDir()
	got, err := rcFileForShell("/bin/zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".zshrc")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRcFileForShell_bash_profile_exists(t *testing.T) {
	home, _ := os.UserHomeDir()
	profile := filepath.Join(home, ".bash_profile")
	// Only run this assertion if .bash_profile actually exists on this machine.
	if _, err := os.Stat(profile); err == nil {
		got, err := rcFileForShell("/bin/bash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != profile {
			t.Errorf("got %q, want %q", got, profile)
		}
	}
}

func TestRcFileForShell_unknown(t *testing.T) {
	_, err := rcFileForShell("/bin/fish")
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error should mention unsupported shell, got: %v", err)
	}
}

func TestRcFileForShell_empty(t *testing.T) {
	_, err := rcFileForShell("")
	if err == nil {
		t.Fatal("expected error for empty shell, got nil")
	}
}

// --- alreadyInstalled ---

func TestAlreadyInstalled_present(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := `# existing config
source "/usr/local/monom/src/monom"
`
	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := alreadyInstalled(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected alreadyInstalled=true")
	}
}

func TestAlreadyInstalled_commented_out_not_counted(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := `# source "/usr/local/monom/src/monom"
# old setup, kept for reference
`
	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := alreadyInstalled(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected alreadyInstalled=false for commented-out line")
	}
}

func TestAlreadyInstalled_absent(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(rc, []byte("# empty\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := alreadyInstalled(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected alreadyInstalled=false")
	}
}

func TestAlreadyInstalled_file_not_exist(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc") // does not exist
	got, err := alreadyInstalled(rc, "/some/path/src/monom")
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if got {
		t.Error("expected alreadyInstalled=false for missing file")
	}
}

// --- needsLeadingNewline ---

func TestNeedsLeadingNewline_ends_with_newline(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "rc")
	if err := os.WriteFile(f, []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if needsLeadingNewline(f) {
		t.Error("expected false when file ends with newline")
	}
}

func TestNeedsLeadingNewline_missing_trailing_newline(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "rc")
	if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if !needsLeadingNewline(f) {
		t.Error("expected true when file does not end with newline")
	}
}

func TestNeedsLeadingNewline_empty_file(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "rc")
	if err := os.WriteFile(f, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if needsLeadingNewline(f) {
		t.Error("expected false for empty file")
	}
}

// --- appendSourceLine ---

func TestAppendSourceLine_adds_line_with_trailing_newline(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(rc, []byte("# existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	srcMonom := "/usr/local/monom/src/monom"
	if err := appendSourceLine(rc, srcMonom); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(rc)
	content := string(data)
	want := `source "` + srcMonom + `"`
	if !strings.Contains(content, want) {
		t.Errorf("rc file missing source line; got:\n%s", content)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Error("rc file should end with newline after append")
	}
}

func TestAppendSourceLine_prepends_newline_when_missing(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	// No trailing newline.
	if err := os.WriteFile(rc, []byte("# existing"), 0644); err != nil {
		t.Fatal(err)
	}
	srcMonom := "/usr/local/monom/src/monom"
	if err := appendSourceLine(rc, srcMonom); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(rc)
	content := string(data)
	want := "\n" + `source "` + srcMonom + `"`
	if !strings.Contains(content, want) {
		t.Errorf("expected leading newline before source line; got:\n%s", content)
	}
}

// --- resolveSrcMonom ---

func TestResolveSrcMonom_sibling_path(t *testing.T) {
	// Create a fake bin/mnmd structure in a temp dir.
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(binDir, "mnmd")
	if err := os.WriteFile(exe, []byte(""), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveSrcMonom(exe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Resolve dir symlinks (macOS /var → /private/var) before comparing.
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks on tempdir: %v", err)
	}
	want := filepath.Join(realDir, "src", "monom")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
