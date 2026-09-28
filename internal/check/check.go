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
	"strings"

	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/discover"
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
		res := discover.Discover(in.Root, cfg.Declared, settings.Hide)
		rep.Commands = res.Paths()
		rep.Discovered = true
		for _, msg := range res.Invalid {
			errorf(Declared, "%s", msg)
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

// hasSpaceInSegment reports whether any slash-delimited segment contains a space.
func hasSpaceInSegment(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if strings.ContainsRune(seg, ' ') {
			return true
		}
	}
	return false
}
