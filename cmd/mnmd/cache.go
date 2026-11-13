package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/debuglog"
	"github.com/adamgen/monom/internal/proc"
	"github.com/adamgen/monom/internal/root"
	"github.com/adamgen/monom/internal/runcache"
)

// runCache is `mnmd cache <next...>`: the cache as an ordinary run
// middleware. <next...> is the rest of the chain, ending in the command; the
// command's identity comes from MONOM_COMMAND_PATH, which monom() sets for
// every chain.
//
//   - no entry for the command (or no monom-cache.json, or MONOM_CACHE=off,
//     or not run by monom): exec next. Nothing stays resident.
//   - a hit: restore outputs, replay stdout then stderr, exit 0. Next is not
//     called.
//   - a miss: run next, tee its output, store the result only when it exits
//     0 and every declared output exists; exit with its status.
func runCache() error {
	next := os.Args[2:]
	if len(next) == 0 {
		return cli.WrapError(errors.New("usage: mnmd cache <next> [args...]  (a run middleware: see the middleware hook)"))
	}
	pass := func() error { return cli.WrapError(proc.Exec(next)) }

	rel := os.Getenv("MONOM_COMMAND_PATH")
	if rel == "" || os.Getenv("MONOM_CACHE") == "off" {
		return pass()
	}
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return cli.WrapError(err)
	}
	entries, err := runcache.Load(projectRoot)
	if err != nil {
		return cli.WrapError(err)
	}
	entry, ok := entries[rel]
	if !ok {
		return pass()
	}

	key, err := runcache.Key(projectRoot, rel, next, entry, os.LookupEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "monom: cache bypassed for %s: %v\n", rel, err)
		return pass()
	}
	debuglog.Log("[mnmd cache] %s key=%s", rel, key)

	if res, hit := runcache.Lookup(projectRoot, key); hit {
		if err := runcache.Restore(projectRoot, key, entry); err != nil {
			fmt.Fprintf(os.Stderr, "monom: cache bypassed for %s: %v\n", rel, err)
		} else {
			fmt.Fprintf(os.Stderr, "monom: cache hit for %s (%s)\n", rel, key[:12])
			os.Stdout.Write(res.Stdout)
			os.Stderr.Write(res.Stderr)
			return nil
		}
	}

	var out, errb bytes.Buffer
	status, err := proc.Run(next, io.MultiWriter(os.Stdout, &out), io.MultiWriter(os.Stderr, &errb))
	if err != nil {
		return cli.WrapError(err)
	}
	if status == 0 {
		if err := runcache.Store(projectRoot, key, entry, runcache.Result{Stdout: out.Bytes(), Stderr: errb.Bytes()}); err != nil {
			fmt.Fprintf(os.Stderr, "monom: not cached: %v\n", err)
		}
	}
	os.Exit(status)
	return nil
}
