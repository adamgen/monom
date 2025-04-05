package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectConfig holds the configuration for the monom project
type ProjectConfig struct {
	RootPath string
}

// findMonomRoot searches for a monom file in the current directory and its parents
// until it finds one or reaches the root directory.
func findMonomRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		if isProjectRoot(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("monom file not found in any parent directory")
		}
		dir = parent
	}
}

// isProjectRoot checks if the given directory is a monom project root
// by verifying the existence of a monom file.
func isProjectRoot(dir string) bool {
	monomFile := filepath.Join(dir, "monom")
	_, err := os.Stat(monomFile)
	return err == nil
}

// loadProjectConfig loads the project configuration from environment
// or tries to detect it automatically.
func loadProjectConfig() (*ProjectConfig, error) {
	projectRoot := os.Getenv("MONOM_PROJECT_ROOT")
	var err error

	if projectRoot == "" {
		projectRoot, err = findMonomRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to detect project root: %w", err)
		}
		
		if err := os.Setenv("MONOM_PROJECT_ROOT", projectRoot); err != nil {
			return nil, fmt.Errorf("failed to set MONOM_PROJECT_ROOT: %w", err)
		}
	}

	if err := validateProjectRoot(projectRoot); err != nil {
		return nil, err
	}

	return &ProjectConfig{
		RootPath: projectRoot,
	}, nil
}

// validateProjectRoot ensures the project root path is valid and contains
// the required monom file.
func validateProjectRoot(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("invalid project root path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("project root path '%s' is not a directory", path)
	}

	monomFile := filepath.Join(path, "monom")
	if _, err := os.Stat(monomFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("monom file not found in project root")
		}
		return fmt.Errorf("failed to access monom file: %w", err)
	}

	return nil
}

func run() error {
	config, err := loadProjectConfig()
	if err != nil {
		return err
	}

	fmt.Printf("MONOM_PROJECT_ROOT: %s\n", config.RootPath)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

