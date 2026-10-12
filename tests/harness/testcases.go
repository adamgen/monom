// Package harness loads and strictly validates the declarative CLI test
// cases in tests/cases/*.yaml. It is test tooling: nothing in mnmd imports it.
// The runner that executes the cases is TestCases in cases_test.go. This is a
// separate Go module (tests/harness) so its test-only dependencies stay out
// of the root go.mod.
//
// Schema (see tests/README.md for the full guide):
//
//	root: fixtures/demo-project     # required unless every group sets it
//	shells: [bash, zsh]             # optional; default [bash, zsh]
//	env: {NAME: value}              # optional; also on groups and cases, merged
//	cases:                          # optional
//	  - name: <unique in the file>  # required
//	    input: <CLI line>           # required
//	    action: enter | tab | keys  # required
//	    exit: <int>                 # optional, enter only, default 0
//	    match: exact | normalized   # optional, enter and tab, default exact
//	    expect: <string> | [list]   # enter and tab: a string for enter, a list for tab
//	    tabs: 1 | 2 | 3             # keys only, required: Tab presses after the input
//	    line: <string>              # keys only, required: the edit buffer afterwards
//	    candidates: [list]          # keys only, optional: the listing on screen, as a set
//	    env: {NAME: value}          # optional, merged over the group's
//	groups:                         # optional
//	  - name: <group name>          # required
//	    root: <dir>                 # optional override
//	    shells: [bash]              # optional override
//	    env: {NAME: value}          # optional, merged over the suite's
//	    cases: [...]                # required
package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Case is one validated case, with its suite and group config resolved.
type Case struct {
	Name   string
	File   string // as given to Load, for error and failure context
	Line   int    // line where the case starts
	Group  string // "" for suite-level cases
	Root   string // relative to the repo root
	Shells []string
	// Env is set in the case's shell on top of the scrubbed environment.
	// "{root}" in a value stands for the absolute path of Root.
	Env    map[string]string
	Input  string
	Action string // ActionEnter or ActionTab
	Expect string // enter and tab; tab candidates are newline-joined
	Exit   int
	Match  string // MatchExact or MatchNormalized

	// Action keys only: an interactive shell in a pseudo-terminal.
	Tabs            int      // Tab presses after typing Input
	WantLine        string   // the line editor's buffer afterwards
	Candidates      []string // the completion listing on screen, compared as a set
	CheckCandidates bool     // whether `candidates` was given ([] means no listing)
}

const (
	ActionEnter     = "enter"
	ActionTab       = "tab"
	ActionKeys      = "keys"
	maxTabs         = 3
	MatchExact      = "exact"
	MatchNormalized = "normalized"
)

var (
	fileKeys  = []string{"root", "shells", "env", "cases", "groups"}
	groupKeys = []string{"name", "root", "shells", "env", "cases"}
	caseKeys  = []string{"name", "input", "action", "env", "exit", "match", "expect", "tabs", "line", "candidates"}

	envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// Errors collects every validation problem in a file.
type Errors []string

func (e Errors) Error() string { return strings.Join(e, "\n") }

type lineError struct {
	line int
	msg  string
}

type config struct {
	root   string
	shells []string
	env    map[string]string
}

type loader struct {
	file     string
	repoRoot string
	errs     []lineError
	cases    []Case
	names    map[string]int
}

// Load parses and validates one case file. A relative file is resolved
// against repoRoot but reported as given, and every root must be a directory
// under repoRoot. On any problem it returns all of them as Errors, each
// prefixed with file:line and, where known, the group and case.
func Load(file, repoRoot string) ([]Case, error) {
	path := file
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	l := &loader{file: file, repoRoot: repoRoot, names: map[string]int{}}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, Errors{fmt.Sprintf("%s: invalid YAML: %v", file, err)}
	}
	if len(doc.Content) == 0 {
		return nil, Errors{fmt.Sprintf("%s: empty file", file)}
	}
	l.suite(doc.Content[0])
	if len(l.errs) > 0 {
		sort.SliceStable(l.errs, func(i, j int) bool { return l.errs[i].line < l.errs[j].line })
		out := make(Errors, len(l.errs))
		for i, e := range l.errs {
			out[i] = e.msg
		}
		return nil, out
	}
	return l.cases, nil
}

func (l *loader) errorf(n *yaml.Node, ctx, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	if ctx != "" {
		msg = ctx + ": " + msg
	}
	l.errs = append(l.errs, lineError{n.Line, fmt.Sprintf("%s:%d: %s", l.file, n.Line, msg)})
}

// fields returns the mapping's key→value nodes, reporting unknown and
// duplicate keys.
func (l *loader) fields(n *yaml.Node, ctx string, allowed []string) map[string]*yaml.Node {
	out := map[string]*yaml.Node{}
	if n.Kind != yaml.MappingNode {
		l.errorf(n, ctx, "expected a mapping")
		return out
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		switch {
		case !contains(allowed, k.Value):
			l.errorf(k, ctx, "unknown key %q (want one of: %s)", k.Value, strings.Join(allowed, ", "))
		case out[k.Value] != nil:
			l.errorf(k, ctx, "duplicate key %q", k.Value)
		default:
			out[k.Value] = v
		}
	}
	return out
}

func (l *loader) suite(n *yaml.Node) {
	f := l.fields(n, "", fileKeys)
	cfg := l.config(f, "", config{shells: []string{"bash", "zsh"}})

	if f["cases"] == nil && f["groups"] == nil {
		l.errorf(n, "", "no cases: add `cases:` or `groups:`")
	}
	if c := f["cases"]; c != nil {
		l.caseList(c, "", "", cfg)
	}
	if g := f["groups"]; g != nil {
		if g.Kind != yaml.SequenceNode {
			l.errorf(g, "", "groups must be a list")
			return
		}
		for _, gn := range g.Content {
			l.group(gn, cfg)
		}
	}
}

func (l *loader) group(n *yaml.Node, suite config) {
	ctx := label("group", peekName(n))
	f := l.fields(n, ctx, groupKeys)
	name := l.str(f["name"], ctx, "name", n)
	cfg := l.config(f, ctx, suite)
	if f["cases"] == nil {
		l.errorf(n, ctx, "missing required key \"cases\"")
		return
	}
	l.caseList(f["cases"], name, ctx, cfg)
}

// config applies root/shells overrides on top of base.
func (l *loader) config(f map[string]*yaml.Node, ctx string, base config) config {
	cfg := base
	if r := f["root"]; r != nil {
		cfg.root = l.str(r, ctx, "root", r)
		if cfg.root != "" {
			if fi, err := os.Stat(filepath.Join(l.repoRoot, cfg.root)); err != nil || !fi.IsDir() {
				l.errorf(r, ctx, "root %q is not a directory (relative to the repo root)", cfg.root)
			}
		}
	}
	if s := f["shells"]; s != nil {
		cfg.shells = nil
		if s.Kind != yaml.SequenceNode || len(s.Content) == 0 {
			l.errorf(s, ctx, "shells must be a non-empty list of bash and/or zsh")
		} else {
			for _, sh := range s.Content {
				if sh.Value != "bash" && sh.Value != "zsh" {
					l.errorf(sh, ctx, "unknown shell %q (want bash or zsh)", sh.Value)
					continue
				}
				cfg.shells = append(cfg.shells, sh.Value)
			}
		}
	}
	if e := f["env"]; e != nil {
		cfg.env = l.env(e, ctx, base.env)
	}
	return cfg
}

// env returns base with the NAME: value pairs of n laid over it.
func (l *loader) env(n *yaml.Node, ctx string, base map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	if n.Kind != yaml.MappingNode {
		l.errorf(n, ctx, "env must be a mapping of NAME: value")
		return out
	}
	seen := map[string]bool{}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		switch {
		case !envName.MatchString(k.Value):
			l.errorf(k, ctx, "env name %q is not a valid variable name", k.Value)
		case seen[k.Value]:
			l.errorf(k, ctx, "duplicate env name %q", k.Value)
		case v.Kind != yaml.ScalarNode:
			l.errorf(v, ctx, "env %s must be a string", k.Value)
		default:
			out[k.Value] = v.Value
		}
		seen[k.Value] = true
	}
	return out
}

func (l *loader) caseList(n *yaml.Node, group, ctx string, cfg config) {
	if n.Kind != yaml.SequenceNode {
		l.errorf(n, ctx, "cases must be a list")
		return
	}
	for _, cn := range n.Content {
		l.oneCase(cn, group, ctx, cfg)
	}
}

func (l *loader) oneCase(n *yaml.Node, group, groupCtx string, cfg config) {
	ctx := label("case", peekName(n))
	if groupCtx != "" {
		ctx = groupCtx + ": " + ctx
	}
	f := l.fields(n, ctx, caseKeys)
	c := Case{File: l.file, Line: n.Line, Group: group, Root: cfg.root, Shells: cfg.shells, Env: cfg.env, Match: MatchExact}

	c.Name = l.str(f["name"], ctx, "name", n)
	if c.Name != "" {
		if prev, dup := l.names[c.Name]; dup {
			l.errorf(n, ctx, "duplicate case name (first defined on line %d)", prev)
		} else {
			l.names[c.Name] = n.Line
		}
	}
	if e := f["env"]; e != nil {
		c.Env = l.env(e, ctx, cfg.env)
	}
	c.Input = l.str(f["input"], ctx, "input", n)
	c.Action = l.str(f["action"], ctx, "action", n)
	if f["action"] != nil && c.Action != ActionEnter && c.Action != ActionTab && c.Action != ActionKeys {
		l.errorf(f["action"], ctx, "action must be %q, %q or %q, got %q", ActionEnter, ActionTab, ActionKeys, c.Action)
	}
	if cfg.root == "" {
		l.errorf(n, ctx, "no root: set `root:` at the top of the file or in the group")
	}

	if c.Action == ActionKeys {
		l.keysFields(&c, f, n, ctx)
	} else if f["action"] != nil {
		for _, k := range []string{"tabs", "line", "candidates"} {
			if f[k] != nil {
				l.errorf(f[k], ctx, "%s applies only to action: keys", k)
			}
		}
		l.expectFields(&c, f, n, ctx)
	}

	l.cases = append(l.cases, c)
}

// expectFields validates exit, match and expect for actions enter and tab.
func (l *loader) expectFields(c *Case, f map[string]*yaml.Node, n *yaml.Node, ctx string) {
	if m := f["match"]; m != nil {
		c.Match = l.str(m, ctx, "match", m)
		if c.Match != MatchExact && c.Match != MatchNormalized {
			l.errorf(m, ctx, "match must be %q or %q, got %q", MatchExact, MatchNormalized, c.Match)
		}
	}
	if e := f["exit"]; e != nil {
		if c.Action == ActionTab {
			l.errorf(e, ctx, "exit applies only to action: enter")
		}
		v, err := strconv.Atoi(e.Value)
		if e.Kind != yaml.ScalarNode || err != nil || v < 0 || v > 255 {
			l.errorf(e, ctx, "exit must be an integer from 0 to 255, got %q", e.Value)
		}
		c.Exit = v
	}

	exp := f["expect"]
	switch {
	case exp == nil:
		l.errorf(n, ctx, "missing required key \"expect\"")
	case c.Action == ActionTab:
		items, ok := l.strList(exp, ctx, "expect for action: tab must be a list of candidates (use [] for none)")
		if ok {
			c.Expect = strings.Join(items, "\n")
		}
	case c.Action == ActionEnter:
		if exp.Kind != yaml.ScalarNode {
			l.errorf(exp, ctx, "expect for action: enter must be a string (use a | block for several lines)")
			break
		}
		c.Expect = strings.TrimRight(exp.Value, "\n")
	}
}

// keysFields validates tabs, line and candidates for action keys, which
// takes no expect, exit or match.
func (l *loader) keysFields(c *Case, f map[string]*yaml.Node, n *yaml.Node, ctx string) {
	for _, k := range []string{"expect", "exit", "match"} {
		if f[k] != nil {
			l.errorf(f[k], ctx, "%s does not apply to action: keys (use line and candidates)", k)
		}
	}
	if strings.ContainsAny(c.Input, "\t\r\n") {
		l.errorf(f["input"], ctx, "input for action: keys must be one line without tabs (Tab presses go in tabs)")
	}
	if t := f["tabs"]; t == nil {
		l.errorf(n, ctx, "missing required key \"tabs\"")
	} else {
		v, err := strconv.Atoi(t.Value)
		if t.Kind != yaml.ScalarNode || err != nil || v < 1 || v > maxTabs {
			l.errorf(t, ctx, "tabs must be an integer from 1 to %d, got %q", maxTabs, t.Value)
		}
		c.Tabs = v
	}
	if ln := f["line"]; ln == nil {
		l.errorf(n, ctx, "missing required key \"line\"")
	} else if ln.Kind != yaml.ScalarNode {
		l.errorf(ln, ctx, "line must be a string (quote it to keep a trailing space)")
	} else {
		c.WantLine = ln.Value
	}
	if cd := f["candidates"]; cd != nil {
		items, ok := l.strList(cd, ctx, "candidates must be a list (use [] for no listing)")
		c.Candidates, c.CheckCandidates = items, ok
	}
}

// strList returns a sequence of scalars as strings, reporting notList when n
// is not a sequence.
func (l *loader) strList(n *yaml.Node, ctx, notList string) ([]string, bool) {
	if n.Kind != yaml.SequenceNode {
		l.errorf(n, ctx, "%s", notList)
		return nil, false
	}
	items := []string{}
	ok := true
	for _, it := range n.Content {
		if it.Kind != yaml.ScalarNode {
			l.errorf(it, ctx, "list items must be strings")
			ok = false
			continue
		}
		items = append(items, it.Value)
	}
	return items, ok
}

// str returns the string value of a required scalar key, reporting a missing
// or non-scalar value against parent.
func (l *loader) str(v *yaml.Node, ctx, key string, parent *yaml.Node) string {
	if v == nil {
		l.errorf(parent, ctx, "missing required key %q", key)
		return ""
	}
	if v.Kind != yaml.ScalarNode {
		l.errorf(v, ctx, "%s must be a string", key)
		return ""
	}
	return v.Value
}

// peekName returns the scalar value of a mapping's "name" key, or "", so
// errors found before the key is validated still name what they are about.
func peekName(n *yaml.Node) string {
	if n.Kind != yaml.MappingNode {
		return ""
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == "name" && n.Content[i+1].Kind == yaml.ScalarNode {
			return n.Content[i+1].Value
		}
	}
	return ""
}

// label renders `kind "name"`, or just kind when the name is unknown.
func label(kind, name string) string {
	if name == "" {
		return kind
	}
	return fmt.Sprintf("%s %q", kind, name)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
