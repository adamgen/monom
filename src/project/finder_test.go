package project

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestFindCommands(t *testing.T) {
	// Create test directory structure
	root := t.TempDir()

	// Helper function to create run files
	createRunFile := func(path string) {
		dir := filepath.Join(root, path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		runFile := filepath.Join(dir, "run")
		if err := os.WriteFile(runFile, []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatalf("Failed to create run file in %s: %v", dir, err)
		}
	}


	for _, path := range PathsTestData {
		createRunFile(path)
	}

	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindCommands(root, tt.pathPrefix)

			// Check error cases
			if tt.wantErr {
				if err == nil {
					t.Error("FindCommands() error = nil, want error")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("FindCommands() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("FindCommands() unexpected error = %v", err)
				return
			}

			// Sort both slices for comparison
			sort.Strings(got)
			sort.Strings(tt.wantPaths)

			if !reflect.DeepEqual(got, tt.wantPaths) {
				t.Errorf("FindCommands() = %v, want %v", got, tt.wantPaths)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
} 