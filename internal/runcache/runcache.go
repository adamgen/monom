// Package runcache is monom's local, opt-in, input-based command cache, run
// as the `mnmd cache` middleware. monom-cache.json at the project root opts
// command files in, with explicit inputs, outputs and env vars:
//
//	{ "build": { "inputs": ["src"], "outputs": ["dist"], "env": ["MODE"] } }
//
// A hit restores the outputs and replays stdout then stderr without calling
// next; a miss runs next and stores only an exit-0 run whose outputs exist.
package runcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// FileName is the cache config at the project root; Dir holds entries.
const (
	FileName = "monom-cache.json"
	Dir      = ".monom/cache"
)

// Entry declares how one command is cached. Nothing is inferred.
type Entry struct {
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
	Env     []string `json:"env"`
}

// Load reads monom-cache.json; a missing file is (nil, nil).
func Load(root string) (map[string]Entry, error) {
	b, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var f map[string]Entry
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	for cmd, e := range f {
		if len(e.Inputs) == 0 {
			return nil, fmt.Errorf("%s: %q: inputs must list at least one path", FileName, cmd)
		}
		for _, p := range append(append([]string{cmd}, e.Inputs...), e.Outputs...) {
			if p == "" || path.IsAbs(p) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") || strings.HasPrefix(p, ".monom") {
				return nil, fmt.Errorf("%s: %q: %q must be a clean relative path inside the project, outside .monom", FileName, cmd, p)
			}
		}
	}
	return f, nil
}

// Key hashes what a result depends on: the command file, next's argv (the
// inner chain and the command), every input file's path, content and execute
// bit, each declared env var (only a hash of its value), and the entry. A
// missing input is an error: the caller bypasses the cache.
func Key(root, cmd string, next []string, e Entry) (string, error) {
	h := sha256.New()
	put := func(parts ...string) { fmt.Fprintf(h, "%s\n", strings.Join(parts, "\x00")) }
	entry, _ := json.Marshal(e)
	put("monom-cache v1", cmd, string(entry))
	put(next...)
	files := []string{cmd}
	for _, in := range e.Inputs {
		err := filepath.WalkDir(filepath.Join(root, in), func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				rel, _ := filepath.Rel(root, p)
				files = append(files, rel)
			}
			return err
		})
		if err != nil {
			return "", fmt.Errorf("input %q: %w", in, err)
		}
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(root, f))
		if err != nil {
			return "", err
		}
		info, _ := os.Stat(filepath.Join(root, f))
		sum := sha256.Sum256(b)
		put("file", f, fmt.Sprint(info.Mode()&0o111 != 0), hex.EncodeToString(sum[:]))
	}
	env := append([]string{}, e.Env...)
	sort.Strings(env)
	for _, n := range env {
		v, ok := os.LookupEnv(n)
		sum := sha256.Sum256([]byte(v))
		put("env", n, fmt.Sprint(ok), hex.EncodeToString(sum[:]))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Middleware is `mnmd cache <next...>`. It returns the exit status to use;
// a nil-error pass-through never returns (it execs next).
func Middleware(root, cmd string, next []string) (int, error) {
	entries, err := Load(root)
	if err != nil {
		return 1, err
	}
	e, ok := entries[cmd]
	if !ok || os.Getenv("MONOM_CACHE") == "off" {
		return 1, Exec(next)
	}
	key, err := Key(root, cmd, next, e)
	if err != nil {
		fmt.Fprintf(os.Stderr, "monom: cache bypassed for %s: %v\n", cmd, err)
		return 1, Exec(next)
	}
	dir := filepath.Join(root, Dir, key)
	if out, err := os.ReadFile(filepath.Join(dir, "stdout")); err == nil {
		errb, _ := os.ReadFile(filepath.Join(dir, "stderr"))
		if err := copyPaths(filepath.Join(dir, "outputs"), root, e.Outputs); err == nil {
			fmt.Fprintf(os.Stderr, "monom: cache hit for %s (%s)\n", cmd, key[:12])
			os.Stdout.Write(out)
			os.Stderr.Write(errb)
			return 0, nil
		}
		os.RemoveAll(dir) // damaged: a miss that replaces it
	}
	var out, errb bytes.Buffer
	status, err := Run(next, io.MultiWriter(os.Stdout, &out), io.MultiWriter(os.Stderr, &errb))
	if err == nil && status == 0 {
		if err := store(root, dir, e.Outputs, out.Bytes(), errb.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "monom: not cached: %v\n", err)
		}
	}
	return status, err
}

// store writes an entry into a temp dir and renames it into place, so a
// reader never sees half of one. stdout is written last: its presence marks
// a complete entry.
func store(root, dir string, outputs []string, out, errb []byte) error {
	base := filepath.Join(root, Dir)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(base, ".tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, o := range outputs {
		if _, err := os.Lstat(filepath.Join(root, o)); err != nil {
			return fmt.Errorf("declared output %s is missing", o)
		}
	}
	if err := copyPaths(root, filepath.Join(tmp, "outputs"), outputs); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, "stderr"), errb, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, "stdout"), out, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dir)
}

// copyPaths replaces each path under dst with its copy from src: regular
// files and directories only, permission bits kept.
func copyPaths(src, dst string, paths []string) error {
	for _, p := range paths {
		if err := os.RemoveAll(filepath.Join(dst, p)); err != nil {
			return err
		}
		err := filepath.WalkDir(filepath.Join(src, p), func(s string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(src, s)
			t := filepath.Join(dst, rel)
			info, err := d.Info()
			switch {
			case err != nil:
				return err
			case d.IsDir():
				return os.MkdirAll(t, info.Mode().Perm()|0o700)
			case !d.Type().IsRegular():
				return fmt.Errorf("%s is not a regular file or directory", s)
			}
			b, err := os.ReadFile(s)
			if err == nil {
				err = os.MkdirAll(filepath.Dir(t), 0o755)
			}
			if err == nil {
				err = os.WriteFile(t, b, info.Mode().Perm())
			}
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Run runs argv with stdin inherited, relaying SIGINT, SIGTERM, SIGHUP and
// SIGQUIT to it until it exits, and returns its status (128+n for a signal).
// A terminal Ctrl-C reaches the child directly too, so it may see SIGINT
// twice; relaying is what lets `kill` or a CI timeout aimed at the
// middleware reach the command.
func Run(argv []string, stdout, stderr io.Writer) (int, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigs)
	if err := cmd.Start(); err != nil {
		return 1, err
	}
	go func() {
		for s := range sigs {
			_ = cmd.Process.Signal(s)
		}
	}()
	err := cmd.Wait()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			return 128 + int(ws.Signal()), nil
		}
		return ee.ExitCode(), nil
	}
	return 0, err
}

// Exec replaces the process with argv, so the terminal, stdio and signals
// reach it untouched.
func Exec(argv []string) error {
	p, err := exec.LookPath(argv[0])
	if err == nil {
		err = syscall.Exec(p, argv, os.Environ())
	}
	return fmt.Errorf("exec %s: %w", argv[0], err)
}
