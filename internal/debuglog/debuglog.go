package debuglog

import (
	"fmt"
	"os"
	"time"
)

// Log appends a timestamped line to the file named by $MONOM_DEBUG_LOG.
// It is a no-op when MONOM_DEBUG_LOG is unset or empty.
// Any file I/O error is silently ignored — debug logging must never affect
// the normal execution path.
func Log(format string, args ...any) {
	path := os.Getenv("MONOM_DEBUG_LOG")
	if path == "" {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	ts := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f, "[%s] %s\n", ts, msg) //nolint:errcheck
}
