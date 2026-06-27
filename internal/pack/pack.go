package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/root"
)

// GroupError is returned by Pack when the user's tokens resolve to a directory
// (a command group) rather than an executable file. A directory is not a
// failure: it is a noun in monom's noun→verb file tree, but it is not itself a
// runnable command.
//
// GroupError carries only the resolved path — deliberately not a child listing.
// Discovery is the job of the `complete` hook (see terminology.md), so the set
// of commands under a group is sourced from `complete | mnmd filter`, never from
// pack reading the filesystem. Keeping pack a pure resolver means it returns an
// error xor a value, and the command tree has a single source of truth.
type GroupError struct {
	Path string // absolute path of the resolved group directory
}

func (e *GroupError) Error() string {
	return "resolved path is a command group: " + e.Path
}

// Pack resolves a space-separated command token sequence to an absolute
// executable path within the monom project.
//
// words are the space-separated CLI arguments the user typed (e.g. ["category",
// "sub_command"]). They are joined with "/" to form a relative file path, which
// is then resolved against the project root discovered by FindProjectRoot.
//
// Returns an error if words is empty, no project root is found, the resolved
// path does not exist, or the file is not executable. When the resolved path is
// a directory, it returns a *GroupError — a distinct outcome signalling "this is
// a command group", not a generic failure. Pack does not enumerate the group's
// children; that is the caller's job via the `complete` discovery hook.
func Pack(words []string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("pack: no command tokens provided")
	}

	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return "", fmt.Errorf("pack: %w", err)
	}

	relPath := strings.Join(words, "/")
	absPath := filepath.Join(projectRoot, relPath)

	fi, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("pack: command not found: %s", absPath)
		}
		return "", fmt.Errorf("pack: cannot stat %s: %w", absPath, err)
	}
	if fi.IsDir() {
		return "", &GroupError{Path: absPath}
	}
	if fi.Mode()&0o111 == 0 {
		return "", fmt.Errorf("pack: command is not executable: %s", absPath)
	}

	return absPath, nil
}
