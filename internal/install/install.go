package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adamgen/monom/internal/rc"
)

// Run executes `mnmd install`: detects the user's shell, resolves the
// src/monom path relative to the running binary, and appends a source line
// to the appropriate rc/profile file if not already present.
func Run(executable string) error {
	srcMonom, err := resolveSrcMonom(executable)
	if err != nil {
		return fmt.Errorf("could not resolve src/monom path: %w", err)
	}

	rcFile, err := rc.FileForShell(os.Getenv("SHELL"))
	if err != nil {
		return err
	}

	installed, err := rc.HasLine(rcFile, srcMonom)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", rcFile, err)
	}
	if installed {
		fmt.Println("already installed")
		return nil
	}

	line := fmt.Sprintf(`source "%s"`, srcMonom)
	if err := rc.AppendLine(rcFile, line); err != nil {
		return fmt.Errorf("could not write to %s: %w", rcFile, err)
	}

	fmt.Printf("added to %s\n", rcFile)
	fmt.Println("restart your shell or run: source " + rcFile)
	return nil
}

// resolveSrcMonom returns the absolute path to src/monom relative to the
// real location of the running binary (symlinks resolved).
func resolveSrcMonom(executable string) (string, error) {
	real, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	binDir := filepath.Dir(real)
	return filepath.Join(binDir, "..", "src", "monom"), nil
}
