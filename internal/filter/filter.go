package filter

import "strings"

// Filter returns the next-level completion tokens from commands matching the
// given words prefix.
//
// commands is a slice of slash-delimited paths (e.g. "category/sub_command").
// words are the raw space-separated tokens the user has typed so far. They are
// joined with "/" internally to build the matching prefix.
//
// A trailing empty string in words signals a completed token (user pressed Tab
// after a space) and causes Filter to drill into the children of that level.
//
// Filter never returns an error and never panics — on any unexpected condition
// it returns an empty slice.
func Filter(commands []string, words []string) (result []string) {
	defer func() {
		if recover() != nil {
			result = []string{}
		}
	}()

	// Split words into the "completed" path segments and the current partial word.
	// A trailing empty string means the last token was completed and we want children.
	//
	// Examples:
	//   []               → completedDepth=0, partial=""
	//   ["cat"]          → completedDepth=0, partial="cat"
	//   ["cat", ""]      → completedDepth=1, partial=""   (drill into "cat")
	//   ["cat", "sub"]   → completedDepth=1, partial="sub"
	//   ["cat", "sub",""]→ completedDepth=2, partial=""

	var completedWords []string
	partial := ""

	if len(words) > 0 {
		last := words[len(words)-1]
		if last == "" {
			// All words are completed segments; we want children.
			completedWords = words[:len(words)-1]
		} else {
			// Last word is a partial; completed words are everything before it.
			completedWords = words[:len(words)-1]
			partial = last
		}
	}

	seen := map[string]bool{}
	for _, cmd := range commands {
		if hasSpaceInSegment(cmd) {
			continue
		}

		parts := strings.Split(cmd, "/")

		// The command must have enough segments to match all completed words.
		if len(parts) <= len(completedWords) {
			continue
		}

		// Each completed word must exactly match the corresponding segment.
		match := true
		for i, w := range completedWords {
			if parts[i] != w {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		// The token at the next level must start with the partial.
		token := parts[len(completedWords)]
		if !strings.HasPrefix(token, partial) {
			continue
		}

		if !seen[token] {
			seen[token] = true
			result = append(result, token)
		}
	}

	if result == nil {
		return []string{}
	}
	return result
}

// hasSpaceInSegment reports whether any slash-delimited segment of path
// contains a space character.
func hasSpaceInSegment(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if strings.ContainsRune(seg, ' ') {
			return true
		}
	}
	return false
}
