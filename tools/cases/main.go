// Command cases is test tooling for tests/monom_cases_test: it loads and
// validates tests/cases/*.yaml and prints each case as a `case_add` line the
// bash runner evaluates. Validation problems go to stderr with exit 1.
//
//	cases -repo <repo root> <file.yaml>...
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adamgen/monom/internal/testcases"
)

func main() {
	repo := flag.String("repo", ".", "repository root; case roots are relative to it")
	flag.Parse()

	failed := false
	var out strings.Builder
	for _, file := range flag.Args() {
		cases, err := testcases.Load(file, *repo)
		if err != nil {
			failed = true
			var errs testcases.Errors
			if !errors.As(err, &errs) {
				errs = testcases.Errors{err.Error()}
			}
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "FATAL:", e)
			}
			continue
		}
		for _, c := range cases {
			fields := []string{
				c.Name, c.File + ":" + strconv.Itoa(c.Line), c.Root, strings.Join(c.Shells, " "),
				c.Input, c.Action, c.Expect, strconv.Itoa(c.Exit), c.Match,
			}
			out.WriteString("case_add")
			for _, f := range fields {
				out.WriteString(" " + shQuote(f))
			}
			out.WriteString("\n")
		}
	}
	if failed {
		os.Exit(1)
	}
	fmt.Print(out.String())
}

// shQuote single-quotes s for bash.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
