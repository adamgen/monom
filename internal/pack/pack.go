package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/root"
)

// Pack resolves a space-separated command token sequence to an absolute
// executable path within the monom project.
//
// words are the space-separated CLI arguments the user typed (e.g. ["category",
// "sub_command"]). They are joined with "/" to form a relative file path, which
// is then resolved against the project root discovered by FindProjectRoot.
//
// Returns an error if words is empty, no project root is found, the resolved
// path does not exist, or the file is not executable.
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
		return "", fmt.Errorf("pack: resolved path is a directory, not a command: %s", absPath)
	}
	if fi.Mode()&0o111 == 0 {
		return "", fmt.Errorf("pack: command is not executable: %s", absPath)
	}

	return absPath, nil
}
