package install

import (
	"os"
	"path/filepath"
	"testing"
)

// --- resolveSrcMonom ---
//
// rc-file behavior (shell selection, line-presence idempotency, newline
// handling) now lives in internal/rc and is tested there.

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
