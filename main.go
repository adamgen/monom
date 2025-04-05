package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adamgen/monom/src/project"
)

func run() error {
	config, err := project.LoadConfig()
	if err != nil {
		return err
	}

	// Check if we have any arguments
	if len(os.Args) < 2 {
		fmt.Printf("MONOM_PROJECT_ROOT: %s\n", config.RootPath)
		return nil
	}

	command := os.Args[1]
	if command == "complete" {
		// Check if we have a path prefix argument
		if len(os.Args) < 3 {
			return fmt.Errorf("path prefix argument required for complete command")
		}
		pathPrefix := os.Args[2]

		// Use test_projects directory for finding commands
		testProjectsPath := filepath.Join(config.RootPath, "test_projects")
		matches, err := project.FindCommands(testProjectsPath, pathPrefix)
		if err != nil {
			return err
		}

		// Print matches
		for _, match := range matches {
			fmt.Println(match)
		}
		return nil
	}

	fmt.Printf("MONOM_PROJECT_ROOT: %s\nCommand: %s\n", config.RootPath, command)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

