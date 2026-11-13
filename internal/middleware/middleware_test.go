package middleware

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseResolvesNamesAndPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mw"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, mode := range map[string]os.FileMode{"mw/ok": 0o755, "mw/plain": 0o644} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("#!/bin/sh\n"), mode); err != nil {
			t.Fatal(err)
		}
	}
	self, _ := os.Executable()

	words, err := Parse(root, []byte("\nmw/ok --flag x\n  mnmd cache  \nsh\n"))
	if err != nil {
		t.Fatal(err)
	}
	if words[0] != filepath.Join(root, "mw/ok") || words[1] != "--flag" || words[2] != "x" ||
		words[3] != self || words[4] != "cache" || !filepath.IsAbs(words[5]) || len(words) != 6 {
		t.Fatalf("words = %q", words)
	}

	for line, want := range map[string]string{
		"mw/plain":          "not executable",
		"mw/none":           "no such file",
		"./mw":              "not a regular file",
		"no-such-tool-here": "not found on PATH",
	} {
		_, err := Parse(root, []byte(line+"\n"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err %v, want %q", line, err, want)
		}
	}
}

func TestChainIsEmptyWithoutAHookScript(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "monom"), []byte("x/y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	words, err := Chain(root, filepath.Join(root, "cmd"))
	if err != nil || words != nil {
		t.Fatalf("words %q, err %v", words, err)
	}
}
