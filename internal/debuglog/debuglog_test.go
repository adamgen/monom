package debuglog_test

import (
	"os"
	"strings"
	"testing"

	"github.com/adamgen/monom/internal/debuglog"
)

func TestLog_NoopWhenEnvUnset(t *testing.T) {
	t.Setenv("MONOM_DEBUG_LOG", "")

	tmp := t.TempDir()
	path := tmp + "/should-not-exist.log"

	debuglog.Log("hello")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected no log file to be created, but %s exists", path)
	}
}

func TestLog_WritesLineWhenEnvSet(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/debug.log"
	t.Setenv("MONOM_DEBUG_LOG", path)

	debuglog.Log("hello world")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "hello world") {
		t.Errorf("expected log to contain 'hello world', got: %q", content)
	}
	// Timestamp prefix: [HH:MM:SS]
	if !strings.HasPrefix(content, "[") {
		t.Errorf("expected line to start with '[', got: %q", content)
	}
}

func TestLog_AppendsOnSecondCall(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/debug.log"
	t.Setenv("MONOM_DEBUG_LOG", path)

	debuglog.Log("first")
	debuglog.Log("second")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "first") {
		t.Errorf("expected log to contain 'first', got: %q", content)
	}
	if !strings.Contains(content, "second") {
		t.Errorf("expected log to contain 'second', got: %q", content)
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), content)
	}
}

func TestLog_SilentOnUnwritablePath(t *testing.T) {
	t.Setenv("MONOM_DEBUG_LOG", "/no/such/directory/debug.log")
	// Must not panic or return an error — just silently no-op.
	debuglog.Log("this should not crash")
}
