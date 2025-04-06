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
