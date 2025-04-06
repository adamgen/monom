package project

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestFindCommands(t *testing.T) {
	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindCommands(PathsTestData, tt.pathPrefix)

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
				t.Errorf("\nFindCommands()\nis   = %v\nwant = %v", got, tt.wantPaths)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
