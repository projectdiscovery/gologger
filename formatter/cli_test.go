package formatter

import (
	"strings"
	"testing"

	"github.com/projectdiscovery/gologger/levels"
)

func TestCLIFormat_PlainMessage(t *testing.T) {
	f := NewCLI(true) // no colors
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if string(out) != "hello" {
		t.Errorf("plain message: got %q, want %q", out, "hello")
	}
}

func TestCLIFormat_WithLabel(t *testing.T) {
	f := NewCLI(true)
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if string(out) != "[INF] hello" {
		t.Errorf("with label: got %q, want %q", out, "[INF] hello")
	}
}

func TestCLIFormat_WithTimestamp(t *testing.T) {
	f := NewCLI(true)
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF", "timestamp": "2024-01-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	want := "[INF] [2024-01-01T00:00:00Z] hello"
	if string(out) != want {
		t.Errorf("with timestamp: got %q, want %q", out, want)
	}
}

func TestCLIFormat_WithMetadata(t *testing.T) {
	f := NewCLI(true)
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF", "user": "john"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if !strings.HasPrefix(string(out), "[INF] hello") {
		t.Errorf("metadata format: prefix mismatch, got %q", out)
	}
	if !strings.Contains(string(out), "user=john") {
		t.Errorf("metadata format: missing user=john, got %q", out)
	}
}

func TestCLIFormat_EmptyLabelOmitted(t *testing.T) {
	f := NewCLI(true)
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": ""},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if strings.Contains(string(out), "[]") {
		t.Errorf("empty label should be omitted, got %q", out)
	}
}

func TestCLIFormat_ColorsAddANSI(t *testing.T) {
	f := NewCLI(false) // colors enabled
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if !strings.Contains(string(out), "\x1b[") {
		t.Errorf("colors enabled: expected ANSI escape in output, got %q", out)
	}
}

func TestCLIFormat_SilentSkipsColor(t *testing.T) {
	f := NewCLI(false)
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelSilent,
		Metadata: map[string]string{"label": "X"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	// Silent must keep the raw label without ANSI wrapping.
	if !strings.Contains(string(out), "[X]") {
		t.Errorf("silent: expected raw [X] label, got %q", out)
	}
	if strings.Contains(string(out), "\x1b[") {
		t.Errorf("silent: should not contain ANSI escapes, got %q", out)
	}
}

// TestCLIFormat_RemovesLabelKeyFromMetadata verifies the side-effect
// that Format() consumes the label/timestamp keys (callers expect to
// see the remaining metadata only after a Format call).
func TestCLIFormat_RemovesLabelKeyFromMetadata(t *testing.T) {
	f := NewCLI(true)
	meta := map[string]string{"label": "INF", "timestamp": "t", "k": "v"}
	if _, err := f.Format(&LogEvent{Message: "m", Level: levels.LevelInfo, Metadata: meta}); err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if _, ok := meta["label"]; ok {
		t.Error("Format must delete the 'label' key from metadata")
	}
	if _, ok := meta["timestamp"]; ok {
		t.Error("Format must delete the 'timestamp' key from metadata")
	}
	if meta["k"] != "v" {
		t.Error("Format must preserve unrelated keys")
	}
}
