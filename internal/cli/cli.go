// Package cli provides the CodedError interface and the central exit-code
// registry for mnmd. Every exit code used by the binary is defined here — no
// integer literals appear elsewhere.
package cli

// ExitCodes is the single source of truth for all exit codes used by mnmd.
//
//   - Success (0): leaf resolved / normal output.
//   - Error (1): generic real error (no args, not found, not executable,
//     no project root, install/check failures).
//   - GroupError (3): pack command-group signal — payload-free, reserved
//     exclusively for the pack subcommand.
var ExitCodes = struct {
	Success    int
	Error      int
	GroupError int
}{
	Success:    0,
	Error:      1,
	GroupError: 3,
}

// CodedError is an error that carries its own exit code. Subcommand outcomes
// are determined by the typed error they return, not by the call site in
// main.go.
type CodedError interface {
	error
	ExitCode() int
}

// wrappedError is a generic CodedError that wraps an existing error with the
// registry's Error code (1).
type wrappedError struct {
	code int
	err  error
}

func (w *wrappedError) Error() string { return w.err.Error() }
func (w *wrappedError) ExitCode() int { return w.code }

// WrapError wraps err as a CodedError with ExitCodes.Error (1).
func WrapError(err error) CodedError {
	return &wrappedError{code: ExitCodes.Error, err: err}
}
