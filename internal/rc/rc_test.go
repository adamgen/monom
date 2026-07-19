package rc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- FileForShell ---

func TestFileForShell_zsh(t *testing.T) {
	home, _ := os.UserHomeDir()
	got, err := FileForShell("/bin/zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".zshrc")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFileForShell_bash_profile_exists(t *testing.T) {
	home, _ := os.UserHomeDir()
	profile := filepath.Join(home, ".bash_profile")
	// Only run this assertion if .bash_profile actually exists on this machine.
	if _, err := os.Stat(profile); err == nil {
		got, err := FileForShell("/bin/bash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != profile {
			t.Errorf("got %q, want %q", got, profile)
		}
	}
}

func TestFileForShell_unknown(t *testing.T) {
	_, err := FileForShell("/bin/fish")
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error should mention unsupported shell, got: %v", err)
	}
}

func TestFileForShell_empty(t *testing.T) {
	_, err := FileForShell("")
	if err == nil {
		t.Fatal("expected error for empty shell, got nil")
	}
}

// --- HasLine ---

func TestHasLine_present(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := `# existing config
source "/usr/local/monom/src/monom"
`
	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := HasLine(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected HasLine=true")
	}
}

func TestHasLine_commented_out_not_counted(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := `# source "/usr/local/monom/src/monom"
# old setup, kept for reference
`
	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := HasLine(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected HasLine=false for commented-out line")
	}
}

func TestHasLine_absent(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(rc, []byte("# empty\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := HasLine(rc, "/usr/local/monom/src/monom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected HasLine=false")
	}
}

func TestHasLine_file_not_exist(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc") // does not exist
	got, err := HasLine(rc, "/some/path/src/monom")
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if got {
		t.Error("expected HasLine=false for missing file")
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

// --- AppendLine ---

func TestAppendLine_adds_line_with_trailing_newline(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(rc, []byte("# existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	line := `source "/usr/local/monom/src/monom"`
	if err := AppendLine(rc, line); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(rc)
	content := string(data)
	if !strings.Contains(content, line) {
		t.Errorf("rc file missing appended line; got:\n%s", content)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Error("rc file should end with newline after append")
	}
}

func TestAppendLine_prepends_newline_when_missing(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	// No trailing newline.
	if err := os.WriteFile(rc, []byte("# existing"), 0644); err != nil {
		t.Fatal(err)
	}
	line := `source "/usr/local/monom/src/monom"`
	if err := AppendLine(rc, line); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(rc)
	content := string(data)
	if !strings.Contains(content, "\n"+line) {
		t.Errorf("expected leading newline before appended line; got:\n%s", content)
	}
}

func TestAppendLine_creates_missing_file(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc") // does not exist yet
	line := `_monom_bind_alias myapp "/path/to/project"`
	if err := AppendLine(rc, line); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("file should have been created: %v", err)
	}
	if string(data) != line+"\n" {
		t.Errorf("got %q, want %q", string(data), line+"\n")
	}
}
