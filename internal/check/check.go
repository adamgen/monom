package check

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/mapfile"
)

// Check runs userConfig with the "complete" subcommand, reads all output paths,
// and returns a list of human-readable problem descriptions. When a command map
// (monom-map.json) exists next to the user config, it is validated too. An
// empty list means the project is healthy.
//
// Returns an error if userConfig is empty, the file does not exist, or it is
// not executable.
func Check(userConfig string) ([]string, error) {
	if userConfig == "" {
		return nil, fmt.Errorf("check: _MONOM_USER_CONFIG is not set")
	}

	fi, err := os.Stat(userConfig)
	if err != nil {
		return nil, fmt.Errorf("check: cannot stat _MONOM_USER_CONFIG (%s): %w", userConfig, err)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("check: _MONOM_USER_CONFIG (%s) is a directory", userConfig)
	}
	if fi.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("check: _MONOM_USER_CONFIG (%s) is not executable", userConfig)
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

	problems = append(problems, checkMap(filepath.Dir(userConfig))...)

	return problems, nil
}

// checkMap validates the command map when one exists at the project root
// (the directory containing the user config). Absent map = no problems: the
// map is optional. Present map = every target must exist, be a regular file,
// and be executable, and no map entry may shadow a tree path that pack could
// otherwise resolve.
func checkMap(rootDir string) []string {
	path := filepath.Join(rootDir, mapfile.FileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []string{fmt.Sprintf("%s: cannot open: %v", mapfile.FileName, err)}
	}
	defer f.Close()

	m, err := mapfile.Parse(f)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", mapfile.FileName, err)}
	}

	var problems []string
	mapfile.Walk(m, func(cmdPath string, n *mapfile.Node) {
		treePath := filepath.Join(rootDir, cmdPath)
		treeInfo, treeErr := os.Stat(treePath)

		if n.IsCommand() {
			// The map wins over the tree (the run hook fires before pack), so
			// any tree entry at the same path is silently unreachable.
			if treeErr == nil {
				problems = append(problems, fmt.Sprintf("%s: command %q shadows tree path %s", mapfile.FileName, cmdPath, treePath))
			}
			problems = append(problems, checkTarget(rootDir, cmdPath, n.Run)...)
			return
		}

		// A map category over a tree DIRECTORY merges (unmapped children fall
		// back to pack) — only a category over a tree FILE shadows a command.
		if treeErr == nil && !treeInfo.IsDir() {
			problems = append(problems, fmt.Sprintf("%s: category %q shadows tree command %s", mapfile.FileName, cmdPath, treePath))
		}
	})
	return problems
}

// checkTarget validates one command's run target on disk.
func checkTarget(rootDir, cmdPath, target string) []string {
	abs := filepath.Join(rootDir, target)
	fi, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("%s: target of %q does not exist: %s", mapfile.FileName, cmdPath, abs)}
		}
		return []string{fmt.Sprintf("%s: cannot stat target of %q: %v", mapfile.FileName, cmdPath, err)}
	}
	if fi.IsDir() {
		return []string{fmt.Sprintf("%s: target of %q is a directory: %s", mapfile.FileName, cmdPath, abs)}
	}
	if fi.Mode()&0o111 == 0 {
		return []string{fmt.Sprintf("%s: target of %q is not executable: %s", mapfile.FileName, cmdPath, abs)}
	}
	return nil
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
