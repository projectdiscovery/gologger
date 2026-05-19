package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/projectdiscovery/gologger/levels"
)

func TestTeeFormat_WritesJSONAndDelegates(t *testing.T) {
	sink := &bytes.Buffer{}
	wrapped := NewCLI(true)
	tee := NewTee(wrapped, sink)

	out, err := tee.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}

	// Output bytes come from the wrapped formatter.
	if string(out) != "[INF] hello" {
		t.Errorf("wrapped output: got %q, want %q", out, "[INF] hello")
	}

	// The sink received JSON terminated by a newline.
	sinked := sink.String()
	if !strings.HasSuffix(sinked, "\n") {
		t.Errorf("tee output is not newline terminated: %q", sinked)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(sinked)), &parsed); err != nil {
		t.Fatalf("tee sink does not contain valid JSON: %v\n%s", err, sinked)
	}
	if parsed["msg"] != "hello" {
		t.Errorf("tee JSON msg = %v, want %q", parsed["msg"], "hello")
	}
}

func TestTeeFormat_NilEventNoop(t *testing.T) {
	sink := &bytes.Buffer{}
	tee := NewTee(NewCLI(true), sink)
	out, err := tee.Format(nil)
	if err != nil {
		t.Fatalf("Format(nil) err: %v", err)
	}
	if len(out) != 0 || sink.Len() != 0 {
		t.Errorf("nil event must not emit any bytes, got out=%q sink=%q", out, sink.String())
	}
}

// TestTeeFormat_PreservesLabelForWrapped guarantees the wrapper still
// sees the label after the JSON sink consumes it (so CLI colors work).
func TestTeeFormat_PreservesLabelForWrapped(t *testing.T) {
	sink := &bytes.Buffer{}
	tee := NewTee(NewCLI(true), sink)
	out, err := tee.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if !strings.HasPrefix(string(out), "[INF] ") {
		t.Errorf("wrapped formatter lost the label, got %q", out)
	}
}
