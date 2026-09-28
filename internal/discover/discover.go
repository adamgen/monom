// Package discover implements monom's default discovery: turning the project
// root's file tree into the set of registered command paths without any help
// from the CLI author.
//
// Discovery runs in two stages. The scan is broad: every executable regular
// file under the root, minus noise (see skipDir). The gate is narrow: a
// scanned executable is registered only if it has a shebang or its name
// matches NamePattern. Explicitly declared paths (from a declarative monom
// config file) are registered regardless of the scan and the gate.
package discover

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/root"
)

// NamePattern is the naming gate for executables without a shebang (typically
// compiled binaries): an extensionless, lowercase command name that is also a
// clean completion token, like the demo project's `release` or `db/migrate`.
var NamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// noiseDirs are directory names whose contents are never scanned: dependency
// trees and build output that routinely contain executables nobody means as
// commands.
var noiseDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	"venv":         true,
	"target":       true,
	"dist":         true,
}

// Via names the gate rule that registered a command.
type Via string

const (
	ViaShebang  Via = "shebang"
	ViaPattern  Via = "name pattern"
	ViaDeclared Via = "declared"
)

// Command is a registered command path.
type Command struct {
	Path string // slash-delimited, relative to the project root
	Via  Via
}

// Skipped is a scanned executable that did not pass the gate.
type Skipped struct {
	Path   string
	Reason string
}

// Result is the outcome of Discover.
type Result struct {
	Commands   []Command // registered: gated scan results plus declarations
	Skipped    []Skipped // executables the scan found but the gate rejected
	Unreadable []string  // directories the scan could not read
	Invalid    []string  // declarations that do not name an executable file
}

// Paths returns the registered command paths.
func (r Result) Paths() []string {
	out := make([]string, len(r.Commands))
	for i, c := range r.Commands {
		out[i] = c.Path
	}
	return out
}

// Discover scans projectRoot and returns the registered command set, sorted by
// path. declared are command paths the author registered explicitly; they
// bypass both the noise rules and the gate, but must name an executable file.
func Discover(projectRoot string, declared []string) Result {
	var res Result
	seen := map[string]bool{}

	for _, d := range declared {
		if seen[d] {
			continue
		}
		if problem := validateDeclared(projectRoot, d); problem != "" {
			res.Invalid = append(res.Invalid, problem)
			continue
		}
		seen[d] = true
		res.Commands = append(res.Commands, Command{Path: d, Via: ViaDeclared})
	}

	_ = filepath.WalkDir(projectRoot, func(p string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(projectRoot, p)
		rel = filepath.ToSlash(rel)
		if err != nil {
			if d != nil && d.IsDir() {
				res.Unreadable = append(res.Unreadable, rel)
				return fs.SkipDir
			}
			return nil
		}
		if p == projectRoot {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if skipDir(p, name) {
				return fs.SkipDir
			}
			return nil
		}
		if isHiddenOrPrivate(name) || rel == root.ConfigFileName || seen[rel] {
			return nil
		}
		if !isExecutableFile(p) {
			return nil
		}

		switch {
		case strings.ContainsRune(rel, ' '):
			res.Skipped = append(res.Skipped, Skipped{Path: rel, Reason: "path contains a space, which cannot be typed as a command token"})
		case config.HasShebang(p):
			res.Commands = append(res.Commands, Command{Path: rel, Via: ViaShebang})
		case NamePattern.MatchString(name):
			res.Commands = append(res.Commands, Command{Path: rel, Via: ViaPattern})
		default:
			res.Skipped = append(res.Skipped, Skipped{Path: rel, Reason: "no shebang, and the name does not match " + NamePattern.String()})
		}
		return nil
	})

	sort.Slice(res.Commands, func(i, j int) bool { return res.Commands[i].Path < res.Commands[j].Path })
	return res
}

// skipDir reports whether a directory's contents are excluded from the scan:
// hidden and underscore-prefixed (private) directories, noise directories, and
// nested monom projects, which own their own command tree.
func skipDir(p, name string) bool {
	return isHiddenOrPrivate(name) || noiseDirs[name] || root.HasConfigFile(p)
}

// isHiddenOrPrivate matches dot-prefixed names and the underscore prefix monom
// uses for internals (e.g. `_monom_cfg`), which authors can use for helper
// scripts that must not become commands.
func isHiddenOrPrivate(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// isExecutableFile reports whether p (following symlinks) is a regular file
// with any execute bit set.
func isExecutableFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular() && fi.Mode()&0o111 != 0
}

// validateDeclared returns a problem description for a declaration that does
// not name an executable regular file inside the root, or "".
func validateDeclared(projectRoot, rel string) string {
	if strings.ContainsRune(rel, ' ') {
		return "declared command " + rel + ": path contains a space, which cannot be typed as a command token"
	}
	p := filepath.Join(projectRoot, filepath.FromSlash(rel))
	fi, err := os.Stat(p)
	switch {
	case err != nil:
		return "declared command " + rel + ": not found"
	case fi.IsDir():
		return "declared command " + rel + ": is a directory"
	case fi.Mode()&0o111 == 0:
		return "declared command " + rel + ": not executable"
	}
	return ""
}
