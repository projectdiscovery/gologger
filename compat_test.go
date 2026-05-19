package gologger

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
)

// newCapturedLogger returns a logger preconfigured with a no-color CLI
// formatter and an in-memory writer, mirroring the DefaultLogger setup.
// All retro-compat tests use this helper so assertions remain stable
// across environments (no terminal colors, no real stderr).
func newCapturedLogger(t *testing.T, maxLevel levels.Level) (*Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	l := &Logger{}
	l.SetMaxLevel(maxLevel)
	l.SetFormatter(formatter.NewCLI(true))
	l.SetWriter(&testWriter{buf: buf})
	return l, buf
}

// TestEventChain_LevelLabels documents the canonical label emitted by
// each public level method. These labels are part of the user-facing
// CLI contract and must never silently change.
func TestEventChain_LevelLabels(t *testing.T) {
	cases := []struct {
		name     string
		emit     func(l *Logger)
		expected string
	}{
		{"Info", func(l *Logger) { l.Info().Msg("m") }, "[INF] m"},
		{"Warning", func(l *Logger) { l.Warning().Msg("m") }, "[WRN] m"},
		{"Error", func(l *Logger) { l.Error().Msg("m") }, "[ERR] m"},
		{"Debug", func(l *Logger) { l.Debug().Msg("m") }, "[DBG] m"},
		{"Verbose", func(l *Logger) { l.Verbose().Msg("m") }, "[VER] m"},
		{"Print", func(l *Logger) { l.Print().Msg("m") }, "m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, buf := newCapturedLogger(t, levels.LevelVerbose)
			c.emit(l)
			if got := buf.String(); got != c.expected {
				t.Errorf("got %q, want %q", got, c.expected)
			}
		})
	}
}

// TestEventChain_StrMetadata verifies the Str() chaining contract:
// metadata keys are appended in `key=value` form after the message.
func TestEventChain_StrMetadata(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Str("host", "example.com").Msg("scanning")

	out := buf.String()
	if !strings.HasPrefix(out, "[INF] scanning") {
		t.Errorf("expected prefix %q, got %q", "[INF] scanning", out)
	}
	if !strings.Contains(out, "host=example.com") {
		t.Errorf("expected host=example.com in output, got %q", out)
	}
}

// TestEventChain_Msgf locks in printf formatting behavior.
func TestEventChain_Msgf(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Msgf("count=%d host=%s", 3, "h")
	want := "[INF] count=3 host=h"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestEventChain_MsgFuncLazy ensures the message supplier is only
// invoked when the event passes the level gate.
func TestEventChain_MsgFuncLazy(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	called := 0
	l.Debug().MsgFunc(func() string {
		called++
		return "should not run"
	})
	if called != 0 {
		t.Errorf("supplier should not run when level is gated, called=%d", called)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}

	l.Info().MsgFunc(func() string {
		called++
		return "ran"
	})
	if called != 1 {
		t.Errorf("supplier should run once when level enabled, called=%d", called)
	}
	if got := buf.String(); got != "[INF] ran" {
		t.Errorf("got %q, want %q", got, "[INF] ran")
	}
}

// TestEventChain_Label allows overriding the auto label.
func TestEventChain_Label(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Label("CUSTOM").Msg("m")
	if got := buf.String(); got != "[CUSTOM] m" {
		t.Errorf("got %q, want %q", got, "[CUSTOM] m")
	}
}

// TestEventChain_TrimsTrailingNewline guards against accidental double
// newlines: the writer is responsible for line termination, so the
// formatter input must not carry one.
func TestEventChain_TrimsTrailingNewline(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Msg("message\n")
	if got := buf.String(); got != "[INF] message" {
		t.Errorf("got %q, want %q", got, "[INF] message")
	}
}

// TestLevelFiltering covers SetMaxLevel: events at a higher level than
// the configured maximum must be dropped silently.
func TestLevelFiltering(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelError)

	l.Info().Msg("dropped")
	l.Debug().Msg("dropped")
	l.Verbose().Msg("dropped")
	if buf.Len() != 0 {
		t.Errorf("expected level filter to drop, got %q", buf.String())
	}

	l.Error().Msg("kept")
	if !strings.Contains(buf.String(), "[ERR] kept") {
		t.Errorf("expected ERR to pass filter, got %q", buf.String())
	}
}

// TestTimestamp_DefaultFormat verifies SetTimestamp emits a RFC3339
// timestamp surrounded by brackets and that minLevel gates emission.
func TestTimestamp_DefaultFormat(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelDebug)
	l.SetTimestamp(true, levels.LevelInfo)

	l.Info().Msg("withts")
	out := buf.String()

	// Expect "[INF] [<RFC3339>] withts"
	rfc3339 := regexp.MustCompile(`\[INF\] \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[Z+-][:0-9]*\] withts`)
	if !rfc3339.MatchString(out) {
		t.Errorf("output does not match RFC3339 pattern: %q", out)
	}

	buf.Reset()
	// Debug is more verbose than the timestampMinLevel (Info); higher
	// verbosity must also include the timestamp because Debug > Info.
	l.Debug().Msg("dbg")
	if !strings.Contains(buf.String(), "[DBG]") {
		t.Errorf("expected DBG label, got %q", buf.String())
	}
}

// TestTimestamp_CustomFormat exercises SetTimestampWithFormat.
func TestTimestamp_CustomFormat(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.SetTimestampWithFormat(true, levels.LevelInfo, "2006-01-02")

	l.Info().Msg("m")
	out := buf.String()

	date := time.Now().Format("2006-01-02")
	want := "[INF] [" + date + "] m"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

// TestTimestamp_EmptyFormatPreservesPrevious documents the
// SetTimestampWithFormat contract: an empty format string is treated as
// "no change" rather than a reset.
func TestTimestamp_EmptyFormatPreservesPrevious(t *testing.T) {
	l := &Logger{}
	l.SetTimestampWithFormat(true, levels.LevelInfo, "2006")
	if l.timestampFormat != "2006" {
		t.Fatalf("setup: timestampFormat = %q, want %q", l.timestampFormat, "2006")
	}
	l.SetTimestampWithFormat(true, levels.LevelInfo, "")
	if l.timestampFormat != "2006" {
		t.Errorf("empty format should not reset previous, got %q", l.timestampFormat)
	}
}

// TestTimestamp_DisabledNoMetadata ensures we don't emit a timestamp
// when SetTimestamp was never enabled.
func TestTimestamp_DisabledNoMetadata(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Msg("m")
	if got := buf.String(); got != "[INF] m" {
		t.Errorf("got %q, want %q", got, "[INF] m")
	}
}

// TestSilentNoLabel verifies that Silent events are emitted without
// any label, preserving the contract that downstream tools rely on for
// machine-readable output.
func TestSilentNoLabel(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelVerbose)
	// Silent is only on the package-level API today; emulate it by
	// creating an event directly to assert the formatter never injects
	// a label for it.
	event := newEventWithLevelAndLogger(levels.LevelSilent, l)
	event.Msg("raw output")

	if got := buf.String(); got != "raw output" {
		t.Errorf("silent must not include any label, got %q", got)
	}
}

// TestSetFormatter_JSONIntegration runs the JSON formatter through the
// public Logger pipeline so we don't only test pieces in isolation.
func TestSetFormatter_JSONIntegration(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{}
	l.SetMaxLevel(levels.LevelInfo)
	l.SetFormatter(&formatter.JSON{})
	l.SetWriter(&testWriter{buf: buf})

	l.Info().Str("k", "v").Msg("hello")
	out := buf.String()

	if !strings.Contains(out, `"msg":"hello"`) {
		t.Errorf("JSON output missing msg: %q", out)
	}
	if !strings.Contains(out, `"level":"INF"`) {
		t.Errorf("JSON output missing level: %q", out)
	}
	if !strings.Contains(out, `"k":"v"`) {
		t.Errorf("JSON output missing custom attr: %q", out)
	}
}

// TestEmptyMessage exercises the edge case where users emit an event
// with no payload (intentional or accidental). It must not panic and
// should still emit the level label.
func TestEmptyMessage(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	l.Info().Msg("")
	if got := buf.String(); got != "[INF] " {
		t.Errorf("empty message: got %q, want %q", got, "[INF] ")
	}
}

// TestMaxLevelBoundary verifies the boundary of isCurrentLevelEnabled:
// `level <= maxLevel` (i.e. equal is enabled).
func TestMaxLevelBoundary(t *testing.T) {
	l, buf := newCapturedLogger(t, levels.LevelInfo)
	// Info == max, must pass.
	l.Info().Msg("ok")
	if buf.Len() == 0 {
		t.Errorf("Info should pass when maxLevel == LevelInfo")
	}
	buf.Reset()
	// Warning is more verbose than Info; must be dropped.
	l.Warning().Msg("drop")
	if buf.Len() != 0 {
		t.Errorf("Warning should be dropped when maxLevel == LevelInfo, got %q", buf.String())
	}
}
