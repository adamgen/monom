package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindCommands searches for command paths in the given root directory that match the provided prefix.
// It returns a list of paths relative to the root directory that:
// 1. Match the given prefix
// 2. Contain a 'run' file
// The function returns an error if:
// - The prefix is empty
// - No matching commands are found
// - There's an error accessing the filesystem
func FindCommands(root, pathPrefix string) ([]string, error) {
	if pathPrefix == "" {
		return nil, fmt.Errorf("path prefix cannot be empty")
	}

	// Clean the prefix to handle trailing slashes
	pathPrefix = strings.TrimSuffix(pathPrefix, "/")

	// Use a map to deduplicate paths
	matchMap := make(map[string]bool)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == root {
			return nil
		}

		// Get the relative path from the root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Check if this is a 'run' file
		if !info.IsDir() && info.Name() == "run" {
			// Get the directory containing the run file
			cmdPath := filepath.Dir(relPath)

			// Check if this command path or any of its parent directories match our prefix
			if strings.HasPrefix(cmdPath, pathPrefix) {
				// Add all directories in the path that contain a run file
				matchMap[cmdPath] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	// Convert map to slice
	var matches []string
	for match := range matchMap {
		matches = append(matches, match)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no commands found with prefix '%s'", pathPrefix)
	}

	return matches, nil
} 