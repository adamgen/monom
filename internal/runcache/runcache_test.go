package runcache

import (
	"os"
	"path/filepath"
	"testing"
)

func project(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for p, c := range map[string]string{"cmd": "#!/bin/sh\n", "src/a": "a", "src/b/c": "c"} {
		f := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func noEnv(string) (string, bool) { return "", false }

func TestKeyChangesWithEverythingItDependsOn(t *testing.T) {
	root := project(t)
	e := Entry{Inputs: []string{"src"}, Outputs: []string{"out"}, Env: []string{"MODE"}}
	base, err := Key(root, "cmd", nil, e, noEnv)
	if err != nil {
		t.Fatal(err)
	}
	same, _ := Key(root, "cmd", nil, e, noEnv)
	if same != base {
		t.Fatal("key is not deterministic")
	}
	changed := map[string]func() (string, error){
		"args": func() (string, error) { return Key(root, "cmd", []string{"x"}, e, noEnv) },
		"env set": func() (string, error) {
			return Key(root, "cmd", nil, e, func(string) (string, bool) { return "", true })
		},
		"outputs": func() (string, error) {
			return Key(root, "cmd", nil, Entry{Inputs: e.Inputs, Outputs: []string{"other"}, Env: e.Env}, noEnv)
		},
		"nested input content": func() (string, error) {
			if err := os.WriteFile(filepath.Join(root, "src/b/c"), []byte("C"), 0o644); err != nil {
				t.Fatal(err)
			}
			return Key(root, "cmd", nil, e, noEnv)
		},
		"input mode": func() (string, error) {
			if err := os.Chmod(filepath.Join(root, "src/a"), 0o755); err != nil {
				t.Fatal(err)
			}
			return Key(root, "cmd", nil, e, noEnv)
		},
	}
	seen := map[string]string{base: "base"}
	for name, f := range changed {
		k, err := f()
		if err != nil {
			t.Fatal(name, err)
		}
		if prev, dup := seen[k]; dup {
			t.Errorf("%s: key equals %s's", name, prev)
		}
		seen[k] = name
	}
}

func TestKeyFailsOnMissingInput(t *testing.T) {
	if _, err := Key(project(t), "cmd", nil, Entry{Inputs: []string{"nope"}}, noEnv); err == nil {
		t.Fatal("want an error for a missing input")
	}
}

func TestLoadRejectsUnsafeOrIncompleteEntries(t *testing.T) {
	for name, body := range map[string]string{
		"no inputs":      `{"cmd": {"outputs": ["o"]}}`,
		"escaping input": `{"cmd": {"inputs": ["../x"]}}`,
		"absolute out":   `{"cmd": {"inputs": ["a"], "outputs": ["/tmp/x"]}}`,
		"into the cache": `{"cmd": {"inputs": ["a"], "outputs": [".monom/cache/x"]}}`,
		"unknown key":    `{"cmd": {"inputs": ["a"], "args": []}}`,
		"bad env":        `{"cmd": {"inputs": ["a"], "env": ["A=B"]}}`,
	} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, FileName), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestStoreRestoreRoundTrip(t *testing.T) {
	root := project(t)
	e := Entry{Inputs: []string{"src"}, Outputs: []string{"src/b"}}
	if err := Store(root, "k", e, Result{Stdout: []byte("o"), Stderr: []byte("e")}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(root, "src/b")); err != nil {
		t.Fatal(err)
	}
	r, ok := Lookup(root, "k")
	if !ok || string(r.Stdout) != "o" || string(r.Stderr) != "e" {
		t.Fatalf("lookup: %v %q %q", ok, r.Stdout, r.Stderr)
	}
	if err := Restore(root, "k", e); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "src/b/c")); string(b) != "c" {
		t.Fatalf("restored %q", b)
	}
	if err := Store(root, "k2", Entry{Inputs: []string{"src"}, Outputs: []string{"missing"}}, Result{}); err == nil {
		t.Fatal("want an error for a missing output")
	}
	if _, ok := Lookup(root, "k2"); ok {
		t.Fatal("a failed store left an entry")
	}
}
