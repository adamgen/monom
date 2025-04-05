package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestMonomProjectRoot(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "monom-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file for negative testing
	tempFile := filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Create a valid directory with monom file
	validDir := filepath.Join(tempDir, "valid")
	if err := os.Mkdir(validDir, 0755); err != nil {
		t.Fatalf("Failed to create valid directory: %v", err)
	}
	monomFile := filepath.Join(validDir, "monom")
	if err := os.WriteFile(monomFile, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("Failed to create monom file: %v", err)
	}

	// Get the real path of validDir
	realValidDir, err := filepath.EvalSymlinks(validDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Create nested directories for parent search testing
	nestedDir := filepath.Join(validDir, "level1", "level2")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create a directory without monom file
	emptyDir := filepath.Join(tempDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	tests := []struct {
		name    string
		envVar  string
		workDir string // working directory for the test
		wantErr bool
	}{
		{
			name:    "Valid directory with monom file",
			envVar:  validDir,
			workDir: "",
			wantErr: false,
		},
		{
			name:    "Directory without monom file",
			envVar:  emptyDir,
			workDir: "",
			wantErr: true,
		},
		{
			name:    "Find monom in parent directory",
			envVar:  "",
			workDir: nestedDir,
			wantErr: false,
		},
		{
			name:    "No monom file in any parent",
			envVar:  "",
			workDir: emptyDir,
			wantErr: true,
		},
		{
			name:    "Non-existent directory",
			envVar:  "/path/that/does/not/exist",
			workDir: "",
			wantErr: true,
		},
		{
			name:    "File instead of directory",
			envVar:  tempFile,
			workDir: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set working directory if specified
			if tt.workDir != "" {
				if err := os.Chdir(tt.workDir); err != nil {
					t.Fatalf("Failed to change working directory: %v", err)
				}
			}

			// Set environment variable for this test
			if tt.envVar != "" {
				os.Setenv("MONOM_PROJECT_ROOT", tt.envVar)
			} else {
				os.Unsetenv("MONOM_PROJECT_ROOT")
			}
			defer os.Unsetenv("MONOM_PROJECT_ROOT")

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the function
			err := run()

			// Restore stdout and get output
			w.Close()
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()
			os.Stdout = oldStdout
			output := <-outC

			if tt.wantErr && err == nil {
				t.Errorf("run() expected error but got none, output: %s", output)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("run() unexpected error: %v, output: %s", err, output)
			}

			// If we expect success and didn't set MONOM_PROJECT_ROOT, verify it was set correctly
			if !tt.wantErr && tt.envVar == "" {
				got := os.Getenv("MONOM_PROJECT_ROOT")
				// Compare the paths after resolving symlinks
				gotReal, err := filepath.EvalSymlinks(got)
				if err != nil {
					t.Errorf("Failed to resolve symlinks for got path: %v", err)
				}
				if gotReal != realValidDir {
					t.Errorf("MONOM_PROJECT_ROOT = %q (resolved to %q), want %q", got, gotReal, realValidDir)
				}
			}

			t.Logf("Output: %s", output)
			t.Logf("Error: %v", err)
		})
	}
}
