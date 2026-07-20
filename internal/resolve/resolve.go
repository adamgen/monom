// Package resolve implements table-driven command mapping for the
// `mnmd resolve-run` and `mnmd resolve-complete` subcommands. One value-first
// mapping table drives both user-config hooks: Run resolves words for the
// `run` hook, Complete lists the keys for the `complete` hook.
package resolve

import (
	"errors"
	"strings"
)

// entry is one parsed mapping-table line: a value and the key words it maps from.
type entry struct {
	value string
	key   []string
}

// parse extracts the entries from a mapping table. Each entry line holds a
// value followed by the key words it maps from, separated by whitespace:
//
//	/path/to/scripts/migrate_db.sh    db migrate
//
// The value comes first because it can never contain whitespace, while the
// key may span any number of words. Fields are whitespace-delimited, so
// columns may be freely aligned. Blank lines and lines starting with "#" are
// skipped; so are lines with no key words (a bare value can never match).
func parse(table string) []entry {
	var entries []entry
	for _, line := range strings.Split(table, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		entries = append(entries, entry{value: fields[0], key: fields[1:]})
	}
	return entries
}

// Run resolves words for a `run` hook: it returns the value of the first
// entry whose key words equal words, or — when nothing matches — the words
// joined back into a single space-separated line. The passthrough makes a
// miss a non-event: the hook echoes the user's words unchanged and monom's
// default tree resolution takes over, so a run hook needs no fallback branch
// and no stderr silencing.
func Run(table string, words []string) (string, error) {
	if len(words) == 0 {
		return "", errors.New("no command words given")
	}
	for _, e := range parse(table) {
		if wordsEqual(e.key, words) {
			return e.value, nil
		}
	}
	return strings.Join(words, " "), nil
}

// Complete lists the table's keys for a `complete` hook: one slash-delimited
// path per entry ("db migrate" → "db/migrate"), in table order — the
// discovery format the completion pipeline expects. Mapped commands thus
// tab-complete exactly like commands living in the file tree.
func Complete(table string) []string {
	var keys []string
	for _, e := range parse(table) {
		keys = append(keys, strings.Join(e.key, "/"))
	}
	return keys
}

// wordsEqual reports whether two word slices are element-wise equal.
func wordsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
