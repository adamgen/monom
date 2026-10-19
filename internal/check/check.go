// Package check implements `mnmd check`, monom's doctor: it validates the
// project's registered command set and reports problems with a severity.
//
// Every check reports errors, except root-contents, whose severity comes from
// config: warning by default, so a fresh zero-config project never fails.
package check

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/discover"
	"github.com/adamgen/monom/internal/mapfile"
)

// Check names, printed with every problem.
const (
	// PathSpaces: a registered path has a space in a segment, so `mnmd filter`
	// silently drops it and the command can never be completed.
	PathSpaces = "path-spaces"
	// Declared: a declaration in a declarative config names no executable file.
	Declared = "declared-commands"
	// Config: the config file, the config hook, or a config env var is invalid.
	Config = "config"
	// RootContents: the project root's contents cannot all be used as they
	// are — executables default discovery found but did not register,
	// directories it could not read, or a hook script missing its execute bit.
	// The only check whose severity is configurable.
	RootContents = "root-contents"
	// CommandNames: suffix stripping could not give a script its command
	// name — two scripts share a stem, or the stem names an existing entry.
	CommandNames = "command-names"
	// CommandMap: the command map (monom-map.json) does not parse, a run
	// target is unusable, or a map entry shadows a tree path.
	CommandMap = "command-map"
)

// UserSeverityEnv is the user-level (global) setting for the root-contents
// severity. A project's `check.root-contents` setting overrides it.
const UserSeverityEnv = "MONOM_CHECK_ROOT_CONTENTS"

// Problem is one finding.
type Problem struct {
	Check    string
	Severity string // config.SeverityWarning or config.SeverityError
	Message  string
}

func (p Problem) String() string {
	return fmt.Sprintf("%s: [%s] %s", p.Severity, p.Check, p.Message)
}

// Input names what to check.
type Input struct {
	Root         string // project root; needed only when default discovery runs
	UserConfig   string // path of the monom config file; it need not exist
	UserSeverity string // raw value of UserSeverityEnv
}

// Report is the result of a check run.
type Report struct {
	Commands   []string  // the registered command set that was validated
	Discovered bool      // Commands came from default discovery, not the complete hook
	Problems   []Problem // in check order
}

// Count returns the number of problems with the given severity.
func (r Report) Count(severity string) int {
	n := 0
	for _, p := range r.Problems {
		if p.Severity == severity {
			n++
		}
	}
	return n
}

// Check validates the project. It returns an error only when the project
// cannot be inspected at all (no config path, an unreadable config, a failing
// complete hook, or default discovery with no root); everything else is a
// Problem in the report.
func Check(in Input) (Report, error) {
	var rep Report
	if in.UserConfig == "" {
		return rep, fmt.Errorf("check: no monom config file path")
	}
	cfg, err := config.Load(in.UserConfig)
	if err != nil {
		return rep, fmt.Errorf("check: %w", err)
	}

	errorf := func(check, format string, a ...any) {
		rep.Problems = append(rep.Problems, Problem{check, config.SeverityError, fmt.Sprintf(format, a...)})
	}
	var rootContents []string

	settings := config.Settings{}
	if in.UserSeverity != "" {
		if v, ok := config.ParseSeverity(in.UserSeverity); ok {
			settings.RootContents = v
		} else {
			errorf(Config, "%s=%q: want %q or %q", UserSeverityEnv, in.UserSeverity, config.SeverityWarning, config.SeverityError)
		}
	}

	var completeOut []byte
	switch cfg.Kind {
	case config.Declarative:
		for _, p := range cfg.Problems {
			errorf(Config, "%s: %s", in.UserConfig, p)
		}
		settings = settings.Merge(cfg.Settings)
	case config.Script:
		out, err := config.RunHook(in.UserConfig, "config")
		if err != nil {
			errorf(Config, "config hook failed: %v", err)
		} else {
			_, hookSettings, problems := config.Parse(bytes.NewReader(out), false)
			for _, p := range problems {
				errorf(Config, "config hook output %s", p)
			}
			settings = settings.Merge(hookSettings)
		}
		if completeOut, err = config.RunHook(in.UserConfig, "complete"); err != nil {
			return rep, fmt.Errorf("check: running %s complete: %w", in.UserConfig, err)
		}
	case config.InertScript:
		rootContents = append(rootContents, fmt.Sprintf("%s has a shebang but is not executable, so its hooks never run (chmod +x %s)", in.UserConfig, in.UserConfig))
	}

	rep.Commands = nonEmptyLines(completeOut)
	if len(rep.Commands) == 0 {
		if in.Root == "" {
			return rep, fmt.Errorf("check: default discovery needs a project root, and none was found")
		}
		res := discover.Discover(in.Root, cfg.Declared, settings)
		rep.Commands = res.Paths()
		rep.Discovered = true
		for _, msg := range res.Invalid {
			errorf(Declared, "%s", msg)
		}
		for _, msg := range res.Conflicts {
			errorf(CommandNames, "%s", msg)
		}
		for _, s := range res.Skipped {
			rootContents = append(rootContents, fmt.Sprintf("executable not registered: %s (%s); add a shebang, rename it, or declare it in the monom config file", s.Path, s.Reason))
		}
		for _, d := range res.Unreadable {
			rootContents = append(rootContents, fmt.Sprintf("directory could not be read, so its commands are not registered: %s", d))
		}
	}

	for _, c := range rep.Commands {
		if hasSpaceInSegment(c) {
			errorf(PathSpaces, "path has space in segment: %q", c)
		}
	}

	if in.Root != "" {
		for _, msg := range checkMap(in.Root) {
			errorf(CommandMap, "%s", msg)
		}
	}

	severity := settings.RootContents
	if severity == "" {
		severity = config.SeverityWarning
	}
	for _, msg := range rootContents {
		rep.Problems = append(rep.Problems, Problem{RootContents, severity, msg})
	}
	return rep, nil
}

func nonEmptyLines(b []byte) []string {
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// checkMap validates the command map when one exists at the project root.
// Absent map = no problems: the
// map is optional. Present map = every target must exist, be a regular file,
// and be executable, and no map entry may shadow a tree path that pack could
// otherwise resolve.
func checkMap(rootDir string) []string {
	f, err := os.Open(filepath.Join(rootDir, mapfile.FileName))
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
				problems = append(problems, fmt.Sprintf("%s: command %q shadows tree path %s", mapfile.FileName, cmdPath, cmdPath))
			} else if dir, stem := path.Split(cmdPath); os.IsNotExist(treeErr) {
				if m := discover.StemMatches(filepath.Join(rootDir, dir), stem); len(m) == 1 {
					problems = append(problems, fmt.Sprintf("%s: command %q shadows tree path %s", mapfile.FileName, cmdPath, dir+m[0]))
				}
			}
			problems = append(problems, checkTarget(rootDir, cmdPath, n.Run)...)
			return
		}

		// A map category over a tree DIRECTORY merges (unmapped children fall
		// back to pack) — only a category over a tree FILE shadows a command.
		if treeErr == nil && !treeInfo.IsDir() {
			problems = append(problems, fmt.Sprintf("%s: category %q shadows tree command %s", mapfile.FileName, cmdPath, cmdPath))
		}
	})
	return problems
}

// checkTarget validates one command's run target on disk. Messages name the
// target as the map spells it, relative to the root.
func checkTarget(rootDir, cmdPath, target string) []string {
	abs := filepath.Join(rootDir, target)
	fi, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("%s: target of %q does not exist: %s", mapfile.FileName, cmdPath, target)}
		}
		return []string{fmt.Sprintf("%s: cannot stat target of %q: %v", mapfile.FileName, cmdPath, err)}
	}
	if fi.IsDir() {
		return []string{fmt.Sprintf("%s: target of %q is a directory: %s", mapfile.FileName, cmdPath, target)}
	}
	if fi.Mode()&0o111 == 0 {
		return []string{fmt.Sprintf("%s: target of %q is not executable: %s", mapfile.FileName, cmdPath, target)}
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
