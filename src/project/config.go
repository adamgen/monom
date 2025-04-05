package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the configuration for the monom project
type Config struct {
	RootPath string
}

// FindMonomRoot searches for a monom file in the current directory and its parents
// until it finds one or reaches the root directory.
func FindMonomRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		if IsProjectRoot(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("monom file not found in any parent directory")
		}
		dir = parent
	}
}

// IsProjectRoot checks if the given directory is a monom project root
// by verifying the existence of a monom file.
func IsProjectRoot(dir string) bool {
	monomFile := filepath.Join(dir, "monom")
	_, err := os.Stat(monomFile)
	return err == nil
}

// LoadConfig loads the project configuration from environment
// or tries to detect it automatically.
func LoadConfig() (*Config, error) {
	projectRoot := os.Getenv("MONOM_PROJECT_ROOT")
	var err error

	if projectRoot == "" {
		projectRoot, err = FindMonomRoot()
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

	return &Config{
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