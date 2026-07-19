// Package alias implements `mnmd alias-save`: it validates an alias name and a
// project root, then idempotently persists a define-only bind line to the
// user's rc file so the alias is re-created (with zero subprocesses) on each
// new shell. The in-shell definition itself is done by the `_monom_bind_alias`
// shell helper; this package only owns validation and persistence.
package alias

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/adamgen/monom/internal/rc"
	"github.com/adamgen/monom/internal/root"
)

// namePattern restricts alias names to a shell-identifier-safe token, so the
// generated `_monom_bind_alias <name> ...` line and the shell function it
// defines are safe to eval and cannot inject arbitrary shell.
var namePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// Save validates name and projectPath, then appends a define-only bind line to
// the user's rc file unless an identical line is already present.
func Save(name, projectPath string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid alias name %q: must match [A-Za-z_][A-Za-z0-9_-]*", name)
	}

	abs, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("could not resolve path %q: %w", projectPath, err)
	}
	if !root.IsValidProjectRoot(abs) {
		return fmt.Errorf("not a monom project root (no executable 'monom' file): %s", abs)
	}

	rcFile, err := rc.FileForShell(os.Getenv("SHELL"))
	if err != nil {
		return err
	}

	line := BindLine(name, abs)
	present, err := rc.HasLine(rcFile, line)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", rcFile, err)
	}
	if present {
		fmt.Println("already saved")
		return nil
	}
	if err := rc.AppendLine(rcFile, line); err != nil {
		return fmt.Errorf("could not write to %s: %w", rcFile, err)
	}

	fmt.Printf("aliased %s -> %s (added to %s)\n", name, abs, rcFile)
	return nil
}

// BindLine returns the rc line that binds the alias in a new shell. It is a
// direct call to the `_monom_bind_alias` shell function, so sourcing it spawns
// no subprocess.
func BindLine(name, absPath string) string {
	return fmt.Sprintf(`_monom_bind_alias %s "%s"`, name, absPath)
}
