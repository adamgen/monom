package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// testEnv holds the test environment configuration
type testEnv struct {
	t          *testing.T
	tempDir    string
	originalWd string
}

// newTestEnv creates a new test environment
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "monom-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Get the real path (resolve symlinks)
	realTempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Create monom file
	monomFile := filepath.Join(tempDir, "monom")
	if err := os.WriteFile(monomFile, []byte("#!/bin/sh\n"), 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create monom file: %v", err)
	}

	return &testEnv{
		t:          t,
		tempDir:    realTempDir,
		originalWd: originalWd,
	}
}

// cleanup restores the original state
func (e *testEnv) cleanup() {
	os.RemoveAll(e.tempDir)
	os.Chdir(e.originalWd)
}

func TestMainOutput(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	// Change to the temp directory
	if err := os.Chdir(env.tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Run in a goroutine so we can read the pipe
	errCh := make(chan error, 1)
	go func() {
		errCh <- run()
		w.Close()
	}()

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	if err := <-errCh; err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	expectedOutput := "MONOM_PROJECT_ROOT: " + env.tempDir + "\n"
	if got := buf.String(); got != expectedOutput {
		t.Errorf("run() output = %q, want %q", got, expectedOutput)
	}
}
