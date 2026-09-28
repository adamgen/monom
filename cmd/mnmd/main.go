package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamgen/monom/internal/check"
	"github.com/adamgen/monom/internal/cli"
	"github.com/adamgen/monom/internal/config"
	"github.com/adamgen/monom/internal/debuglog"
	"github.com/adamgen/monom/internal/discover"
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
	case "discover":
		err = runDiscover()
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
	fmt.Fprintln(os.Stderr, "subcommands: filter, root, pack, discover, check, install")
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

// runDiscover prints the registered command paths found by default discovery,
// in the same format as the `complete` hook.
func runDiscover() error {
	projectRoot, err := root.FindProjectRoot()
	if err != nil {
		debuglog.Log("[mnmd discover] no root: %v", err)
		return cli.WrapError(err)
	}
	cfg, err := config.Load(filepath.Join(projectRoot, root.ConfigFileName))
	if err != nil {
		debuglog.Log("[mnmd discover] config: %v", err)
		return cli.WrapError(err)
	}
	// A hook script's settings come from its config hook. The shell only calls
	// discover when the complete hook printed nothing, so this runs at most
	// one extra hook per Tab, and only in hook projects without a complete.
	settings := cfg.ProjectSettings()
	res := discover.Discover(projectRoot, cfg.Declared, settings.Hide)
	debuglog.Log("[mnmd discover] root=%s registered=%d skipped=%d", projectRoot, len(res.Commands), len(res.Skipped))
	for _, p := range res.Paths() {
		fmt.Println(p)
	}
	return nil
}

// runCheck is the doctor. Warnings are printed but never fail the run; any
// error-severity problem makes it exit non-zero.
func runCheck() error {
	projectRoot, rootErr := root.FindProjectRoot()
	userConfig := os.Getenv("_MONOM_USER_CONFIG")
	if userConfig == "" {
		if rootErr != nil {
			debuglog.Log("[mnmd check] no root: %v", rootErr)
			return cli.WrapError(rootErr)
		}
		userConfig = filepath.Join(projectRoot, root.ConfigFileName)
	}
	debuglog.Log("[mnmd check] root=%s config=%s", projectRoot, userConfig)

	report, err := check.Check(check.Input{
		Root:         projectRoot,
		UserConfig:   userConfig,
		UserSeverity: os.Getenv(check.UserSeverityEnv),
	})
	if err != nil {
		debuglog.Log("[mnmd check] failed: %v", err)
		return cli.WrapError(err)
	}

	for _, p := range report.Problems {
		fmt.Println(p)
	}
	errs := report.Count(config.SeverityError)
	warns := report.Count(config.SeverityWarning)
	debuglog.Log("[mnmd check] commands=%d errors=%d warnings=%d", len(report.Commands), errs, warns)
	if errs > 0 {
		return cli.WrapError(fmt.Errorf("%d error(s), %d warning(s)", errs, warns))
	}
	summary := fmt.Sprintf("✔ %d commands OK", len(report.Commands))
	if report.Discovered {
		summary += " (default discovery)"
	}
	if warns > 0 {
		summary += fmt.Sprintf(", %d warning(s)", warns)
	}
	fmt.Println(summary)
	return nil
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
