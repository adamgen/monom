package discover

import (
	"os"
	"path/filepath"
	"strings"
)

// Stem returns name without its suffix — the last "." and what follows it —
// or "" when name has none. A leading dot (".env") or a trailing one ("x.")
// is not a suffix.
func Stem(name string) string {
	i := strings.LastIndexByte(name, '.')
	if i <= 0 || i == len(name)-1 {
		return ""
	}
	return name[:i]
}

// StemMatches returns the executable files in dir whose stem is stem, sorted.
// Pack uses it to resolve a command name that suffix stripping produced; one
// match is the command, more than one is ambiguous.
func StemMatches(dir, stem string) []string {
	entries, _ := os.ReadDir(dir)
	var out []string
	for _, e := range entries {
		if Stem(e.Name()) == stem && isExecutableFile(filepath.Join(dir, e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out
}

// dirIndex is what suffix stripping needs to know about one directory.
type dirIndex struct {
	names map[string]bool // every entry
	stems map[string]int  // executable files with a suffix, counted by stem
}

func readDirIndex(dir string) dirIndex {
	idx := dirIndex{names: map[string]bool{}, stems: map[string]int{}}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		idx.names[e.Name()] = true
		if s := Stem(e.Name()); s != "" && isExecutableFile(filepath.Join(dir, e.Name())) {
			idx.stems[s]++
		}
	}
	return idx
}

// commandName returns the command name of the file name in idx's directory:
// its stem, unless that would be ambiguous — an entry is already named the
// stem, or another executable there has the same stem — in which case the
// name keeps its suffix. These are exactly the cases where pack, given the
// stem, would not resolve to this file.
func (idx dirIndex) commandName(name string) string {
	s := Stem(name)
	if s == "" || idx.names[s] || idx.stems[s] != 1 {
		return name
	}
	return s
}
