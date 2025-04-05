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

	// Create test structure
	paths := []string{
		// Original test paths
		"project1/command_1",
		"project1/command_1/sub_command_1",
		"project1/command_1/sub_command_2",
		"project1/command_2",
		"project1/command_2/sub_command_3",
		"project1/command_2/sub_command_4",
		"project2",

		// Additional test paths with different naming patterns
		"tools/git-helper",
		"tools/git-helper/commit",
		"tools/git-helper/push",
		"tools/docker/build",
		"tools/docker/run",
		"tools/docker/compose",
		
		// Paths with numbers and special characters
		"api-v1/auth",
		"api-v1/users",
		"api-v2/auth",
		"api-v2/users",
		
		// Deep nested paths
		"services/backend/api/v1/handlers/users",
		"services/backend/api/v1/handlers/auth",
		"services/backend/api/v2/handlers/users",
		
		// Similar prefix paths
		"prod/deploy",
		"prod-staging/deploy",
		"production/deploy",
		
		// Paths with underscores and hyphens mixed
		"cloud_functions/auth-service",
		"cloud_functions/user-service",
		"cloud-storage/backup_daily",
		"cloud-storage/backup_weekly",
		
		// Single level commands with similar names
		"build",
		"builder",
		"building",
	}

	for _, path := range paths {
		createRunFile(path)
	}

	tests := []struct {
		name        string
		pathPrefix  string
		wantPaths   []string
		wantErr     bool
		errContains string
	}{
		// Original test cases
		{
			name:       "Find all with 'pro' prefix",
			pathPrefix: "pro",
			wantPaths: []string{
				"prod/deploy",
				"prod-staging/deploy",
				"production/deploy",
				"project1/command_1",
				"project1/command_1/sub_command_1",
				"project1/command_1/sub_command_2",
				"project1/command_2",
				"project1/command_2/sub_command_3",
				"project1/command_2/sub_command_4",
				"project2",
			},
		},
		{
			name:       "Find project1 commands",
			pathPrefix: "project1",
			wantPaths: []string{
				"project1/command_1",
				"project1/command_1/sub_command_1",
				"project1/command_1/sub_command_2",
				"project1/command_2",
				"project1/command_2/sub_command_3",
				"project1/command_2/sub_command_4",
			},
		},
		{
			name:       "Find project1 commands by prefix",
			pathPrefix: "project1/com",
			wantPaths: []string{
				"project1/command_1",
				"project1/command_1/sub_command_1",
				"project1/command_1/sub_command_2",
				"project1/command_2",
				"project1/command_2/sub_command_3",
				"project1/command_2/sub_command_4",
			},
		},
		{
			name:       "Find project1 commands by prefix 2",
			pathPrefix: "project1/command_1",
			wantPaths: []string{
				"project1/command_1",
				"project1/command_1/sub_command_1",
				"project1/command_1/sub_command_2",
			},
		},
		{
			name:       "Find project1 commands by prefix 3",
			pathPrefix: "project1/command_1/s",
			wantPaths: []string{
				"project1/command_1/sub_command_1",
				"project1/command_1/sub_command_2",
			},
		},
		{
			name:       "Find tools by first letter",
			pathPrefix: "t",
			wantPaths: []string{
				"tools/docker/build",
				"tools/docker/compose",
				"tools/docker/run",
				"tools/git-helper",
				"tools/git-helper/commit",
				"tools/git-helper/push",
			},
		},
		{
			name:       "Find tools",
			pathPrefix: "tools",
			wantPaths: []string{
				"tools/docker/build",
				"tools/docker/compose",
				"tools/docker/run",
				"tools/git-helper",
				"tools/git-helper/commit",
				"tools/git-helper/push",
			},
		},
		{
			name:       "Find git-helper commands",
			pathPrefix: "tools/git",
			wantPaths: []string{
				"tools/git-helper",
				"tools/git-helper/commit",
				"tools/git-helper/push",
			},
		},
		{
			name:       "Find docker commands",
			pathPrefix: "tools/docker",
			wantPaths: []string{
				"tools/docker/build",
				"tools/docker/run",
				"tools/docker/compose",
			},
		},
		{
			name:       "Find API v1 endpoints",
			pathPrefix: "api-v1",
			wantPaths: []string{
				"api-v1/auth",
				"api-v1/users",
			},
		},
		{
			name:       "Find all API endpoints",
			pathPrefix: "api-v",
			wantPaths: []string{
				"api-v1/auth",
				"api-v1/users",
				"api-v2/auth",
				"api-v2/users",
			},
		},
		{
			name:       "Find deep nested services",
			pathPrefix: "services/backend/api/v1/handlers",
			wantPaths: []string{
				"services/backend/api/v1/handlers/users",
				"services/backend/api/v1/handlers/auth",
			},
		},
		{
			name:       "Find production related commands",
			pathPrefix: "prod",
			wantPaths: []string{
				"prod/deploy",
				"prod-staging/deploy",
				"production/deploy",
			},
		},
		{
			name:       "Find cloud functions",
			pathPrefix: "cloud_functions",
			wantPaths: []string{
				"cloud_functions/auth-service",
				"cloud_functions/user-service",
			},
		},
		{
			name:       "Find cloud storage backups",
			pathPrefix: "cloud-storage/backup",
			wantPaths: []string{
				"cloud-storage/backup_daily",
				"cloud-storage/backup_weekly",
			},
		},
		{
			name:       "Find build related commands",
			pathPrefix: "build",
			wantPaths: []string{
				"build",
				"builder",
				"building",
			},
		},
		{
			name:        "Non-existent prefix",
			pathPrefix:  "nonexistent",
			wantPaths:   nil,
			wantErr:     true,
			errContains: "no commands found",
		},
		{
			name:        "Empty prefix",
			pathPrefix:  "",
			wantErr:     true,
			errContains: "path prefix cannot be empty",
		},
	}

	for _, tt := range tests {
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