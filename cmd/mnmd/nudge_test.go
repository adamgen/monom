package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr temporarily replaces os.Stderr and returns what was written.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = orig })

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()
	return buf.String()
}

func TestCheckNudge_fires_when_MONOM_ACTIVE_unset(t *testing.T) {
	t.Setenv("MONOM_ACTIVE", "")
	out := captureStderr(t, func() { checkNudge("filter") })
	if !strings.Contains(out, "mnmd install") {
		t.Errorf("expected nudge on stderr, got: %q", out)
	}
}

func TestCheckNudge_suppressed_when_MONOM_ACTIVE_set(t *testing.T) {
	t.Setenv("MONOM_ACTIVE", "1")
	out := captureStderr(t, func() { checkNudge("filter") })
	if out != "" {
		t.Errorf("expected no nudge when MONOM_ACTIVE=1, got: %q", out)
	}
}

func TestCheckNudge_suppressed_for_install_subcommand(t *testing.T) {
	t.Setenv("MONOM_ACTIVE", "")
	out := captureStderr(t, func() { checkNudge("install") })
	if out != "" {
		t.Errorf("expected no nudge for install subcommand, got: %q", out)
	}
}
