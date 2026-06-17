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
	"github.com/adamgen/monom/internal/install"
	"github.com/adamgen/monom/internal/pack"
	"github.com/adamgen/monom/internal/root"
)

func main() {
	subcommand := ""
	if len(os.Args) >= 2 {
		subcommand = os.Args[1]
	}

	checkNudge(subcommand)

	if subcommand == "" {
		usage()
		os.Exit(1)
	}

	debuglog.Log("[mnmd] dispatch: args=(%s)", strings.Join(os.Args[1:], " "))

	switch os.Args[1] {
	case "filter":
		runFilter()
	case "root":
		runRoot()
	case "pack":
		runPack()
	case "check":
		runCheck()
	case "install":
		runInstall()
	default:
		debuglog.Log("[mnmd] unknown subcommand: %q", os.Args[1])
		fmt.Fprintf(os.Stderr, "mnmd: unknown subcommand %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mnmd <subcommand> [args...]")
	fmt.Fprintln(os.Stderr, "subcommands: filter, root, pack, check, install")
}

// checkNudge prints a hint to stderr when the shell integration is not active
// (MONOM_ACTIVE unset), except when the user is already running `mnmd install`.
func checkNudge(subcommand string) {
	if subcommand == "install" {
		return
	}
	if os.Getenv("MONOM_ACTIVE") == "" {
		fmt.Fprintln(os.Stderr, "hint: run 'mnmd install' to activate shell integration")
	}
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

	debuglog.Log("[mnmd filter] words=(%s) commands=%d", strings.Join(words, " "), len(commands))

	results := filter.Filter(commands, words)
	for _, r := range results {
		fmt.Println(r)
	}
	debuglog.Log("[mnmd filter] returning %d token(s): (%s)", len(results), strings.Join(results, " "))
	os.Exit(0)
}

// runRoot prints the project root to stdout and exits 0, or prints an error
// to stderr and exits 1.
func runRoot() {
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		debuglog.Log("[mnmd root] failed: %v", err)
		fmt.Fprintln(os.Stderr, "mnmd root:", err)
		os.Exit(1)
	}
	debuglog.Log("[mnmd root] found: %s", projectRoot)
	fmt.Println(projectRoot)
}

// runPack resolves os.Args[2:] to an absolute executable path and prints it,
// or prints an error to stderr and exits 1.
func runPack() {
	words := os.Args[2:]
	debuglog.Log("[mnmd pack] words=(%s)", strings.Join(words, " "))
	absPath, err := pack.Pack(words)
	if err != nil {
		debuglog.Log("[mnmd pack] failed: %v", err)
		fmt.Fprintln(os.Stderr, "mnmd pack:", err)
		os.Exit(1)
	}
	debuglog.Log("[mnmd pack] resolved: %s", absPath)
	fmt.Println(absPath)
}

// runCheck runs _MONOM_USER_CONFIG complete, reports problems, and exits
// non-zero if any are found.
func runCheck() {
	userConfig := os.Getenv("_MONOM_USER_CONFIG")
	debuglog.Log("[mnmd check] config=%s", userConfig)
	problems, err := check.Check(userConfig)
	if err != nil {
		debuglog.Log("[mnmd check] failed: %v", err)
		fmt.Fprintln(os.Stderr, "mnmd check:", err)
		os.Exit(1)
	}
	if len(problems) == 0 {
		n := countLines(userConfig)
		debuglog.Log("[mnmd check] OK: %d commands", n)
		fmt.Printf("✔ %d commands OK\n", n)
		return
	}
	debuglog.Log("[mnmd check] %d problem(s) found", len(problems))
	for _, p := range problems {
		fmt.Println(p)
	}
	fmt.Fprintf(os.Stderr, "mnmd check: %d problem(s) found\n", len(problems))
	os.Exit(1)
}

// runInstall writes a source line for src/monom into the user's shell rc file.
func runInstall() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "mnmd install: could not determine binary path:", err)
		os.Exit(1)
	}
	if err := install.Run(exe); err != nil {
		fmt.Fprintln(os.Stderr, "mnmd install:", err)
		os.Exit(1)
	}
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
