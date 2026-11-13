package main

import (
	"errors"
	"os"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/root"
	"github.com/adamgen/monom/internal/runcache"
)

// runCache is `mnmd cache <next...>`, the cache as a run middleware. Run
// outside a monom chain (no MONOM_COMMAND_PATH) it just execs next.
func runCache() error {
	next := os.Args[2:]
	if len(next) == 0 {
		return cli.WrapError(errors.New("usage: mnmd cache <next> [args...]"))
	}
	cmd := os.Getenv("MONOM_COMMAND_PATH")
	if cmd == "" {
		return cli.WrapError(runcache.Exec(next))
	}
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return cli.WrapError(err)
	}
	status, err := runcache.Middleware(projectRoot, cmd, next)
	if err != nil {
		return cli.WrapError(err)
	}
	os.Exit(status)
	return nil
}
