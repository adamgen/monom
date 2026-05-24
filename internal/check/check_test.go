package check

import (
	"os"
	"path/filepath"
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
