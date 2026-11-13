package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/debuglog"
	"github.com/adamgen/monom/internal/middleware"
	"github.com/adamgen/monom/internal/root"
)

// runMiddleware is `mnmd middleware <abs-command-path>`: prints the run
// middleware chain for one resolved command, one argv word per line, for
// monom() to put in front of the command. Nothing printed means no chain.
func runMiddleware() error {
	if len(os.Args) != 3 {
		return cli.WrapError(errors.New("usage: mnmd middleware <command-path>"))
	}
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return cli.WrapError(err)
	}
	abs, err := filepath.Abs(os.Args[2])
	if err != nil {
		return cli.WrapError(err)
	}
	words, err := middleware.Chain(projectRoot, abs)
	if err != nil {
		return cli.WrapError(err)
	}
	debuglog.Log("[mnmd middleware] %s chain=%q", abs, words)
	for _, w := range words {
		fmt.Println(w)
	}
	return nil
}
