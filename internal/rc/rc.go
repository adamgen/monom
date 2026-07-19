// Package rc centralizes reading and writing the user's shell rc/profile file:
// selecting the file by shell, checking whether a line is already present, and
// appending a line with correct newline handling. It is the single owner of
// rc-file logic, shared by `mnmd install` and `mnmd alias-save`.
package rc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileForShell returns the rc/profile file path for the given $SHELL value.
func FileForShell(shell string) (string, error) {
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

// HasLine reports whether rcFile contains a non-comment line that includes needle.
// A missing file is treated as "not present" rather than an error.
func HasLine(rcFile, needle string) (bool, error) {
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
		if strings.Contains(line, needle) {
			return true, nil
		}
	}
	return false, nil
}

// AppendLine appends line to rcFile (creating it if absent), prepending a
// newline when the file does not already end with one so the line stands alone.
func AppendLine(rcFile, line string) error {
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if needsLeadingNewline(rcFile) {
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
