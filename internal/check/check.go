package check

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Check runs userConfig with the "complete" subcommand, reads all output paths,
// and returns a list of human-readable problem descriptions. An empty list means
// the project is healthy.
//
// Returns an error if userConfig is empty, the file does not exist, or it is
// not executable.
func Check(userConfig string) ([]string, error) {
	if userConfig == "" {
		return nil, fmt.Errorf("check: MONOM_USER_CONFIG is not set")
	}

	fi, err := os.Stat(userConfig)
	if err != nil {
		return nil, fmt.Errorf("check: cannot stat MONOM_USER_CONFIG (%s): %w", userConfig, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("check: MONOM_USER_CONFIG (%s) is a directory", userConfig)
	}
	if fi.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("check: MONOM_USER_CONFIG (%s) is not executable", userConfig)
	}

	cmd := exec.Command(userConfig, "complete")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("check: running %s complete: %w", userConfig, err)
	}

	var problems []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if hasSpaceInSegment(line) {
			problems = append(problems, fmt.Sprintf("path has space in segment: %q", line))
		}
	}

	return problems, nil
}

// hasSpaceInSegment reports whether any slash-delimited segment contains a space.
func hasSpaceInSegment(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if strings.ContainsRune(seg, ' ') {
			return true
		}
	}
	return false
}
