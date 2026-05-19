package formatter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/projectdiscovery/gologger/levels"
)

func TestJSONFormat_BasicShape(t *testing.T) {
	f := &JSON{}
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF", "user": "john"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}

	if got["msg"] != "hello" {
		t.Errorf("msg = %v, want %q", got["msg"], "hello")
	}
	if got["level"] != "INF" {
		t.Errorf("level = %v, want %q", got["level"], "INF")
	}
	if got["user"] != "john" {
		t.Errorf("user = %v, want %q", got["user"], "john")
	}
	if _, ok := got["timestamp"]; !ok {
		t.Error("output is missing the timestamp key")
	}
}

func TestJSONFormat_EmptyLabelOmitsLevel(t *testing.T) {
	f := &JSON{}
	out, err := f.Format(&LogEvent{
		Message:  "hello",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": ""},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	if strings.Contains(string(out), `"level":`) {
		t.Errorf("empty label should not emit a level field, got %s", out)
	}
}

func TestJSONFormat_SortedKeys(t *testing.T) {
	// SortMapKeys is enabled in the package init; capture the contract.
	f := &JSON{}
	out, err := f.Format(&LogEvent{
		Message:  "m",
		Level:    levels.LevelInfo,
		Metadata: map[string]string{"label": "INF", "bbb": "2", "aaa": "1"},
	})
	if err != nil {
		t.Fatalf("Format err: %v", err)
	}
	// Field "aaa" must appear before "bbb" in the serialized output.
	a := strings.Index(string(out), `"aaa"`)
	b := strings.Index(string(out), `"bbb"`)
	if a < 0 || b < 0 || a > b {
		t.Errorf("expected sorted keys (aaa before bbb), got %s", out)
	}
}
