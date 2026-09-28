// Package testcases loads and strictly validates the declarative CLI test
// cases in tests/cases/*.yaml. It is test tooling: nothing in mnmd imports it.
//
// Schema (see tests/README.md for the full guide):
//
//	root: fixtures/demo-project     # required unless every group sets it
//	shells: [bash, zsh]             # optional; default [bash, zsh]
//	cases:                          # optional
//	  - name: <unique in the file>  # required
//	    input: <CLI line>           # required
//	    action: enter | tab         # required
//	    exit: <int>                 # optional, enter only, default 0
//	    match: exact | normalized   # optional, default exact
//	    expect: <string> | [list]   # required: a string for enter, a list for tab
//	groups:                         # optional
//	  - name: <group name>          # required
//	    root: <dir>                 # optional override
//	    shells: [bash]              # optional override
//	    cases: [...]                # required
package testcases

import (
	"fmt"
	"os"
	"path/filepath"
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
	Input  string
	Action string // ActionEnter or ActionTab
	Expect string // tab candidates are newline-joined
	Exit   int
	Match  string // MatchExact or MatchNormalized
}

const (
	ActionEnter     = "enter"
	ActionTab       = "tab"
	MatchExact      = "exact"
	MatchNormalized = "normalized"
)

var (
	fileKeys  = []string{"root", "shells", "cases", "groups"}
	groupKeys = []string{"name", "root", "shells", "cases"}
	caseKeys  = []string{"name", "input", "action", "exit", "match", "expect"}
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
}

type loader struct {
	file     string
	repoRoot string
	errs     []lineError
	cases    []Case
	names    map[string]int
}

// Load parses and validates one case file. repoRoot is used to check that
// every root is a directory. On any problem it returns all of them as Errors,
// each prefixed with file:line and, where known, the group and case.
func Load(file, repoRoot string) ([]Case, error) {
	data, err := os.ReadFile(file)
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
	return cfg
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
	c := Case{File: l.file, Line: n.Line, Group: group, Root: cfg.root, Shells: cfg.shells, Match: MatchExact}

	c.Name = l.str(f["name"], ctx, "name", n)
	if c.Name != "" {
		if prev, dup := l.names[c.Name]; dup {
			l.errorf(n, ctx, "duplicate case name (first defined on line %d)", prev)
		} else {
			l.names[c.Name] = n.Line
		}
	}
	c.Input = l.str(f["input"], ctx, "input", n)
	c.Action = l.str(f["action"], ctx, "action", n)
	if f["action"] != nil && c.Action != ActionEnter && c.Action != ActionTab {
		l.errorf(f["action"], ctx, "action must be %q or %q, got %q", ActionEnter, ActionTab, c.Action)
	}
	if cfg.root == "" {
		l.errorf(n, ctx, "no root: set `root:` at the top of the file or in the group")
	}

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
		if exp.Kind != yaml.SequenceNode {
			l.errorf(exp, ctx, "expect for action: tab must be a list of candidates (use [] for none)")
			break
		}
		var items []string
		for _, it := range exp.Content {
			if it.Kind != yaml.ScalarNode {
				l.errorf(it, ctx, "tab candidates must be strings")
				continue
			}
			items = append(items, it.Value)
		}
		c.Expect = strings.Join(items, "\n")
	case c.Action == ActionEnter:
		if exp.Kind != yaml.ScalarNode {
			l.errorf(exp, ctx, "expect for action: enter must be a string (use a | block for several lines)")
			break
		}
		c.Expect = strings.TrimRight(exp.Value, "\n")
	}

	l.cases = append(l.cases, c)
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
