package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adamgen/monom/src/go_utils"
)

func run() error {
	// Check if we have any arguments
	if len(os.Args) < 2 {
		fmt.Printf("monom tools command\n")
		return nil
	}

	command := os.Args[1]
	if command == "complete" {
		var commandPath string
		if len(os.Args) > 2 {
			commandPath = os.Args[2]
		}

		// Read from stdin
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		stdinStr := strings.TrimSpace(string(stdinBytes))
		commands := strings.Split(stdinStr, "\n")

		// Use test_projects directory for finding commands
		matches, err := go_utils.FindCommands(commands, commandPath)
		if err != nil {
			return err
		}

		// Print matches
		for _, match := range matches {
			fmt.Println(match)
		}
		return nil
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
