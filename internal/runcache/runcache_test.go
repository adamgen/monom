package runcache

import (
	"bytes"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A signal sent to the middleware itself reaches the command, and the
// middleware survives to return the command's status.
func TestRunRelaysSIGTERMAndSIGINT(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), sig)
		}()
		var out bytes.Buffer
		script := `trap 'echo got; exit 5' TERM INT; i=0; while [ $i -lt 50 ]; do sleep 0.1; i=$((i+1)); done; exit 9`
		status, err := Run([]string{"sh", "-c", script}, &out, &out)
		if err != nil || status != 5 || !strings.Contains(out.String(), "got") {
			t.Fatalf("%v: status %d, err %v, out %q", sig, status, err, out.String())
		}
	}
}
