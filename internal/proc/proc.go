// Package proc runs a child process for middleware that must stay resident
// around it (to act after it exits, or to capture its output), and execs one
// for middleware that only passes through.
package proc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// forwarded are the signals Run relays to the child. A terminal's Ctrl-C
// reaches the child directly too (same process group), so on a Ctrl-C the
// child can see SIGINT twice; Run relays it anyway so a signal sent to the
// middleware alone (kill, a CI runner's timeout) still reaches the command.
var forwarded = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT}

// Run runs argv with stdin inherited and stdout/stderr written to the given
// writers, relaying signals until it exits. It returns the child's exit
// status, or 128+n when a signal killed it, like a shell reports it.
func Run(argv []string, stdout, stderr io.Writer) (int, error) {
	if len(argv) == 0 {
		return 0, errors.New("nothing to run")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr

	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, forwarded...)
	defer signal.Stop(sigs)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	done := make(chan struct{})
	go func() {
		for {
			select {
			case s := <-sigs:
				_ = cmd.Process.Signal(s)
			case <-done:
				return
			}
		}
	}()
	err := cmd.Wait()
	close(done)
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return 0, err
	}
	if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal()), nil
	}
	return ee.ExitCode(), nil
}

// Exec replaces the current process with argv: the middleware steps out of
// the way, so the terminal, signals and stdio reach what follows untouched.
func Exec(argv []string) error {
	if len(argv) == 0 {
		return errors.New("nothing to run")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}
	err = syscall.Exec(path, argv, os.Environ())
	return fmt.Errorf("exec %s: %w", argv[0], err)
}
