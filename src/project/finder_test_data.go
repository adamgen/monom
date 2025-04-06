package project

var TestCases = []struct {
	name        string
	pathPrefix  string
	wantPaths   []string
	wantErr     bool
	errContains string
}{
	{
		name:       "Find_all_with_pro_prefix",
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
		name:       "Find_all_with_proj_prefix",
		pathPrefix: "proj",
		wantPaths: []string{
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
		name:       "Find_all_with_prod_prefix",
		pathPrefix: "prod",
		wantPaths: []string{
			"prod/deploy",
			"prod-staging/deploy",
			"production/deploy",
		},
	},
	{
		name:       "Find_project1_commands",
		pathPrefix: "project1",
		wantPaths: []string{
			"command_1",
			"command_1/sub_command_1",
			"command_1/sub_command_2",
			"command_2",
			"command_2/sub_command_3",
			"command_2/sub_command_4",
		},
	},
	{
		name:       "Find_project1_commands_by_prefix",
		pathPrefix: "project1/com",
		wantPaths: []string{
			"command_1",
			"command_1/sub_command_1",
			"command_1/sub_command_2",
			"command_2",
			"command_2/sub_command_3",
			"command_2/sub_command_4",
		},
	},
	{
		name:       "Find_project1_commands_by_prefix_2",
		pathPrefix: "project1/command_1",
		wantPaths: []string{
			"sub_command_1",
			"sub_command_2",
		},
	},
	{
		name:       "Find_project1_commands_by_prefix_3",
		pathPrefix: "project1/command_1/s",
		wantPaths: []string{
			"sub_command_1",
			"sub_command_2",
		},
	},
	{
		name:       "Find_tools_by_first_letter",
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
		name:       "Find_tools",
		pathPrefix: "tools",
		wantPaths: []string{
			"docker/build",
			"docker/compose",
			"docker/run",
			"git-helper",
			"git-helper/commit",
			"git-helper/push",
		},
	},
	{
		name:       "Find_git-helper_commands",
		pathPrefix: "tools/git",
		wantPaths: []string{
			"git-helper",
			"git-helper/commit",
			"git-helper/push",
		},
	},
	{
		name:       "Find_docker_commands",
		pathPrefix: "tools/docker",
		wantPaths: []string{
			"build",
			"run",
			"compose",
		},
	},
	{
		name:       "Find_API_v1_endpoints",
		pathPrefix: "api-v1",
		wantPaths: []string{
			"auth",
			"users",
		},
	},
	{
		name:       "Find_all_API_endpoints",
		pathPrefix: "api-v",
		wantPaths: []string{
			"api-v1/auth",
			"api-v1/users",
			"api-v2/auth",
			"api-v2/users",
		},
	},
	{
		name:       "Find_deep_nested_services",
		pathPrefix: "services/backend/api/v1/handlers",
		wantPaths: []string{
			"users",
			"auth",
		},
	},
	{
		name:       "Find_production_related_commands",
		pathPrefix: "prod",
		wantPaths: []string{
			"deploy",
			"prod-staging/deploy",
			"production/deploy",
		},
	},
	{
		name:       "Find_cloud_functions",
		pathPrefix: "cloud_functions",
		wantPaths: []string{
			"auth-service",
			"user-service",
		},
	},
	{
		name:       "Find_cloud_storage_backups",
		pathPrefix: "cloud-storage/backup",
		wantPaths: []string{
			"backup_daily",
			"backup_weekly",
		},
	},
	{
		name:       "Find_build_related_commands",
		pathPrefix: "build",
		wantPaths: []string{
			"build",
			"builder",
			"building",
		},
	},
	{
		name:        "Non-existent_prefix",
		pathPrefix:  "nonexistent",
		wantPaths:   nil,
		wantErr:     true,
		errContains: "no commands found",
	},
	{
		name:        "Empty_prefix",
		pathPrefix:  "",
		wantErr:     true,
		errContains: "path prefix cannot be empty",
	},
} 

var PathsTestData = []string{
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