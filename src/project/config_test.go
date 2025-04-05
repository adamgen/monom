package project

import (
	"os"
	"path/filepath"
	"testing"
)

// testEnv holds the test environment configuration
type testEnv struct {
	t          *testing.T
	tempDir    string
	validDir   string
	emptyDir   string
	nestedDir  string
	tempFile   string
	originalWd string
}

// newTestEnv creates a new test environment with all necessary directories and files
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

	env := &testEnv{
		t:          t,
		tempDir:    tempDir,
		originalWd: originalWd,
	}

	// Create test file
	env.tempFile = filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(env.tempFile, []byte("test"), 0644); err != nil {
		env.cleanup()
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Create valid directory with monom file
	env.validDir = filepath.Join(tempDir, "valid")
	if err := os.Mkdir(env.validDir, 0755); err != nil {
		env.cleanup()
		t.Fatalf("Failed to create valid directory: %v", err)
	}
	if err := env.createMonomFile(env.validDir); err != nil {
		env.cleanup()
		t.Fatalf("Failed to create monom file: %v", err)
	}

	// Create nested directories
	env.nestedDir = filepath.Join(env.validDir, "level1", "level2")
	if err := os.MkdirAll(env.nestedDir, 0755); err != nil {
		env.cleanup()
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create empty directory
	env.emptyDir = filepath.Join(tempDir, "empty")
	if err := os.Mkdir(env.emptyDir, 0755); err != nil {
		env.cleanup()
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	return env
}

// cleanup removes temporary test files and restores the working directory
func (e *testEnv) cleanup() {
	os.RemoveAll(e.tempDir)
	os.Chdir(e.originalWd)
}

// createMonomFile creates a monom file in the specified directory
func (e *testEnv) createMonomFile(dir string) error {
	return os.WriteFile(filepath.Join(dir, "monom"), []byte("#!/bin/sh\n"), 0755)
}

func TestProjectConfig(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	tests := []struct {
		name    string
		envVar  string
		workDir string // working directory for the test
		wantErr bool
	}{
		{
			name:    "Valid directory with monom file",
			envVar:  env.validDir,
			workDir: "",
			wantErr: false,
		},
		{
			name:    "Directory without monom file",
			envVar:  env.emptyDir,
			workDir: "",
			wantErr: true,
		},
		{
			name:    "Find monom in parent directory",
			envVar:  "",
			workDir: env.nestedDir,
			wantErr: false,
		},
		{
			name:    "No monom file in any parent",
			envVar:  "",
			workDir: env.emptyDir,
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
			envVar:  env.tempFile,
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

			// Test LoadConfig
			config, err := LoadConfig()

			if tt.wantErr && err == nil {
				t.Error("LoadConfig() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("LoadConfig() unexpected error: %v", err)
			}

			// If we expect success and didn't set MONOM_PROJECT_ROOT, verify it was set correctly
			if !tt.wantErr {
				if config == nil {
					t.Fatal("LoadConfig() returned nil config when no error was expected")
				}

				// Compare the paths after resolving symlinks
				gotReal, err := filepath.EvalSymlinks(config.RootPath)
				if err != nil {
					t.Errorf("Failed to resolve symlinks for got path: %v", err)
				}
				realValidDir, err := filepath.EvalSymlinks(env.validDir)
				if err != nil {
					t.Errorf("Failed to resolve symlinks for valid path: %v", err)
				}
				if gotReal != realValidDir {
					t.Errorf("config.RootPath = %q (resolved to %q), want %q", config.RootPath, gotReal, realValidDir)
				}

				// Verify environment variable was set correctly when not provided
				if tt.envVar == "" {
					got := os.Getenv("MONOM_PROJECT_ROOT")
					if got != config.RootPath {
						t.Errorf("MONOM_PROJECT_ROOT = %q, want %q", got, config.RootPath)
					}
				}
			}
		})
	}
}

func TestIsProjectRoot(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "Valid project root",
			dir:  env.validDir,
			want: true,
		},
		{
			name: "Not a project root",
			dir:  env.tempDir,
			want: false,
		},
		{
			name: "Non-existent directory",
			dir:  filepath.Join(env.tempDir, "nonexistent"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProjectRoot(tt.dir); got != tt.want {
				t.Errorf("IsProjectRoot() = %v, want %v", got, tt.want)
			}
		})
	}
} 