// Package runcache is monom's local, opt-in command cache, run as the
// `mnmd cache` middleware. A command is cached when the monom config sets
//
//	cache.inputs.<command file> = <shell command>
//
// The inputs command runs from the project root; its stdout, the command
// file, the rest of the chain and the cache settings form the key. A hit
// restores the stored files and replays stdout then stderr without calling
// next; a miss runs next and stores an exit-0 run. Stored files are the
// cache.outputs paths when set, otherwise the files git sees the run create,
// change or delete (ignored files included). Without git: stdout/stderr only.
package runcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/adamgen/monom/internal/config"
)

// Dir holds the entries, relative to the project root.
const Dir = ".monom/cache"

// Middleware is `mnmd cache <next...>` for a command with a cache rule. It
// returns the exit status to use; a pass-through execs next and never returns.
func Middleware(root, cmd string, rule config.CacheRule, next []string) (int, error) {
	if os.Getenv("MONOM_CACHE") == "off" {
		return 1, Exec(next)
	}
	key, err := Key(root, cmd, rule, next)
	if err != nil {
		fmt.Fprintf(os.Stderr, "monom: cache bypassed for %s: %v\n", cmd, err)
		return 1, Exec(next)
	}
	dir := filepath.Join(root, Dir, key)
	if out, err := os.ReadFile(filepath.Join(dir, "stdout")); err == nil {
		errb, _ := os.ReadFile(filepath.Join(dir, "stderr"))
		if err := restore(root, dir); err == nil {
			fmt.Fprintf(os.Stderr, "monom: cache hit for %s (%s)\n", cmd, key[:12])
			os.Stdout.Write(out)
			os.Stderr.Write(errb)
			return 0, nil
		}
		os.RemoveAll(dir) // damaged: a miss that replaces it
	}
	before, git := snapshot(root, nil)
	if len(rule.Outputs) > 0 {
		git = false
	}
	var out, errb bytes.Buffer
	status, err := Run(next, io.MultiWriter(os.Stdout, &out), io.MultiWriter(os.Stderr, &errb))
	if err != nil || status != 0 {
		return status, err
	}
	files, replace, deleted := rule.Outputs, rule.Outputs, []string(nil)
	if git {
		files, deleted = changes(root, before)
	}
	if err := store(root, dir, files, replace, deleted, out.Bytes(), errb.Bytes()); err != nil {
		fmt.Fprintf(os.Stderr, "monom: not cached: %v\n", err)
	}
	return 0, nil
}

// Key hashes the inputs command's stdout with the command file, next's argv
// and the rule. A failing inputs command is an error: the caller runs the
// command uncached.
func Key(root, cmd string, rule config.CacheRule, next []string) (string, error) {
	c := exec.Command("sh", "-c", rule.Inputs)
	c.Dir = root
	var stderr bytes.Buffer
	c.Stderr = &stderr
	inputs, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("inputs command failed (%v) %s", err, strings.TrimSpace(stderr.String()))
	}
	body, err := os.ReadFile(filepath.Join(root, cmd))
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, part := range append([]string{"monom-cache v2", cmd, string(body), rule.Inputs, strings.Join(rule.Outputs, " "), string(inputs)}, next...) {
		fmt.Fprintf(h, "%d:%s", len(part), part)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// snapshot fingerprints (size, mtime, mode) every path `git status` lists
// under root (modified, untracked and ignored files, one by one), plus the
// extra paths, skipping .monom/. ok is false outside a git work tree.
func snapshot(root string, extra []string) (map[string]string, bool) {
	top, err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, false
	}
	out, err := exec.Command("git", "-C", root, "status", "--porcelain", "-z", "--untracked-files=all", "--ignored", "--", ".").Output()
	if err != nil {
		return nil, false
	}
	fps := map[string]string{}
	add := func(rel string) {
		if rel != ".monom" && !strings.HasPrefix(rel, ".monom/") {
			fps[rel] = fingerprint(filepath.Join(root, rel))
		}
	}
	recs := strings.Split(string(out), "\x00")
	for i := 0; i < len(recs); i++ {
		if len(recs[i]) < 4 {
			continue
		}
		abs := filepath.Join(strings.TrimSpace(string(top)), recs[i][3:])
		if rel, err := filepath.Rel(root, abs); err == nil && !strings.HasPrefix(rel, "..") {
			add(filepath.ToSlash(rel))
		}
		if recs[i][0] == 'R' || recs[i][0] == 'C' {
			i++ // the rename's source path
		}
	}
	for _, p := range extra {
		add(p)
	}
	return fps, true
}

func fingerprint(p string) string {
	fi, err := os.Lstat(p)
	if err != nil {
		return "missing"
	}
	return fmt.Sprint(fi.Size(), fi.ModTime().UnixNano(), fi.Mode())
}

// changes compares a fresh snapshot with before: paths whose fingerprint
// changed and still exist are files to store; the rest were deleted.
func changes(root string, before map[string]string) (files, deleted []string) {
	var known []string
	for p := range before {
		known = append(known, p)
	}
	after, _ := snapshot(root, known)
	for p, fp := range after {
		switch {
		case fp == before[p]:
		case fp == "missing":
			deleted = append(deleted, p)
		default:
			files = append(files, p)
		}
	}
	return files, deleted
}

// store writes an entry into a temp dir and renames it into place, so a
// reader never sees half of one; stdout, written last, marks it complete.
// replace lists paths restore removes before copying files back; deleted
// lists paths restore removes.
func store(root, dir string, files, replace, deleted []string, out, errb []byte) error {
	base := filepath.Join(root, Dir)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(base, ".tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, f := range files {
		if _, err := os.Lstat(filepath.Join(root, f)); err != nil {
			return fmt.Errorf("declared output %s is missing", f)
		}
		if err := copyTree(filepath.Join(root, f), filepath.Join(tmp, "files", f)); err != nil {
			return err
		}
	}
	for name, b := range map[string][]byte{
		"replace": []byte(strings.Join(replace, "\x00")),
		"deleted": []byte(strings.Join(deleted, "\x00")),
		"stderr":  errb,
	} {
		if err := os.WriteFile(filepath.Join(tmp, name), b, 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "stdout"), out, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dir)
}

// restore replays an entry's file effects: remove the replaced paths, copy
// the stored files back, remove the deleted paths.
func restore(root, dir string) error {
	list := func(name string) []string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return strings.FieldsFunc(string(b), func(r rune) bool { return r == 0 })
	}
	for _, p := range list("replace") {
		if err := os.RemoveAll(filepath.Join(root, p)); err != nil {
			return err
		}
	}
	if err := copyTree(filepath.Join(dir, "files"), root); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, p := range list("deleted") {
		if err := os.RemoveAll(filepath.Join(root, p)); err != nil {
			return err
		}
	}
	return nil
}

// copyTree copies a regular file or a directory of them, keeping permission
// bits; anything else is refused.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(s string, d fs.DirEntry, err error) error {
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
