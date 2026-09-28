// Package config reads the monom config file: the file named "monom" at the
// project root.
//
// The file takes one of three shapes, decided by its first line and mode:
//
//   - a hook script: an executable file whose first line is a shebang. monom
//     runs it for hooks (complete, run, debug, config) and never parses it.
//   - a declarative file: any file whose first line is not a shebang,
//     including an empty file. monom parses it line by line and never runs it.
//   - absent: zero-config. Equivalent to an empty declarative file.
//
// A declarative file, and the output of a hook script's `config` hook, share
// one line format:
//
//	# comment
//	key = value        a setting
//	tools/build        a declared command path, relative to the project root
//
// Most settings take one value, and a later line replaces an earlier one.
// discover.hide is repeatable: every line adds one pattern.
package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

// Kind classifies a monom config file.
type Kind int

const (
	// Absent means there is no config file: zero-config.
	Absent Kind = iota
	// Declarative is a file whose first line is not a shebang (or is empty).
	Declarative
	// Script is an executable file whose first line is a shebang.
	Script
	// InertScript has a shebang but no execute bit, so its hooks cannot run.
	InertScript
)

// Severity values accepted by severity-valued settings.
const (
	SeverityWarning = "warning"
	SeverityError   = "error"
)

// KeyRootContents selects the severity of the check that validates the
// project root's contents (see internal/check).
const KeyRootContents = "check.root-contents"

// KeyDiscoverHide adds a pattern whose matches default discovery treats as
// hidden, like dot-files: a matching file is never registered and a matching
// directory is never scanned (see internal/discover). Repeatable. A pattern
// without a slash matches an entry's name at any depth; one with a slash
// matches its path relative to the project root. Both use path.Match syntax.
const KeyDiscoverHide = "discover.hide"

// Settings holds the recognised config keys. A zero value field means unset.
type Settings struct {
	RootContents string   // SeverityWarning, SeverityError, or "" when unset
	Hide         []string // discover.hide patterns, normalized, in order
}

// File is a loaded monom config file.
type File struct {
	Path     string
	Kind     Kind
	Declared []string // declared command paths, slash-delimited, as written
	Settings Settings
	Problems []string // invalid lines, unknown keys, invalid values
}

// Load classifies and, when declarative, parses the config file at path. A
// missing file is not an error: it yields Kind Absent. Only a script's shape
// is inspected; its contents are never parsed.
func Load(path string) (*File, error) {
	f := &File{Path: path}
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot stat monom config file %s: %w", path, err)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("monom config file %s is not a regular file", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read monom config file %s: %w", path, err)
	}

	if bytes.HasPrefix(content, []byte("#!")) {
		f.Kind = InertScript
		if fi.Mode()&0o111 != 0 {
			f.Kind = Script
		}
		return f, nil
	}

	f.Kind = Declarative
	f.Declared, f.Settings, f.Problems = Parse(bytes.NewReader(content), true)
	return f, nil
}

// Parse reads the shared line format. When allowDeclarations is false (the
// output of a `config` hook, which can only carry settings), a command path
// line is reported as a problem instead of being declared.
func Parse(r io.Reader, allowDeclarations bool) (declared []string, settings Settings, problems []string) {
	scanner := bufio.NewScanner(r)
	n := 0
	for scanner.Scan() {
		n++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if key, value, ok := strings.Cut(line, "="); ok {
			if p := settings.set(strings.TrimSpace(key), strings.TrimSpace(value)); p != "" {
				problems = append(problems, fmt.Sprintf("line %d: %s", n, p))
			}
			continue
		}

		if !allowDeclarations {
			problems = append(problems, fmt.Sprintf("line %d: expected 'key = value', got %q", n, line))
			continue
		}
		decl, ok := normalizeDeclaration(line)
		if !ok {
			problems = append(problems, fmt.Sprintf("line %d: invalid declaration %q: must be a path inside the project root", n, line))
			continue
		}
		declared = append(declared, decl)
	}
	return declared, settings, problems
}

// Merge returns s with every single-valued field that is set in over
// replacing its value. Repeatable fields (Hide) are appended.
func (s Settings) Merge(over Settings) Settings {
	if over.RootContents != "" {
		s.RootContents = over.RootContents
	}
	if len(over.Hide) > 0 {
		s.Hide = append(s.Hide[:len(s.Hide):len(s.Hide)], over.Hide...)
	}
	return s
}

// set applies one key/value pair. It returns a problem description, or "".
func (s *Settings) set(key, value string) string {
	switch key {
	case KeyRootContents:
		v, ok := ParseSeverity(value)
		if !ok {
			return fmt.Sprintf("invalid value %q for %s: want %q or %q", value, key, SeverityWarning, SeverityError)
		}
		s.RootContents = v
		return ""
	case KeyDiscoverHide:
		p, problem := NormalizeHidePattern(value)
		if problem != "" {
			return fmt.Sprintf("invalid value %q for %s: %s", value, key, problem)
		}
		s.Hide = append(s.Hide, p)
		return ""
	default:
		return fmt.Sprintf("unknown setting %q", key)
	}
}

// ParseSeverity validates a severity value.
func ParseSeverity(v string) (string, bool) {
	switch v {
	case SeverityWarning, SeverityError:
		return v, true
	}
	return "", false
}

// NormalizeHidePattern validates a discover.hide pattern and trims a leading
// "./" and a trailing "/". It returns the pattern, or a problem description.
func NormalizeHidePattern(v string) (string, string) {
	p := strings.TrimSuffix(strings.TrimPrefix(v, "./"), "/")
	switch {
	case p == "":
		return "", "want a file name or a path pattern"
	case strings.HasPrefix(p, "/") || dotSegment.MatchString(p):
		return "", "must be a name or a path inside the project root"
	}
	if _, err := path.Match(p, ""); err != nil {
		return "", "invalid pattern (see path.Match)"
	}
	return p, ""
}

var dotSegment = regexp.MustCompile(`(^|/)\.\.?(/|$)`)

// normalizeDeclaration trims "./" and trailing "/" and rejects paths that are
// absolute or climb out of the root.
func normalizeDeclaration(line string) (string, bool) {
	p := strings.TrimSuffix(strings.TrimPrefix(line, "./"), "/")
	if p == "" || strings.HasPrefix(p, "/") || dotSegment.MatchString(p) {
		return "", false
	}
	return path.Clean(p), true
}

// HasShebang reports whether the file at p starts with "#!". Unreadable files
// report false.
func HasShebang(p string) bool {
	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 2)
	n, _ := io.ReadFull(f, buf)
	return n == 2 && buf[0] == '#' && buf[1] == '!'
}

// RunHook runs the hook script with one subcommand and returns its stdout. A
// non-zero exit is an error that carries the script's stderr, if any.
func RunHook(script, hook string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(script, hook)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// ProjectSettings returns the project's own settings: a declarative file's,
// or what a hook script's `config` hook prints. It is the lenient reader for
// callers that must never fail, like discovery on Tab: an invalid line, or a
// config hook that fails, contributes nothing. `mnmd check` reports those.
func (f *File) ProjectSettings() Settings {
	switch f.Kind {
	case Declarative:
		return f.Settings
	case Script:
		out, err := RunHook(f.Path, "config")
		if err != nil {
			return Settings{}
		}
		_, settings, _ := Parse(bytes.NewReader(out), false)
		return settings
	}
	return Settings{}
}
