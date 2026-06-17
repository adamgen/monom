package root

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindProjectRoot returns the absolute path of the nearest monom project root.
//
// It first checks $_MONOM_PROJECT_ROOT: if set and the directory contains an
// executable file named "monom", that directory is returned without walking.
// Otherwise it walks upward from $PWD until it finds such a directory or
// reaches the filesystem root, in which case it returns an error.
func FindProjectRoot() (string, error) {
	if envRoot := os.Getenv("_MONOM_PROJECT_ROOT"); envRoot != "" {
		if isValidProjectRoot(envRoot) {
			resolved, err := filepath.EvalSymlinks(envRoot)
			if err != nil {
				return "", err
			}
			return resolved, nil
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}

	// Resolve symlinks in pwd so the walk and returned path are consistent.
	pwd, err = filepath.EvalSymlinks(pwd)
	if err != nil {
		return "", fmt.Errorf("cannot resolve working directory: %w", err)
	}

	dir := pwd
	for {
		if isValidProjectRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no monom project root found (no executable 'monom' file in %s or any parent)", pwd)
}

// isValidProjectRoot reports whether dir contains an executable file named "monom".
func isValidProjectRoot(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	monomPath := filepath.Join(dir, "monom")
	fi, err := os.Stat(monomPath)
	if err != nil || fi.IsDir() {
		return false
	}
	return fi.Mode()&0o111 != 0
}
