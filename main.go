package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func findMonomRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	for {
		monomFile := filepath.Join(dir, "monom")
		if _, err := os.Stat(monomFile); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("monom file not found in any parent directory")
		}
		dir = parent
	}
}

func run() error {
	projectRoot := os.Getenv("MONOM_PROJECT_ROOT")
	if projectRoot == "" {
		// Try to find monom file in parent directories
		root, err := findMonomRoot()
		if err != nil {
			return fmt.Errorf("monom project root not set and %v", err)
		}
		projectRoot = root
		if err := os.Setenv("MONOM_PROJECT_ROOT", projectRoot); err != nil {
			return fmt.Errorf("error setting MONOM_PROJECT_ROOT: %v", err)
		}
	}

	info, err := os.Stat(projectRoot)
	if err != nil {
		return fmt.Errorf("error accessing monom project root path: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("monom project root path '%s' is not a directory", projectRoot)
	}

	monomFile := filepath.Join(projectRoot, "monom")
	if _, err := os.Stat(monomFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("monom file not found in project root")
		}
		return fmt.Errorf("error accessing monom file: %v", err)
	}

	fmt.Printf("MONOM_PROJECT_ROOT: %s\n", projectRoot)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
