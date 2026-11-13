// Package runcache is the decision and storage half of monom's local,
// opt-in, input-based command cache; cmd/mnmd's `mnmd cache` runs it as an
// ordinary run middleware (internal/proc runs the child). A project opts a
// command in by naming its file in monom-cache.json at the project root,
// with explicit inputs, outputs and env vars:
//
//	{
//	  "build/compile.sh": {
//	    "inputs":  ["src", "go.mod"],
//	    "outputs": ["dist/app"],
//	    "env":     ["GOOS"]
//	  }
//	}
//
// Key hashes what a result depends on, Lookup and Restore serve a hit, and
// Store saves only a successful run whose declared outputs all exist,
// atomically.
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
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// FileName is the cache config at the project root.
const FileName = "monom-cache.json"

// Dir is where entries live, relative to the project root. Dot-prefixed, so
// discovery never scans it.
const Dir = ".monom/cache"

// keyVersion changes whenever the key's composition does, so old entries
// can never be hit by a new monom.
const keyVersion = "monom-cache v1"

// Entry declares how one command is cached. Every field is explicit: monom
// never infers inputs.
type Entry struct {
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
	Env     []string `json:"env"`
}

// Load reads monom-cache.json. A missing file is (nil, nil).
func Load(root string) (map[string]Entry, error) {
	b, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var f map[string]Entry
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	for cmd, e := range f {
		if err := validPath(cmd); err != nil {
			return nil, fmt.Errorf("%s: command %q: %w", FileName, cmd, err)
		}
		if len(e.Inputs) == 0 {
			return nil, fmt.Errorf("%s: %q: inputs must list at least one path", FileName, cmd)
		}
		for _, p := range append(append([]string{}, e.Inputs...), e.Outputs...) {
			if err := validPath(p); err != nil {
				return nil, fmt.Errorf("%s: %q: %q: %w", FileName, cmd, p, err)
			}
		}
		for _, n := range e.Env {
			if n == "" || strings.ContainsAny(n, "=\x00") {
				return nil, fmt.Errorf("%s: %q: invalid env name %q", FileName, cmd, n)
			}
		}
	}
	return f, nil
}

// validPath accepts a clean, slash-separated path inside the root.
func validPath(p string) error {
	if p == "" || path.IsAbs(p) || path.Clean(p) != p || p == "." || p == ".." || strings.HasPrefix(p, "../") || strings.HasPrefix(p, Dir) || p == ".monom" {
		return errors.New("must be a clean relative path inside the project, outside .monom")
	}
	return nil
}

// Key hashes everything a cached result depends on. A missing input is an
// error: the caller bypasses the cache rather than caching a guess.
func Key(root, cmd string, args []string, e Entry, getenv func(string) (string, bool)) (string, error) {
	h := sha256.New()
	put := func(parts ...string) { fmt.Fprintf(h, "%s\n", strings.Join(parts, "\x00")) }
	entry, _ := json.Marshal(e)
	put(keyVersion)
	put("command", cmd)
	put("entry", string(entry))
	for _, a := range args {
		put("arg", a)
	}
	files := map[string]bool{cmd: true}
	for _, in := range e.Inputs {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(in)), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("input %s is not a regular file", p)
			}
			rel, _ := filepath.Rel(root, p)
			files[filepath.ToSlash(rel)] = true
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("input %q: %w", in, err)
		}
	}
	names := make([]string, 0, len(files))
	for f := range files {
		names = append(names, f)
	}
	sort.Strings(names)
	for _, f := range names {
		sum, mode, err := fileSum(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil {
			return "", err
		}
		put("file", f, mode, sum)
	}
	env := append([]string{}, e.Env...)
	sort.Strings(env)
	for _, n := range env {
		// Only a hash of the value enters the key; values are never stored.
		if v, ok := getenv(n); ok {
			s := sha256.Sum256([]byte(v))
			put("env", n, "set", hex.EncodeToString(s[:]))
		} else {
			put("env", n, "unset")
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func fileSum(p string) (string, string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", "", err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", "", err
	}
	mode := "-"
	if info.Mode()&0o111 != 0 {
		mode = "x"
	}
	return hex.EncodeToString(h.Sum(nil)), mode, nil
}

// Result is a stored successful run.
type Result struct {
	Stdout, Stderr []byte
}

func entryDir(root, key string) string { return filepath.Join(root, filepath.FromSlash(Dir), key) }

// Lookup returns the stored result for key, or ok=false. A damaged entry is a
// miss.
func Lookup(root, key string) (Result, bool) {
	dir := entryDir(root, key)
	if _, err := os.Stat(filepath.Join(dir, "ok")); err != nil {
		return Result{}, false
	}
	out, err1 := os.ReadFile(filepath.Join(dir, "stdout"))
	errb, err2 := os.ReadFile(filepath.Join(dir, "stderr"))
	if err1 != nil || err2 != nil {
		return Result{}, false
	}
	return Result{Stdout: out, Stderr: errb}, true
}

// Restore puts the entry's outputs back at their declared paths, replacing
// what is there.
func Restore(root, key string, e Entry) error {
	for _, o := range e.Outputs {
		dst := filepath.Join(root, filepath.FromSlash(o))
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := copyTree(filepath.Join(entryDir(root, key), "outputs", filepath.FromSlash(o)), dst); err != nil {
			return fmt.Errorf("restore %s: %w", o, err)
		}
	}
	return nil
}

// Store saves a successful run. It writes into a temp dir and renames it into
// place, so a reader never sees a half-written entry. A missing declared
// output is an error and nothing is stored.
func Store(root, key string, e Entry, r Result) error {
	base := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(base, ".tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, o := range e.Outputs {
		src := filepath.Join(root, filepath.FromSlash(o))
		if _, err := os.Lstat(src); err != nil {
			return fmt.Errorf("declared output %s is missing", o)
		}
		if err := copyTree(src, filepath.Join(tmp, "outputs", filepath.FromSlash(o))); err != nil {
			return fmt.Errorf("output %s: %w", o, err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "stdout"), r.Stdout, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, "stderr"), r.Stderr, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, "ok"), nil, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, entryDir(root, key)); err != nil {
		if _, ok := Lookup(root, key); ok {
			return nil // a concurrent run stored the same key first
		}
		return err
	}
	return nil
}

// copyTree copies a regular file or a directory of regular files, keeping
// permission bits. Anything else (symlinks, devices) is refused.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		case d.Type().IsRegular():
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return os.WriteFile(target, b, info.Mode().Perm())
		default:
			return fmt.Errorf("%s is not a regular file or directory", p)
		}
	})
}
