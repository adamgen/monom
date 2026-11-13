package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/root"
)

// runMiddleware is `mnmd middleware <abs-command-path>`: runs the config's
// `middleware` hook with the command's project-relative path and prints the
// chain it names, one argv word per line, for monom() to exec in front of the
// command. Each hook line is whitespace-separated words; the first word is
// the executable: "mnmd" is this binary, a word with a "/" is a path
// relative to the project root, anything else is looked up on PATH.
func runMiddleware() error {
	if len(os.Args) != 3 {
		return cli.WrapError(errors.New("usage: mnmd middleware <command-path>"))
	}
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		return cli.WrapError(err)
	}
	cfg, err := config.Load(filepath.Join(projectRoot, "monom"))
	if err != nil {
		return cli.WrapError(err)
	}
	if cfg.Kind != config.Script {
		return nil // no hook script: no chain
	}
	rel, _ := filepath.Rel(projectRoot, os.Args[2])
	out, err := config.RunHook(cfg.Path, "middleware", filepath.ToSlash(rel))
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			msg := strings.TrimPrefix(strings.TrimPrefix(err.Error(), ee.Error()), ": ")
			return cli.WrapError(fmt.Errorf("middleware hook failed (exit %d)\n%s", ee.ExitCode(), msg))
		}
		return cli.WrapError(err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		exe, err := resolveMiddleware(projectRoot, words[0])
		if err != nil {
			return cli.WrapError(fmt.Errorf("middleware %q: %w", words[0], err))
		}
		fmt.Println(strings.Join(append([]string{exe}, words[1:]...), "\n"))
	}
	return nil
}

func resolveMiddleware(projectRoot, name string) (string, error) {
	switch {
	case name == "mnmd":
		return os.Executable()
	case !strings.Contains(name, "/"):
		return exec.LookPath(name)
	}
	if !filepath.IsAbs(name) {
		name = filepath.Join(projectRoot, name)
	}
	if fi, err := os.Stat(name); err != nil || !fi.Mode().IsRegular() || fi.Mode()&0o111 == 0 {
		return "", errors.New("not an executable file")
	}
	return name, nil
}
