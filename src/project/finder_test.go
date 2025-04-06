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

			if err != nil {
				t.Errorf("FindCommands() unexpected error = %v", err)
				return
			}

			// Sort both slices for comparison
			sort.Strings(got)
			sort.Strings(tt.wantPaths)

			if !reflect.DeepEqual(got, tt.wantPaths) && (len(got) + len(tt.wantPaths) != 0) {
				t.Errorf("\nFindCommands()\nis   = %v\nwant = %v", got, tt.wantPaths)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
