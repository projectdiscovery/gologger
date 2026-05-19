package writer

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/projectdiscovery/gologger/levels"
)

// captureFD redirects the given *os.File pointer (Stdout or Stderr) to
// a pipe, runs fn, restores the original FD and returns the captured
// bytes. It's used to verify that the CLI writer routes silent messages
// to stdout and everything else to stderr.
func captureFD(target **os.File, fn func()) string {
	orig := *target
	r, w, _ := os.Pipe()
	*target = w

	done := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	*target = orig
	return <-done
}

func TestCLIWriter_SilentGoesToStdout(t *testing.T) {
	cli := NewCLI()
	got := captureFD(&os.Stdout, func() {
		cli.Write([]byte("silent payload"), levels.LevelSilent)
	})
	if got != "silent payload"+NewLine {
		t.Errorf("stdout = %q, want %q", got, "silent payload"+NewLine)
	}
}

func TestCLIWriter_NonSilentGoesToStderr(t *testing.T) {
	cli := NewCLI()
	got := captureFD(&os.Stderr, func() {
		cli.Write([]byte("err payload"), levels.LevelError)
	})
	if got != "err payload"+NewLine {
		t.Errorf("stderr = %q, want %q", got, "err payload"+NewLine)
	}
}

// TestCLIWriter_Concurrent exercises the internal mutex; it would race
// under `-race` without it.
func TestCLIWriter_Concurrent(t *testing.T) {
	cli := NewCLI()
	captureFD(&os.Stderr, func() {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cli.Write([]byte("x"), levels.LevelInfo)
			}()
		}
		wg.Wait()
	})
}
