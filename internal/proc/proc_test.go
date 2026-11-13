package proc

import (
	"bytes"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunReturnsTheChildsStatus(t *testing.T) {
	var out bytes.Buffer
	status, err := Run([]string{"sh", "-c", "echo hi; exit 7"}, &out, &out)
	if err != nil || status != 7 || out.String() != "hi\n" {
		t.Fatalf("status %d, err %v, out %q", status, err, out.String())
	}
}

func TestRunReportsADeathBySignalAs128PlusN(t *testing.T) {
	status, err := Run([]string{"sh", "-c", "kill -KILL $$"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || status != 128+int(syscall.SIGKILL) {
		t.Fatalf("status %d, err %v", status, err)
	}
}

// A signal sent to the middleware's own process reaches the child, and the
// middleware survives to report the child's status.
func TestRunForwardsSIGTERMAndSIGINTToTheChild(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		var out bytes.Buffer
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), sig)
		}()
		script := `trap 'echo got; exit 5' TERM INT; echo ready; i=0; while [ $i -lt 50 ]; do sleep 0.1; i=$((i+1)); done; exit 9`
		status, err := Run([]string{"sh", "-c", script}, &out, &out)
		if err != nil || status != 5 || !strings.Contains(out.String(), "got") {
			t.Fatalf("%v: status %d, err %v, out %q", sig, status, err, out.String())
		}
	}
}
