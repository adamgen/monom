package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/adamgen/monom/internal/check"
	"github.com/adamgen/monom/internal/cli"
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
		os.Exit(cli.ExitCodes.Error)
	}

	debuglog.Log("[mnmd] dispatch: args=(%s)", strings.Join(os.Args[1:], " "))

	var err error
	switch os.Args[1] {
	case "filter":
		runFilter()
		return
	case "root":
		err = runRoot()
	case "pack":
		err = runPack()
	case "check":
		err = runCheck()
	case "install":
		err = runInstall()
	default:
		debuglog.Log("[mnmd] unknown subcommand: %q", os.Args[1])
		fmt.Fprintf(os.Stderr, "mnmd: unknown subcommand %q\n", os.Args[1])
		usage()
		os.Exit(cli.ExitCodes.Error)
	}

	handleError(os.Args[1], err)
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

// handleError is the uniform error→exit-code dispatch tail. It resolves the
// exit code from a CodedError when present, defaults to ExitCodes.Error
// otherwise. The GroupError code suppresses stderr (it is a payload-free
// signal).
func handleError(sub string, err error) {
	if err == nil {
		return
	}
	var ce cli.CodedError
	if errors.As(err, &ce) {
		if ce.ExitCode() != cli.ExitCodes.GroupError {
			fmt.Fprintln(os.Stderr, "mnmd "+sub+":", ce)
		}
		os.Exit(ce.ExitCode())
	}
	fmt.Fprintln(os.Stderr, "mnmd "+sub+":", err)
	os.Exit(cli.ExitCodes.Error)
}

// runFilter always exits 0 — any error results in empty output per spec.
// It is exempt from the CodedError dispatch.
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

	debuglog.Log("[mnmd filter] words=(%s) commands=%d", strings.Join(words, " "), len(commands))

	results := filter.Filter(commands, words)
	for _, r := range results {
		fmt.Println(r)
	}
	debuglog.Log("[mnmd filter] returning %d token(s): (%s)", len(results), strings.Join(results, " "))
	os.Exit(cli.ExitCodes.Success)
}

func runRoot() error {
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		debuglog.Log("[mnmd root] failed: %v", err)
		return cli.WrapError(err)
	}
	debuglog.Log("[mnmd root] found: %s", projectRoot)
	fmt.Println(projectRoot)
	return nil
}

func runPack() error {
	words := os.Args[2:]
	debuglog.Log("[mnmd pack] words=(%s)", strings.Join(words, " "))
	absPath, err := pack.Pack(words)
	if err != nil {
		var ge *pack.GroupError
		if errors.As(err, &ge) {
			debuglog.Log("[mnmd pack] command group: %s", ge.Path)
			return err
		}
		debuglog.Log("[mnmd pack] failed: %v", err)
		return cli.WrapError(err)
	}
	debuglog.Log("[mnmd pack] resolved: %s", absPath)
	fmt.Println(absPath)
	return nil
}

func runCheck() error {
	userConfig := os.Getenv("_MONOM_USER_CONFIG")
	debuglog.Log("[mnmd check] config=%s", userConfig)
	problems, err := check.Check(userConfig)
	if err != nil {
		debuglog.Log("[mnmd check] failed: %v", err)
		return cli.WrapError(err)
	}
	if len(problems) == 0 {
		n := countLines(userConfig)
		debuglog.Log("[mnmd check] OK: %d commands", n)
		fmt.Printf("✔ %d commands OK\n", n)
		return nil
	}
	debuglog.Log("[mnmd check] %d problem(s) found", len(problems))
	for _, p := range problems {
		fmt.Println(p)
	}
	return cli.WrapError(fmt.Errorf("%d problem(s) found", len(problems)))
}

func runInstall() error {
	exe, err := os.Executable()
	if err != nil {
		return cli.WrapError(fmt.Errorf("could not determine binary path: %w", err))
	}
	if err := install.Run(exe); err != nil {
		return cli.WrapError(err)
	}
	return nil
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
