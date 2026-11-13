package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/root"
	"github.com/adamgen/monom/internal/runcache"
)

// runCache is `mnmd cache <next...>`, the cache as a run middleware. Commands
// without a cache.inputs setting, or run outside a monom chain (no
// MONOM_COMMAND_PATH), pass straight through to next.
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
	cfg, err := config.Load(filepath.Join(projectRoot, "monom"))
	if err != nil {
		return cli.WrapError(err)
	}
	settings, err := cfg.StrictSettings()
	if err != nil {
		return cli.WrapError(err)
	}
	rule, ok := settings.Cache[cmd]
	if !ok || rule.Inputs == "" {
		return cli.WrapError(runcache.Exec(next))
	}
	status, err := runcache.Middleware(projectRoot, cmd, rule, next)
	if err != nil {
		return cli.WrapError(err)
	}
	os.Exit(status)
	return nil
}
