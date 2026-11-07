package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/debuglog"
	"github.com/adamgen/monom/internal/root"
	"github.com/adamgen/monom/internal/runcache"
)

// runCache dispatches `mnmd cache run <abs-path> [args...]`, the run
// middleware monom() calls instead of exec when monom-cache.json exists.
func runCache() error {
	if len(os.Args) < 4 || os.Args[2] != "run" {
		return cli.WrapError(errors.New("usage: mnmd cache run <path> [args...]"))
	}
	return cacheRun(os.Args[3], os.Args[4:])
}

// cacheRun runs one resolved command through the cache middleware:
//
//   - no entry for the command, or MONOM_CACHE=off: exec it, exactly as
//     monom() would have.
//   - before: key the run; a hit restores outputs, replays stdout then
//     stderr, and exits 0 without running the command. A key that can't be
//     computed (a missing input) bypasses the cache with a notice.
//   - after: run the command with its output teed; store the result only
//     when it exits 0 and every declared output exists. A failed store is a
//     notice, never a change to the command's exit status.
func cacheRun(abs string, args []string) error {
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return cli.WrapError(err)
	}
	entries, err := runcache.Load(projectRoot)
	if err != nil {
		return cli.WrapError(err)
	}
	rel, err := filepath.Rel(projectRoot, abs)
	if err != nil {
		return execCommand(abs, args)
	}
	rel = filepath.ToSlash(rel)
	entry, ok := entries[rel]
	if !ok || os.Getenv("MONOM_CACHE") == "off" {
		return execCommand(abs, args)
	}

	key, err := runcache.Key(projectRoot, rel, args, entry, os.LookupEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "monom: cache bypassed for %s: %v\n", rel, err)
		return execCommand(abs, args)
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
	cmd := exec.Command(abs, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errb)
	// The terminal delivers Ctrl-C to the command too; mnmd stays to report
	// its status instead of dying first.
	signal.Ignore(syscall.SIGINT, syscall.SIGQUIT)
	status := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return cli.WrapError(err)
		}
		status = ee.ExitCode()
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			status = 128 + int(ws.Signal())
		}
	}
	if status == 0 {
		if err := runcache.Store(projectRoot, key, entry, runcache.Result{Stdout: out.Bytes(), Stderr: errb.Bytes()}); err != nil {
			fmt.Fprintf(os.Stderr, "monom: not cached: %v\n", err)
		}
	}
	os.Exit(status)
	return nil
}

// execCommand replaces mnmd with the command: no middleware, no resident
// parent, the same as monom()'s own exec.
func execCommand(abs string, args []string) error {
	err := syscall.Exec(abs, append([]string{abs}, args...), os.Environ())
	return cli.WrapError(fmt.Errorf("exec %s: %w", abs, err))
}
