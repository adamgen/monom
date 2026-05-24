package filter

import (
	"sort"
	"testing"
)

func sortedEqual(t *testing.T, got, want []string) {
	t.Helper()
	a := append([]string(nil), got...)
	b := append([]string(nil), want...)
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("got %v, want %v", got, want)
			return
		}
	}
}

func TestFilter_NoWordsReturnsAllTopLevelTokens(t *testing.T) {
	commands := []string{
		"category1/sub1",
		"category1/sub2",
		"command1",
		"command2",
	}
	got := Filter(commands, nil)
	sortedEqual(t, got, []string{"category1", "command1", "command2"})
}

func TestFilter_PartialWordMatchesAtTopLevel(t *testing.T) {
	commands := []string{"command1", "command2", "category1/sub1"}
	got := Filter(commands, []string{"com"})
	sortedEqual(t, got, []string{"command1", "command2"})
}

func TestFilter_PartialCategoryWordReturnsCategoryToken(t *testing.T) {
	commands := []string{"category1/sub1", "category1/sub2", "command1"}
	got := Filter(commands, []string{"categ"})
	sortedEqual(t, got, []string{"category1"})
}

func TestFilter_CompleteCategoryPlusEmptyDrillsIntoChildren(t *testing.T) {
	commands := []string{
		"category1/sub_command1",
		"category1/sub_command2",
		"command1",
	}
	got := Filter(commands, []string{"category1", ""})
	sortedEqual(t, got, []string{"sub_command1", "sub_command2"})
}

func TestFilter_PartialWordWithinCategory(t *testing.T) {
	commands := []string{
		"category1/sub_command1",
		"category1/sub_command2",
	}
	got := Filter(commands, []string{"category1", "sub_c"})
	sortedEqual(t, got, []string{"sub_command1", "sub_command2"})
}

func TestFilter_NestedDrillDown(t *testing.T) {
	commands := []string{
		"infra/cloud/deploy",
		"infra/cloud/teardown",
		"infra/local/start",
	}
	got := Filter(commands, []string{"infra", "cloud", ""})
	sortedEqual(t, got, []string{"deploy", "teardown"})
}

func TestFilter_NoMatchesReturnsEmpty(t *testing.T) {
	commands := []string{"command1", "command2"}
	got := Filter(commands, []string{"xyz"})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestFilter_NonExistentChildOfExistingCategory(t *testing.T) {
	commands := []string{"category1/sub1", "category1/sub2"}
	got := Filter(commands, []string{"category1", "xyz"})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestFilter_DrillingIntoNonExistentCategory(t *testing.T) {
	commands := []string{"category1/sub1", "command1"}
	got := Filter(commands, []string{"nonexistent", ""})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestFilter_DuplicatesAreDeduped(t *testing.T) {
	commands := []string{"category1/sub1", "category1/sub2"}
	got := Filter(commands, nil)
	if len(got) != 1 || got[0] != "category1" {
		t.Errorf("expected [category1], got %v", got)
	}
}

func TestFilter_PathWithSpaceInSegmentSilentlyExcluded(t *testing.T) {
	commands := []string{"my command/sub", "command1", "command2"}
	got := Filter(commands, nil)
	sortedEqual(t, got, []string{"command1", "command2"})
}

func TestFilter_AllInvalidPathsReturnsEmpty(t *testing.T) {
	commands := []string{"my command/sub", "bad path/x"}
	got := Filter(commands, nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestFilter_NeverPanics(t *testing.T) {
	// Should not panic with nil commands or nil words.
	_ = Filter(nil, nil)
	_ = Filter(nil, []string{"x"})
	_ = Filter([]string{}, nil)
}
