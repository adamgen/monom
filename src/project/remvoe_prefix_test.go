package project

import (
	"testing"
)

func TestRemovePrefix(t *testing.T) {
	tests := []struct {
		name    string
		command string
		prefix  string
		want    string
	}{
		{
			name:    "Empty_prefix_should_return_original_command",
			command: "project1/command_1",
			prefix:  "",
			want:    "project1/command_1",
		},
		{
			name:    "Partial_project_prefix_should_not_affect_command",
			command: "project1/command_1",
			prefix:  "proj",
			want:    "project1/command_1",
		},
		{
			name:    "Project_name_without_slash_should_not_affect_command",
			command: "project1/command_1",
			prefix:  "project1",
			want:    "project1/command_1",
		},
		{
			name:    "Project_name_with_slash_should_remove_project_prefix",
			command: "project1/command_1",
			prefix:  "project1/",
			want:    "command_1",
		},
		{
			name:    "Partial_command_prefix_should_return_remaining_command",
			command: "project1/command_1",
			prefix:  "project1/c",
			want:    "command_1",
		},
		{
			name:    "Full_command_as_prefix_should_return_command_name",
			command: "project1/command_1",
			prefix:  "project1/command_1",
			want:    "command_1",
		},
		{
			name:    "Deep_path_prefix_before_first_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project",
			want:    "project1/subdir/nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_at_first_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_after_first_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/sub",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_before_second_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subd",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_at_second_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_after_second_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nest",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_before_third_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/neste",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep_path_prefix_at_third_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/",
			want:    "deep/command_1",
		},
		{
			name:    "Deep_path_prefix_after_third_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/de",
			want:    "deep/command_1",
		},
		{
			name:    "Deep_path_prefix_before_fourth_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/dee",
			want:    "deep/command_1",
		},
		{
			name:    "Deep_path_prefix_at_fourth_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/deep/",
			want:    "command_1",
		},
		{
			name:    "Deep_path_prefix_after_fourth_slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/deep/com",
			want:    "command_1",
		},
		{
			name:    "Deep_path_unrelated_prefix",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "a",
			want:    "",
		},
		{
			name:    "ADD_NAME",
			command: "api-v1/auth",
			prefix:  "pro",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemovePrefix(tt.command, tt.prefix)
			if err != nil {
				t.Errorf("\nRemovePrefix() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("\nRemovePrefix()\nis   = %v\nwant = %v", got, tt.want)
			}
		})
	}
}
