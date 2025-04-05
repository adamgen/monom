package main

import (
	"fmt"
	"os"

	"github.com/adamgen/monom/src/project"
)

func run() error {
	config, err := project.LoadConfig()
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

