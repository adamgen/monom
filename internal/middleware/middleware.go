// Package middleware resolves the run middleware chain for one command: the
// executables monom() puts in front of a resolved command as
//
//	( exec mw1 mw2 ... /abs/command )
//
// The chain comes from the optional `middleware` hook of a hook-script
// config, called with the command's project-relative path. Each non-blank
// line of its output is one middleware: whitespace-separated words, the first
// naming the executable, the rest its own leading arguments (like the run
// hook's output, there is no quoting). The first line is the outermost.
//
// The first word resolves like this:
//
//   - "mnmd": the running mnmd binary, so built-in middleware such as
//     `mnmd cache` works whether or not mnmd is on PATH.
//   - contains a "/": a path, relative to the project root unless absolute.
//   - otherwise: looked up on PATH.
//
// A missing or non-executable middleware is an error: monom does not run the
// command without the middleware the project asked for.
package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/config"
)

// Hook is the hook subcommand that prints a command's chain.
const Hook = "middleware"

// Chain returns the chain for the command at abs (an absolute path inside
// root) as argv words, the first word of each middleware resolved to an
// absolute path. No hook script, or a hook that prints nothing, is an empty
// chain.
func Chain(root, abs string) ([]string, error) {
	cfg, err := config.Load(filepath.Join(root, "monom"))
	if err != nil {
		return nil, err
	}
	if cfg.Kind != config.Script {
		return nil, nil
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return nil, fmt.Errorf("%s is not inside the project root %s", abs, root)
	}
	out, err := config.RunHook(cfg.Path, Hook, filepath.ToSlash(rel))
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("middleware hook failed (exit %d)%s", ee.ExitCode(), detail(err, ee))
		}
		return nil, fmt.Errorf("middleware hook failed: %w", err)
	}
	return Parse(root, out)
}

// detail returns the hook's stderr that RunHook appended to the error, on a
// new line, or "".
func detail(err error, ee *exec.ExitError) string {
	msg := strings.TrimPrefix(err.Error(), ee.Error())
	msg = strings.TrimPrefix(msg, ": ")
	if msg == "" {
		return ""
	}
	return "\n" + msg
}

// Parse turns the hook's output into argv words, resolving each line's first
// word.
func Parse(root string, out []byte) ([]string, error) {
	var words []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		exe, err := resolve(root, fields[0])
		if err != nil {
			return nil, fmt.Errorf("middleware %q: %w", fields[0], err)
		}
		words = append(append(words, exe), fields[1:]...)
	}
	return words, sc.Err()
}

func resolve(root, name string) (string, error) {
	if name == "mnmd" {
		return os.Executable()
	}
	if !strings.Contains(name, "/") {
		p, err := exec.LookPath(name)
		if err != nil {
			return "", errors.New("not found on PATH")
		}
		return filepath.Abs(p)
	}
	p := name
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	fi, err := os.Stat(p)
	switch {
	case err != nil:
		return "", errors.New("no such file")
	case !fi.Mode().IsRegular():
		return "", errors.New("not a regular file")
	case fi.Mode()&0o111 == 0:
		return "", errors.New("not executable")
	}
	return p, nil
}
