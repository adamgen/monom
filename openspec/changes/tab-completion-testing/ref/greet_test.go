package greet_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
)

func scriptDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	return filepath.Dir(file)
}

type session struct {
	ptm  *os.File
	cmd  *exec.Cmd
	mu   sync.Mutex
	buf  bytes.Buffer // accumulates all output from the pty
	done chan struct{}
}

func newSession(t *testing.T) *session {
	t.Helper()
	dir := scriptDir(t)

	cmd := exec.Command("bash", "--norc", "--noprofile", "-i")
	cmd.Env = append(os.Environ(), "TERM=xterm", "BASH_SILENCE_DEPRECATION_WARNING=1")

	ptm, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("pty.Start: %v", err)
	}

	s := &session{ptm: ptm, cmd: cmd, done: make(chan struct{})}

	// Read from pty in a background goroutine so reads never block the test.
	go func() {
		defer close(s.done)
		tmp := make([]byte, 256)
		for {
			n, err := ptm.Read(tmp)
			if n > 0 {
				s.mu.Lock()
				s.buf.Write(tmp[:n])
				s.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	s.waitFor(t, regexp.MustCompile(`\$\s`))
	s.send(t, "stty rows 24 cols 200\r")
	s.waitFor(t, regexp.MustCompile(`\$\s`))
	s.send(t, "source "+filepath.Join(dir, "greet")+"\r")
	s.waitFor(t, regexp.MustCompile(`\$\s`))
	return s
}

func (s *session) send(t *testing.T, input string) {
	t.Helper()
	if _, err := s.ptm.WriteString(input); err != nil {
		t.Fatalf("send %q: %v", input, err)
	}
}

// waitFor polls the accumulated buffer until the pattern matches or 5s passes.
func (s *session) waitFor(t *testing.T, re *regexp.Regexp) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		got := s.buf.String()
		s.mu.Unlock()
		if re.MatchString(got) {
			return got
		}
		time.Sleep(20 * time.Millisecond)
	}
	s.mu.Lock()
	got := s.buf.String()
	s.mu.Unlock()
	t.Fatalf("timeout waiting for %q\nbuffer: %q", re, got)
	return ""
}

func (s *session) close() {
	s.ptm.WriteString("exit\r")
	// Close the pty master; this causes the reader goroutine to exit.
	s.ptm.Close()
	<-s.done
	s.cmd.Wait()
}

// ── Tests ──────────────────────────────────────────────────────────────────────

func TestGreetTabCompletesAlice(t *testing.T) {
	s := newSession(t)
	defer s.close()

	s.send(t, "greet al\t") // "al" unambiguously matches only "alice"
	s.waitFor(t, regexp.MustCompile(`alice`))
}

func TestGreetTabCompletesBob(t *testing.T) {
	s := newSession(t)
	defer s.close()

	s.send(t, "greet b\t")
	s.waitFor(t, regexp.MustCompile(`bob`))
}

func TestGreetTabCompletesCarol(t *testing.T) {
	s := newSession(t)
	defer s.close()

	s.send(t, "greet c\t")
	s.waitFor(t, regexp.MustCompile(`carol`))
}

func TestGreetDoubleTabListsANames(t *testing.T) {
	s := newSession(t)
	defer s.close()

	s.send(t, "greet a\t\t") // double Tab — lists alice and arthur
	s.waitFor(t, regexp.MustCompile(`alice`))
	s.waitFor(t, regexp.MustCompile(`arthur`))
}

func TestGreetDoubleTabListsAllNames(t *testing.T) {
	s := newSession(t)
	defer s.close()

	s.send(t, "greet \t\t")
	for _, name := range []string{"alice", "arthur", "bob", "carol", "dave"} {
		if s.waitFor(t, regexp.MustCompile(name)) == "" {
			t.Errorf("name %q not found in completions", name)
		}
	}
}

// Ensure the test binary itself doesn't silently pass if fmt is unused.
var _ = fmt.Sprintf
