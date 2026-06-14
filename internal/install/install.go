package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Run executes `monomd install`: detects the user's shell, resolves the
// src/monom path relative to the running binary, and appends a source line
// to the appropriate rc/profile file if not already present.
func Run(executable string) error {
	srcMonom, err := resolveSrcMonom(executable)
	if err != nil {
		return fmt.Errorf("could not resolve src/monom path: %w", err)
	}

	rcFile, err := rcFileForShell(os.Getenv("SHELL"))
	if err != nil {
		return err
	}

	installed, err := alreadyInstalled(rcFile, srcMonom)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", rcFile, err)
	}
	if installed {
		fmt.Println("already installed")
		return nil
	}

	if err := appendSourceLine(rcFile, srcMonom); err != nil {
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

// rcFileForShell returns the rc/profile file path for the given $SHELL value.
func rcFileForShell(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch {
	case strings.HasSuffix(shell, "/zsh"):
		return filepath.Join(home, ".zshrc"), nil
	case strings.HasSuffix(shell, "/bash"):
		// On macOS, login shells source .bash_profile but not .bashrc.
		// Prefer .bash_profile so the integration loads for interactive use.
		profile := filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(profile); err == nil {
			return profile, nil
		}
		return filepath.Join(home, ".bashrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (only zsh and bash are supported)", shell)
	}
}

// alreadyInstalled reports whether rcFile already contains a line referencing srcMonom.
func alreadyInstalled(rcFile, srcMonom string) (bool, error) {
	data, err := os.ReadFile(rcFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.Contains(line, srcMonom) {
			return true, nil
		}
	}
	return false, nil
}

// appendSourceLine appends `source "<srcMonom>"` to rcFile, prepending a
// newline if the file does not already end with one.
func appendSourceLine(rcFile, srcMonom string) error {
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	prefix := needsLeadingNewline(rcFile)

	line := fmt.Sprintf(`source "%s"`, srcMonom)
	if prefix {
		_, err = fmt.Fprintf(f, "\n%s\n", line)
	} else {
		_, err = fmt.Fprintf(f, "%s\n", line)
	}
	return err
}

// needsLeadingNewline returns true when rcFile exists and its last byte is not '\n'.
func needsLeadingNewline(rcFile string) bool {
	f, err := os.Open(rcFile)
	if err != nil {
		return false
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.Size() == 0 {
		return false
	}

	buf := make([]byte, 1)
	if _, err := f.ReadAt(buf, info.Size()-1); err != nil {
		return false
	}
	return buf[0] != '\n'
}
