package discover

import "testing"

func TestStem_OnlyTheLastSuffixAndNeverALeadingOrTrailingDot(t *testing.T) {
	cases := map[string]string{
		"setup.sh":    "setup",
		"pkg.tar.gz":  "pkg.tar",
		"Weird.BIN":   "Weird",
		"release":     "",
		".env":        "",
		"x.":          "",
		"..":          "",
		"a.b":         "a",
		"run_it.d.sh": "run_it.d",
	}
	for name, want := range cases {
		if got := Stem(name); got != want {
			t.Errorf("Stem(%q) = %q, want %q", name, got, want)
		}
	}
}
