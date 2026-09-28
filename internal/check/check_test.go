package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempScript(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "user_config")
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("writeTempScript: %v", err)
	}
	return path
}

func TestCheck_AllValidPathsReturnsNoProblems(t *testing.T) {
	script := writeTempScript(t, "#!/bin/sh\necho 'category1/sub1\ncategory1/sub2\ncommand1'\n", 0o755)

	problems, err := Check(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 0 {
		t.Errorf("expected no problems, got %v", problems)
	}
}

func TestCheck_PathWithSpaceIsReported(t *testing.T) {
	script := writeTempScript(t, "#!/bin/sh\nprintf 'my command/sub\\ncommand1\\n'\n", 0o755)

	problems, err := Check(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 1 {
		t.Errorf("expected 1 problem, got %v", problems)
	}
}

func TestCheck_MultipleInvalidPathsAllReported(t *testing.T) {
	script := writeTempScript(t, "#!/bin/sh\nprintf 'bad path/x\\nanother bad/y\\ncommand1\\n'\n", 0o755)

	problems, err := Check(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 2 {
		t.Errorf("expected 2 problems, got %v", problems)
	}
}

// --- command map validation (monom-map.json next to the user config) ---

// writeMapProject builds a project dir with an executable user config (empty
// complete output), a command map, and any extra files given as path→mode.
// Returns the user config path.
func writeMapProject(t *testing.T, mapJSON string, files map[string]os.FileMode) string {
	t.Helper()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "monom")
	if err := os.WriteFile(cfg, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "monom-map.json"), []byte(mapJSON), 0o644); err != nil {
		t.Fatalf("write map: %v", err)
	}
	for rel, mode := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	return cfg
}

func checkProblems(t *testing.T, cfg string) []string {
	t.Helper()
	problems, err := Check(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return problems
}

func assertOneProblemContaining(t *testing.T, problems []string, substr string) {
	t.Helper()
	if len(problems) != 1 || !strings.Contains(problems[0], substr) {
		t.Errorf("expected 1 problem containing %q, got %v", substr, problems)
	}
}

func TestCheck_ValidMapReturnsNoProblems(t *testing.T) {
	cfg := writeMapProject(t, `{"db": {"migrate": "scripts/m.sh"}}`,
		map[string]os.FileMode{"scripts/m.sh": 0o755})
	if problems := checkProblems(t, cfg); len(problems) != 0 {
		t.Errorf("expected no problems, got %v", problems)
	}
}

func TestCheck_UnparsableMapIsReported(t *testing.T) {
	cfg := writeMapProject(t, `{"db": {"a": "x.sh", "a": "y.sh"}}`, nil)
	assertOneProblemContaining(t, checkProblems(t, cfg), "duplicate key")
}

func TestCheck_MissingMapTargetIsReported(t *testing.T) {
	cfg := writeMapProject(t, `{"deploy": "ops/gone.sh"}`, nil)
	assertOneProblemContaining(t, checkProblems(t, cfg), "does not exist")
}

func TestCheck_NonExecutableMapTargetIsReported(t *testing.T) {
	cfg := writeMapProject(t, `{"deploy": "ops/d.sh"}`,
		map[string]os.FileMode{"ops/d.sh": 0o644})
	assertOneProblemContaining(t, checkProblems(t, cfg), "not executable")
}

func TestCheck_MapCommandShadowingTreePathIsReported(t *testing.T) {
	cfg := writeMapProject(t, `{"deploy": "ops/d.sh"}`,
		map[string]os.FileMode{"ops/d.sh": 0o755, "deploy": 0o755})
	assertOneProblemContaining(t, checkProblems(t, cfg), "shadows tree path")
}

func TestCheck_MapCategoryShadowingTreeCommandIsReported(t *testing.T) {
	cfg := writeMapProject(t, `{"db": {"migrate": "scripts/m.sh"}}`,
		map[string]os.FileMode{"scripts/m.sh": 0o755, "db": 0o755})
	assertOneProblemContaining(t, checkProblems(t, cfg), "shadows tree command")
}

func TestCheck_MapCategoryOverTreeDirectoryMerges(t *testing.T) {
	// A category over a real directory is the hybrid pattern, not shadowing:
	// unmapped children fall back to pack against that directory.
	cfg := writeMapProject(t, `{"db": {"migrate": "scripts/m.sh"}}`,
		map[string]os.FileMode{"scripts/m.sh": 0o755, "db/seed": 0o755})
	if problems := checkProblems(t, cfg); len(problems) != 0 {
		t.Errorf("expected no problems, got %v", problems)
	}
}

func TestCheck_EmptyUserConfigReturnsError(t *testing.T) {
	_, err := Check("")
	if err == nil {
		t.Fatal("expected error for empty userConfig")
	}
}

func TestCheck_NonExecutableUserConfigReturnsError(t *testing.T) {
	script := writeTempScript(t, "#!/bin/sh\necho 'command1'\n", 0o644)

	_, err := Check(script)
	if err == nil {
		t.Fatal("expected error for non-executable userConfig")
	}
}

func TestCheck_MissingUserConfigReturnsError(t *testing.T) {
	_, err := Check("/tmp/this_path_does_not_exist_monom_test")
	if err == nil {
		t.Fatal("expected error for missing userConfig")
	}
}
