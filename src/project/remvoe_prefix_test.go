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
			name:    "Empty prefix should return original command",
			command: "project1/command_1",
			prefix:  "",
			want:    "project1/command_1",
		},
		{
			name:    "Partial project prefix should not affect command",
			command: "project1/command_1",
			prefix:  "proj",
			want:    "project1/command_1",
		},
		{
			name:    "Project name without slash should not affect command",
			command: "project1/command_1",
			prefix:  "project1",
			want:    "project1/command_1",
		},
		{
			name:    "Project name with slash should remove project prefix",
			command: "project1/command_1",
			prefix:  "project1/",
			want:    "command_1",
		},
		{
			name:    "Partial command prefix should return remaining command",
			command: "project1/command_1",
			prefix:  "project1/c",
			want:    "command_1",
		},
		{
			name:    "Full command as prefix should return command name",
			command: "project1/command_1",
			prefix:  "project1/command_1",
			want:    "command_1",
		},
		{
			name:    "Deep path - prefix before first slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project",
			want:    "project1/subdir/nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix at first slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix after first slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/sub",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix before second slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subd",
			want:    "subdir/nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix at second slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix after second slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nest",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix before third slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/neste",
			want:    "nested/deep/command_1",
		},
		{
			name:    "Deep path - prefix at third slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/",
			want:    "deep/command_1",
		},
		{
			name:    "Deep path - prefix after third slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/de",
			want:    "deep/command_1",
		},
		{
			name:    "Deep path - prefix before fourth slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/dee",
			want:    "deep/command_1",
		},
		{
			name:    "Deep path - prefix at fourth slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/deep/",
			want:    "command_1",
		},
		{
			name:    "Deep path - prefix after fourth slash",
			command: "project1/subdir/nested/deep/command_1",
			prefix:  "project1/subdir/nested/deep/com",
			want:    "command_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemovePrefix(tt.command, tt.prefix)
			if err != nil {
				t.Errorf("RemovePrefix() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("RemovePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
