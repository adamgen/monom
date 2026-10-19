package discover

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/config"
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

// isScript reports whether p is an executable file with a shebang: the only
// files suffix stripping applies to. A suffixed name never matches
// NamePattern, so a script is the only suffixed file the scan registers, and
// pack can tell one apart from the file alone, without reading the config.
func isScript(p string) bool {
	return isExecutableFile(p) && config.HasShebang(p)
}

// StemMatches returns the scripts in dir whose stem is stem, sorted. Pack uses
// it to resolve a command name that suffix stripping produced; one match is
// the command, more than one is ambiguous.
func StemMatches(dir, stem string) []string {
	entries, _ := os.ReadDir(dir)
	var out []string
	for _, e := range entries {
		if Stem(e.Name()) == stem && isScript(filepath.Join(dir, e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out
}

// dirIndex is what suffix stripping needs to know about one directory.
type dirIndex struct {
	entries map[string]os.DirEntry // every entry, by name
	stems   map[string][]string    // scripts with a suffix, by stem, sorted
}

func readDirIndex(dir string) dirIndex {
	idx := dirIndex{entries: map[string]os.DirEntry{}, stems: map[string][]string{}}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		idx.entries[e.Name()] = e
		if s := Stem(e.Name()); s != "" && isScript(filepath.Join(dir, e.Name())) {
			idx.stems[s] = append(idx.stems[s], e.Name())
		}
	}
	return idx
}

// stripSuffixes gives each registered script its command name, its stem, and
// returns the commands that remain plus one message per name it could not
// give. Two scripts sharing a stem keep their suffixes, as pack cannot choose
// between them; a script whose stem names an existing entry is not registered,
// as pack resolves that name to the entry. Each directory is read once.
func stripSuffixes(projectRoot string, commands []Command) ([]Command, []string) {
	dirs := map[string]dirIndex{}
	reported := map[string]bool{}
	var kept []Command
	var conflicts []string
	for _, c := range commands {
		dir, name := path.Split(c.Path)
		s := Stem(name)
		if s == "" {
			kept = append(kept, c)
			continue
		}
		idx, ok := dirs[dir]
		if !ok {
			idx = readDirIndex(filepath.Join(projectRoot, filepath.FromSlash(dir)))
			dirs[dir] = idx
		}
		scripts := idx.stems[s]
		taken, isTaken := idx.entries[s]
		switch {
		case !contains(scripts, name):
			kept = append(kept, c) // not a script: a declared file keeps its name
		case len(scripts) > 1:
			kept = append(kept, c)
			if !reported[dir+s] {
				reported[dir+s] = true
				paths := make([]string, len(scripts))
				for i, f := range scripts {
					paths[i] = dir + f
				}
				conflicts = append(conflicts, fmt.Sprintf("%s share the command name %s, so each keeps its suffix; rename all but one", strings.Join(paths, ", "), dir+s))
			}
		case isTaken:
			kind := "file"
			if taken.IsDir() {
				kind = "directory"
			}
			conflicts = append(conflicts, fmt.Sprintf("%s is not registered: its command name %s is taken by the %s %s; rename one of them", c.Path, dir+s, kind, dir+s))
		default:
			c.Path = dir + s
			kept = append(kept, c)
		}
	}
	return kept, conflicts
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
