// Package mapfile implements the command map: an optional monom-map.json at
// the project root that maps command paths to executable scripts located
// elsewhere in the project. It is a discovery/routing backend for the user
// config's `complete` and `run` hooks — `mnmd pack` remains the sole resolver;
// this package only rewrites the user's tokens into the tokens pack receives.
//
// The node model mirrors the file tree protocol: a string value is a command
// (shorthand for {"run": <target>}), an object with a "run" key is a command,
// and any other object is a category whose keys are children. "$note" is
// allowed anywhere and ignored — it stands in for the comments JSON lacks.
package mapfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adamgen/monom/internal/cli"
)

// FileName is the command map's file name, looked up at the project root.
const FileName = "monom-map.json"

// reservedNames are invalid as command or category names at every level.
// "run" marks a command node today; the others are held for future
// per-command keys (completion, lifecycle hooks) so that adding them later
// is not a breaking change.
var reservedNames = map[string]bool{
	"run":      true,
	"complete": true,
	"pre-run":  true,
	"post-run": true,
}

// Node is one node of the command map tree. Exactly one of Run/Children is
// meaningful: a command node has a Run target and nil Children; a category
// node has a (possibly empty) Children map and an empty Run.
type Node struct {
	Run      string
	Children map[string]*Node
}

// IsCommand reports whether the node is a command node.
func (n *Node) IsCommand() bool { return n.Children == nil }

// GroupError signals that the resolved tokens name a category node — the
// map's equivalent of pack's directory outcome. Payload-free by the same
// contract: the caller sources the child listing from `complete | mnmd filter`.
type GroupError struct {
	Path string // slash-joined command path of the category
}

func (e *GroupError) Error() string { return "resolved path is a command group: " + e.Path }
func (e *GroupError) ExitCode() int { return cli.ExitCodes.GroupError }

// Load reads and parses the command map at rootDir. It is an error for the
// file to be absent: a config that delegates to `mnmd map` without a map file
// is misconfigured, and silence would hide that.
func Load(rootDir string) (*Node, error) {
	path := filepath.Join(rootDir, FileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s found at project root %s", FileName, rootDir)
		}
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()
	return Parse(f)
}

// Parse decodes a command map from r. It uses a token-stream walk rather than
// json.Unmarshal so that duplicate keys — which Unmarshal silently collapses
// to the last occurrence — are a hard error instead of a silently swallowed
// command.
func Parse(r io.Reader) (*Node, error) {
	dec := json.NewDecoder(r)

	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("top level must be a JSON object")
	}

	root, err := parseObject(dec, "")
	if err != nil {
		return nil, err
	}
	if root.IsCommand() {
		return nil, fmt.Errorf(`top level must be a category, not a command (remove the top-level "run" key)`)
	}
	return root, nil
}

// parseObject consumes one JSON object (the opening '{' already read) and
// returns the node it describes. path is the slash-joined command path of the
// node, "" for the root; it exists only for error messages.
func parseObject(dec *json.Decoder, path string) (*Node, error) {
	at := func() string {
		if path == "" {
			return "top level"
		}
		return fmt.Sprintf("%q", path)
	}

	seen := map[string]bool{}
	runTarget := ""
	children := map[string]*Node{}

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		key := keyTok.(string)

		if seen[key] {
			return nil, fmt.Errorf("duplicate key %q at %s", key, at())
		}
		seen[key] = true

		switch {
		case key == "$note":
			if _, err := stringValue(dec); err != nil {
				return nil, fmt.Errorf("$note at %s must be a string", at())
			}

		case strings.HasPrefix(key, "$"):
			return nil, fmt.Errorf("unknown metadata key %q at %s (only $note is recognized)", key, at())

		case key == "run":
			target, err := stringValue(dec)
			if err != nil {
				return nil, fmt.Errorf("\"run\" at %s must be a string target path", at())
			}
			if err := validateTarget(target); err != nil {
				return nil, fmt.Errorf("run target at %s: %w", at(), err)
			}
			runTarget = target

		case reservedNames[key]:
			return nil, fmt.Errorf("key %q at %s is reserved for future per-command hooks", key, at())

		default:
			if err := validateName(key); err != nil {
				return nil, fmt.Errorf("invalid name %q at %s: %w", key, at(), err)
			}
			childPath := key
			if path != "" {
				childPath = path + "/" + key
			}
			child, err := parseValue(dec, childPath)
			if err != nil {
				return nil, err
			}
			children[key] = child
		}
	}

	// Consume the closing '}'.
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if runTarget != "" && len(children) > 0 {
		return nil, fmt.Errorf("%s has both a \"run\" target and children — a node is a command or a category, not both", at())
	}
	if runTarget != "" {
		return &Node{Run: runTarget}, nil
	}
	return &Node{Children: children}, nil
}

// parseValue consumes one value: a string (command shorthand) or an object.
func parseValue(dec *json.Decoder, path string) (*Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	switch v := tok.(type) {
	case string:
		if err := validateTarget(v); err != nil {
			return nil, fmt.Errorf("run target at %q: %w", path, err)
		}
		return &Node{Run: v}, nil
	case json.Delim:
		if v == '{' {
			return parseObject(dec, path)
		}
	}
	return nil, fmt.Errorf("value at %q must be a string target or an object", path)
}

// stringValue consumes one value and requires it to be a JSON string.
func stringValue(dec *json.Decoder) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", err
	}
	s, ok := tok.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %v", tok)
	}
	return s, nil
}

// validateName checks a command or category name: one path segment as the
// shell and `mnmd filter` see it.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty")
	}
	if strings.ContainsAny(name, " \t\n") {
		return fmt.Errorf("name contains whitespace")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("name contains a slash — nest categories instead")
	}
	return nil
}

// validateTarget checks a run target: a root-relative path whose segments
// survive the whitespace-tokenized run-hook → pack handoff and cannot escape
// the project root. Existence and executability are deliberately NOT checked
// here — pack validates them at resolution time, and `mnmd check` reports
// them at development time.
func validateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("target is empty")
	}
	if strings.ContainsAny(target, " \t\n") {
		return fmt.Errorf("target %q contains whitespace", target)
	}
	if filepath.IsAbs(target) {
		return fmt.Errorf("target %q is absolute — targets are relative to the project root", target)
	}
	clean := filepath.Clean(target)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("target %q escapes the project root", target)
	}
	return nil
}

// Complete returns the slash-joined paths of all command nodes, sorted.
// This is the `complete` hook wire format.
func Complete(root *Node) []string {
	var paths []string
	Walk(root, func(path string, n *Node) {
		if n.IsCommand() {
			paths = append(paths, path)
		}
	})
	sort.Strings(paths)
	return paths
}

// Walk visits every node below root in sorted key order, calling fn with the
// node's slash-joined command path.
func Walk(root *Node, fn func(path string, n *Node)) {
	walk(root, "", fn)
}

func walk(n *Node, prefix string, fn func(path string, n *Node)) {
	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		child := n.Children[k]
		path := k
		if prefix != "" {
			path = prefix + "/" + k
		}
		fn(path, child)
		if !child.IsCommand() {
			walk(child, path, fn)
		}
	}
}

// Resolve looks up the user's typed words in the map and returns the run
// target as space-separated path tokens — the run-hook output format that
// `mnmd pack` re-joins with slashes.
//
// Outcomes:
//   - command node matched: the target's tokens, nil error.
//   - category node matched (including zero words = the root): "" and a
//     *GroupError — the payload-free group signal, exit code 3.
//   - no match (unknown name, or words continuing past a command node): ""
//     and nil error. The caller prints nothing and exits 0, which triggers
//     the run-hook fallback — unmapped commands drop through to the file
//     tree, making the map an overlay rather than a replacement.
func Resolve(root *Node, words []string) (string, error) {
	cur := root
	for _, w := range words {
		if cur.IsCommand() {
			return "", nil // words continue past a command — not ours to resolve
		}
		child, ok := cur.Children[w]
		if !ok {
			return "", nil
		}
		cur = child
	}
	if cur.IsCommand() {
		return strings.ReplaceAll(cur.Run, "/", " "), nil
	}
	return "", &GroupError{Path: strings.Join(words, "/")}
}
