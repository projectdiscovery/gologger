package levels

import "testing"

// TestLevelString locks in the textual representation of every level.
// Downstream tools (and the JSON formatter) rely on these strings.
func TestLevelString(t *testing.T) {
	cases := []struct {
		level Level
		want  string
	}{
		{LevelFatal, "fatal"},
		{LevelSilent, "silent"},
		{LevelError, "error"},
		{LevelInfo, "info"},
		{LevelWarning, "warning"},
		{LevelDebug, "debug"},
		{LevelVerbose, "verbose"},
	}
	for _, c := range cases {
		if got := c.level.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.level, got, c.want)
		}
	}
}

// TestLevelOrdering documents the integer ordering used throughout the
// library (callers compare levels with `<=` for filtering). Reshuffling
// the constants without updating consumers would silently break gating.
func TestLevelOrdering(t *testing.T) {
	want := []Level{LevelFatal, LevelSilent, LevelError, LevelInfo, LevelWarning, LevelDebug, LevelVerbose}
	for i, l := range want {
		if int(l) != i {
			t.Errorf("Level constant %v at iota position %d has value %d", l, i, int(l))
		}
	}
}
