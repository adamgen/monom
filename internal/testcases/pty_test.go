//go:build cases

// Interactive shell sessions for action: keys. A case types its input into a
// real interactive bash or zsh running in a pseudo-terminal, presses Tab, and
// reads back the line editor's buffer and the rendered screen.
//
// Waiting never sleeps for a fixed time. The shells write invisible OSC 2
// (window title) sentinels in-band with their output, so once a sentinel has
// been read, everything the shell drew before it has been rendered:
//   - CASE-STATUS:<n>  before every prompt (PROMPT_COMMAND / precmd), with $?
//   - CASE-READY       inside the prompt itself, so it is drawn only after
//     readline/zle put the tty in raw mode (keys sent earlier would be echoed
//     by the tty driver)
//   - CASE-LINE:<buffer>  from a key bound to Ctrl-], which dumps the edit
//     buffer without changing it
package testcases

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/hinshun/vt10x"
)

const (
	ptyCols, ptyRows = 100, 24
	ptyWait          = 5 * time.Second
	ptyPrompt        = "PROMPT$ "
	dumpKey          = "\x1d" // Ctrl-]
)

var (
	statusRe = regexp.MustCompile(`\x1b\]2;CASE-STATUS:(\d+)\x07`)
	readyRe  = regexp.MustCompile(`\x1b\]2;CASE-READY\x07`)
	lineRe   = regexp.MustCompile(`\x1b\]2;CASE-LINE:([^\x07]*)\x07`)
)

// Pinned readline settings: no bracketed paste or bell, the default
// two-Tab listing, and no pager or colours in listings.
const inputrc = `set enable-bracketed-paste off
set bell-style none
set show-all-if-ambiguous off
set show-all-if-unmodified off
set completion-query-items 1000
set page-completions off
set colored-stats off
set colored-completion-prefix off
set horizontal-scroll-mode off
set mark-directories on
`

const bashrc = `PS1='\[\e]2;CASE-READY\a\]PROMPT$ '
PS2=''
unset HISTFILE
PROMPT_COMMAND='printf "\033]2;CASE-STATUS:%%s\007" "$?"'
__case_dump() { printf '\033]2;CASE-LINE:%%s\007' "$READLINE_LINE" > /dev/tty; }
bind -x '"\C-]": __case_dump'
cd %s || exit 97
source %s
`

// zsh: pinned options, so a user's defaults can't change the menu behaviour:
// the first Tab on an ambiguous word lists, the second starts menu
// completion and inserts the first candidate.
const zshrc = `PROMPT=$'%%{\e]2;CASE-READY\a%%}PROMPT$ '
RPROMPT=''
PROMPT_EOL_MARK=''
HISTFILE=''
emulate -R zsh
setopt auto_menu auto_list list_ambiguous no_menu_complete no_list_beep no_beep
zmodload zsh/zle && unset zle_bracketed_paste
autoload -Uz compinit && compinit -C -d %s
precmd() { print -rn -- $'\e]2;CASE-STATUS:'"$?"$'\a' }
__case_dump() { print -rn -- $'\e]2;CASE-LINE:'"$BUFFER"$'\a' > /dev/tty }
zle -N __case_dump
bindkey '^]' __case_dump
cd %s || exit 97
source %s
`

// buildCompdump runs compinit once into dir/.zcompdump, so each zsh session
// can load it with compinit -C instead of scanning $fpath again.
func buildCompdump(t *testing.T, dir string) string {
	t.Helper()
	dump := filepath.Join(dir, ".zcompdump")
	cmd := exec.Command("zsh", "-fc", "autoload -Uz compinit && compinit -d "+shQuote(dump))
	cmd.Env = []string{"HOME=" + dir, "PATH=" + os.Getenv("PATH")}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building zsh compdump: %v\n%s", err, out)
	}
	return dump
}

type ptySession struct {
	t    *testing.T
	sh   string
	f    *os.File
	cmd  *exec.Cmd
	term vt10x.Terminal
	done chan struct{}

	mu   sync.Mutex
	raw  bytes.Buffer
	seen int // offset in raw consumed by the last wait
}

// startPTY starts an interactive shell set up like a user's rc: cd into root,
// source src/monom (zsh: after compinit), and waits for the first prompt.
func startPTY(t *testing.T, repo, root, sh, compdump string) *ptySession {
	t.Helper()
	home := t.TempDir()
	dir := shQuote(filepath.Join(repo, root))
	src := shQuote(filepath.Join(repo, "src", "monom"))
	env := []string{
		"HOME=" + home, "PATH=" + os.Getenv("PATH"), "TERM=xterm", "LANG=C.UTF-8",
		fmt.Sprintf("COLUMNS=%d", ptyCols), fmt.Sprintf("LINES=%d", ptyRows),
	}
	var cmd *exec.Cmd
	switch sh {
	case "bash":
		writeFile(t, filepath.Join(home, ".inputrc"), inputrc)
		writeFile(t, filepath.Join(home, ".bashrc"), fmt.Sprintf(bashrc, dir, src))
		env = append(env, "INPUTRC="+filepath.Join(home, ".inputrc"))
		cmd = exec.Command("bash", "--noprofile", "--rcfile", filepath.Join(home, ".bashrc"), "-i")
	case "zsh":
		writeFile(t, filepath.Join(home, ".zshrc"), fmt.Sprintf(zshrc, shQuote(compdump), dir, src))
		env = append(env, "ZDOTDIR="+home)
		cmd = exec.Command("zsh", "--no-globalrcs", "-i")
	}
	cmd.Env = env
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: ptyCols, Rows: ptyRows})
	if err != nil {
		t.Fatalf("starting %s in a pty: %v", sh, err)
	}
	s := &ptySession{t: t, sh: sh, f: f, cmd: cmd, term: vt10x.New(vt10x.WithSize(ptyCols, ptyRows)), done: make(chan struct{})}
	go s.pump()
	t.Cleanup(s.close)
	if st := s.waitPrompt(); st != "0" {
		s.fail(fmt.Sprintf("rc setup exited %s", st))
	}
	return s
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// pump feeds the shell's output to the raw log and the terminal emulator, in
// order.
func (s *ptySession) pump() {
	defer close(s.done)
	buf := make([]byte, 4096)
	for {
		n, err := s.f.Read(buf)
		if n > 0 {
			s.mu.Lock()
			s.raw.Write(buf[:n])
			_, _ = s.term.Write(buf[:n])
			s.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (s *ptySession) close() {
	_, _ = s.f.Write([]byte("\x15\x04")) // Ctrl-U clears the line; Ctrl-D exits
	select {
	case <-s.done:
	case <-time.After(ptyWait):
		_ = s.cmd.Process.Kill()
		<-s.done
	}
	_ = s.cmd.Wait()
	_ = s.f.Close()
}

func (s *ptySession) send(keys string) {
	if _, err := io.WriteString(s.f, keys); err != nil {
		s.t.Fatalf("writing to the pty: %v", err)
	}
}

// wait blocks until re matches output that arrived after the previous wait
// and returns its submatches, or fails the test after ptyWait with the
// screen and the raw log.
func (s *ptySession) wait(re *regexp.Regexp, what string) []string {
	s.t.Helper()
	deadline := time.Now().Add(ptyWait)
	for {
		s.mu.Lock()
		b := s.raw.Bytes()[s.seen:]
		if loc := re.FindSubmatchIndex(b); loc != nil {
			m := make([]string, 0, len(loc)/2)
			for i := 0; i < len(loc); i += 2 {
				m = append(m, string(b[loc[i]:loc[i+1]]))
			}
			s.seen += loc[1]
			s.mu.Unlock()
			return m
		}
		s.mu.Unlock()
		if time.Now().After(deadline) {
			s.fail(fmt.Sprintf("timed out after %s waiting for %s", ptyWait, what))
		}
		time.Sleep(2 * time.Millisecond) // poll the in-memory log
	}
}

func (s *ptySession) fail(msg string) {
	s.t.Helper()
	s.t.Fatalf("%s (%s)\n  screen:\n%s\n  raw output:\n    %q", msg, s.sh, indent(s.screen(), "    | "), s.rawTail(2000))
}

// waitPrompt waits for the next prompt, returning the previous exit status.
func (s *ptySession) waitPrompt() string {
	st := s.wait(statusRe, "prompt")[1]
	s.wait(readyRe, "line editor ready")
	return st
}

// press types input followed by tabs Tab presses and returns the edit buffer
// once the shell has processed them.
func (s *ptySession) press(input string, tabs int) string {
	s.send(input + strings.Repeat("\t", tabs) + dumpKey)
	line := s.wait(lineRe, "line dump")[1]
	if s.sh == "bash" {
		// bash clears the line before a bind -x command and redraws prompt
		// and buffer after it; wait for the redraw so the screen is whole.
		s.wait(regexp.MustCompile(regexp.QuoteMeta(ptyPrompt+line)), "bash redraw")
	}
	return line
}

// screen returns the rendered screen with trailing blanks trimmed.
func (s *ptySession) screen() string {
	lines := strings.Split(s.term.String(), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \x00")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func (s *ptySession) rawTail(n int) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	b := s.raw.Bytes()
	if len(b) > n {
		b = b[len(b)-n:]
	}
	return string(b)
}

// listing returns the completion candidates shown on screen: the words in
// the rows after the first prompt row, up to the next prompt row (bash
// redraws the prompt below its listing; zsh lists below the line and keeps
// the cursor there). Layout is ignored.
func listing(screen string) []string {
	prompt := strings.TrimSpace(ptyPrompt)
	rows := strings.Split(screen, "\n")
	start := -1
	for i, r := range rows {
		if strings.HasPrefix(r, prompt) {
			start = i
			break
		}
	}
	words := []string{}
	if start < 0 {
		return words
	}
	for _, r := range rows[start+1:] {
		if strings.HasPrefix(r, prompt) {
			break
		}
		words = append(words, strings.Fields(r)...)
	}
	return words
}

// asSet sorts and de-duplicates.
func asSet(words []string) []string {
	out := append([]string{}, words...)
	sort.Strings(out)
	j := 0
	for i, w := range out {
		if i == 0 || w != out[j-1] {
			out[j] = w
			j++
		}
	}
	return out[:j]
}

func indent(s, prefix string) string {
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}

// runKeysCase runs one action: keys case.
func runKeysCase(t *testing.T, repo, compdump string, c Case, sh string) {
	s := startPTY(t, repo, c.Root, sh, compdump)
	line := s.press(c.Input, c.Tabs)
	screen := s.screen()
	got := asSet(listing(screen))

	var problems []string
	if line != c.WantLine {
		problems = append(problems, "line differs")
	}
	want := asSet(c.Candidates)
	if c.CheckCandidates && strings.Join(want, "\n") != strings.Join(got, "\n") {
		problems = append(problems, "candidates differ")
	}
	if len(problems) == 0 {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n  case:   %s  (%s:%d)\n", c.Name, c.File, c.Line)
	fmt.Fprintf(&b, "  shell:  %s    root: %s    terminal: %dx%d\n", sh, c.Root, ptyCols, ptyRows)
	fmt.Fprintf(&b, "  input:  %q\n", c.Input)
	fmt.Fprintf(&b, "  action: keys    tabs: %d\n", c.Tabs)
	fmt.Fprintf(&b, "  result: %s\n", strings.Join(problems, "; "))
	fmt.Fprintf(&b, "  line:       expected %q\n              actual   %q\n", c.WantLine, line)
	if c.CheckCandidates {
		fmt.Fprintf(&b, "  candidates: expected %q\n              actual   %q  (as sets)\n", want, got)
	}
	fmt.Fprintf(&b, "  screen:\n%s\n", indent(screen, "    | "))
	t.Error(b.String())
}
