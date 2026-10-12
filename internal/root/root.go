// Package root determines the active monom project root.
package root

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFileName is the name of the monom config file that marks a project
// root. The file may be empty, declarative, or an executable hook script; its
// mere presence is what makes a directory a root.
const ConfigFileName = "monom"

// FindProjectRoot returns the absolute path of the active monom project root.
//
// Resolution order, first match wins:
//
//  1. $_MONOM_PROJECT_ROOT, when it names an existing directory. This is the
//     explicit pin an alias sets; the directory needs no monom config file.
//  2. The nearest directory, walking upward from $PWD, that contains a regular
//     file named "monom" — executable or not, empty or not.
//  3. The nearest directory, walking upward from $PWD, that contains a ".git"
//     entry (a directory, or a file for worktrees and submodules).
//
// The working directory itself is never a fallback: a root must be pinned,
// marked, or version-controlled, so running monom outside any project cannot
// turn an arbitrary directory (such as $HOME) into a command tree.
func FindProjectRoot() (string, error) {
	if envRoot := os.Getenv("_MONOM_PROJECT_ROOT"); envRoot != "" && isDir(envRoot) {
		resolved, err := filepath.EvalSymlinks(envRoot)
		if err != nil {
			return "", err
		}
		return resolved, nil
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

	gitRoot := ""
	dir := pwd
	for {
		if HasConfigFile(dir) {
			return dir, nil
		}
		if gitRoot == "" && exists(filepath.Join(dir, ".git")) {
			gitRoot = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if gitRoot != "" {
		return gitRoot, nil
	}

	return "", fmt.Errorf("no monom project root found (no 'monom' file or git repository in %s or any parent, and $_MONOM_PROJECT_ROOT is unset)", pwd)
}

// HasConfigFile reports whether dir contains a regular file named "monom".
func HasConfigFile(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, ConfigFileName))
	return err == nil && fi.Mode().IsRegular()
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
