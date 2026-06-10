package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/adamgen/monom/internal/check"
	"github.com/adamgen/monom/internal/debuglog"
	"github.com/adamgen/monom/internal/filter"
	"github.com/adamgen/monom/internal/pack"
	"github.com/adamgen/monom/internal/root"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	debuglog.Log("[monomd] dispatch: args=(%s)", strings.Join(os.Args[1:], " "))

	switch os.Args[1] {
	case "filter":
		runFilter()
	case "root":
		runRoot()
	case "pack":
		runPack()
	case "check":
		runCheck()
	default:
		debuglog.Log("[monomd] unknown subcommand: %q", os.Args[1])
		fmt.Fprintf(os.Stderr, "monomd: unknown subcommand %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: monomd <subcommand> [args...]")
	fmt.Fprintln(os.Stderr, "subcommands: filter, root, pack, check")
}

// runFilter reads newline-delimited command paths from stdin, applies Filter
// with the words from os.Args[2:], and prints matching tokens one per line.
// Always exits 0 — any error results in empty output per spec.
func runFilter() {
	defer func() { recover() }() //nolint:errcheck

	words := os.Args[2:]

	var commands []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			commands = append(commands, line)
		}
	}
	// Ignore scanner.Err() — filter must always exit 0.

	debuglog.Log("[monomd filter] words=(%s) commands=%d", strings.Join(words, " "), len(commands))

	results := filter.Filter(commands, words)
	for _, r := range results {
		fmt.Println(r)
	}
	debuglog.Log("[monomd filter] returning %d token(s): (%s)", len(results), strings.Join(results, " "))
	os.Exit(0)
}

// runRoot prints the project root to stdout and exits 0, or prints an error
// to stderr and exits 1.
func runRoot() {
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		debuglog.Log("[monomd root] failed: %v", err)
		fmt.Fprintln(os.Stderr, "monomd root:", err)
		os.Exit(1)
	}
	debuglog.Log("[monomd root] found: %s", projectRoot)
	fmt.Println(projectRoot)
}

// runPack resolves os.Args[2:] to an absolute executable path and prints it,
// or prints an error to stderr and exits 1.
func runPack() {
	words := os.Args[2:]
	debuglog.Log("[monomd pack] words=(%s)", strings.Join(words, " "))
	absPath, err := pack.Pack(words)
	if err != nil {
		debuglog.Log("[monomd pack] failed: %v", err)
		fmt.Fprintln(os.Stderr, "monomd pack:", err)
		os.Exit(1)
	}
	debuglog.Log("[monomd pack] resolved: %s", absPath)
	fmt.Println(absPath)
}

// runCheck runs MONOM_USER_CONFIG complete, reports problems, and exits
// non-zero if any are found.
func runCheck() {
	userConfig := os.Getenv("MONOM_USER_CONFIG")
	debuglog.Log("[monomd check] config=%s", userConfig)
	problems, err := check.Check(userConfig)
	if err != nil {
		debuglog.Log("[monomd check] failed: %v", err)
		fmt.Fprintln(os.Stderr, "monomd check:", err)
		os.Exit(1)
	}
	if len(problems) == 0 {
		n := countLines(userConfig)
		debuglog.Log("[monomd check] OK: %d commands", n)
		fmt.Printf("✔ %d commands OK\n", n)
		return
	}
	debuglog.Log("[monomd check] %d problem(s) found", len(problems))
	for _, p := range problems {
		fmt.Println(p)
	}
	fmt.Fprintf(os.Stderr, "monomd check: %d problem(s) found\n", len(problems))
	os.Exit(1)
}

// countLines runs userConfig complete and counts non-empty output lines.
func countLines(userConfig string) int {
	if userConfig == "" {
		return 0
	}
	var out bytes.Buffer
	cmd := exec.Command(userConfig, "complete")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
