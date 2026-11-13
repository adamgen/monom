package runcache

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Invalidation by inputs, the command file, env and next's argv is covered
// end to end in tests/monom_cache_test; this checks what that suite can't
// see: the execute bit, the entry itself, and set-but-empty env.
func TestKeyChangesWithModeEntryAndEnv(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"cmd", "src/a"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, f), []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	e := Entry{Inputs: []string{"src"}, Env: []string{"RUNCACHE_T"}}
	key := func(e Entry) string {
		k, err := Key(root, "cmd", nil, e)
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	seen := map[string]string{key(e): "base"}
	for name, change := range map[string]func() string{
		"entry":     func() string { return key(Entry{Inputs: e.Inputs, Env: e.Env, Outputs: []string{"o"}}) },
		"empty env": func() string { t.Setenv("RUNCACHE_T", ""); return key(e) },
		"exec bit": func() string {
			if err := os.Chmod(filepath.Join(root, "src/a"), 0o755); err != nil {
				t.Fatal(err)
			}
			return key(e)
		},
	} {
		k := change()
		if prev, dup := seen[k]; dup {
			t.Errorf("%s: same key as %s", name, prev)
		}
		seen[k] = name
	}
}

func TestLoadRejectsPathsOutsideTheProject(t *testing.T) {
	for _, body := range []string{
		`{"cmd": {"inputs": ["../x"]}}`,
		`{"cmd": {"inputs": ["a"], "outputs": ["/tmp/x"]}}`,
		`{"cmd": {"inputs": ["a"], "outputs": [".monom/cache/x"]}}`,
	} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, FileName), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Errorf("%s: want an error", body)
		}
	}
}

// A signal sent to the middleware itself reaches the command, and the
// middleware survives to return the command's status.
func TestRunRelaysSIGTERMAndSIGINT(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), sig)
		}()
		var out bytes.Buffer
		script := `trap 'echo got; exit 5' TERM INT; i=0; while [ $i -lt 50 ]; do sleep 0.1; i=$((i+1)); done; exit 9`
		status, err := Run([]string{"sh", "-c", script}, &out, &out)
		if err != nil || status != 5 || !strings.Contains(out.String(), "got") {
			t.Fatalf("%v: status %d, err %v, out %q", sig, status, err, out.String())
		}
	}
}
