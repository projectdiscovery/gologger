package gologger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
)

// dualLoggers holds both gologger and slog logger instances that write to separate buffers
type dualLoggersClean struct {
	gologgerBuf *bytes.Buffer
	slogBuf     *bytes.Buffer
	gologger    *Logger
	slogLogger  *slog.Logger
}

// setupDualLoggersClean creates a pair of loggers (gologger and slog with gologger handler)
// that write to separate buffers for output comparison
func setupDualLoggersClean() *dualLoggersClean {
	// Setup gologger instance
	gologgerBuf := &bytes.Buffer{}
	gologger := &Logger{}
	gologger.SetMaxLevel(levels.LevelVerbose) // Enable all levels
	gologger.SetFormatter(formatter.NewCLI(false)) // No colors for comparison
	gologger.SetWriter(&testWriter{buf: gologgerBuf})
	
	// Setup slog instance using gologger handler
	slogBuf := &bytes.Buffer{}
	slogHandler := &Logger{}
	slogHandler.SetMaxLevel(levels.LevelVerbose) // Enable all levels
	slogHandler.SetFormatter(formatter.NewCLI(false)) // No colors for comparison
	slogHandler.SetWriter(&testWriter{buf: slogBuf})
	slogLogger := slog.New(slogHandler)
	
	return &dualLoggersClean{
		gologgerBuf: gologgerBuf,
		slogBuf:     slogBuf,
		gologger:    gologger,
		slogLogger:  slogLogger,
	}
}

// assertOutputsEqualClean compares the outputs from both loggers and fails test if different
func (dl *dualLoggersClean) assertOutputsEqualClean(t *testing.T, testName string) {
	t.Helper()
	
	gologgerOutput := strings.TrimSpace(dl.gologgerBuf.String())
	slogOutput := strings.TrimSpace(dl.slogBuf.String())
	
	if gologgerOutput != slogOutput {
		t.Errorf("%s output mismatch:\nGologger: %q\nSlog:     %q", testName, gologgerOutput, slogOutput)
	}
}

// reset clears both buffers for next test
func (dl *dualLoggersClean) reset() {
	dl.gologgerBuf.Reset()
	dl.slogBuf.Reset()
}

// TestPackageLevelFunctionsClean tests equivalence of package-level gologger functions vs slog
func TestPackageLevelFunctionsClean(t *testing.T) {
	// Setup package-level gologger to use our test buffer
	originalLogger := DefaultLogger
	defer func() { DefaultLogger = originalLogger }() // restore after test
	
	dl := setupDualLoggersClean()
	DefaultLogger = dl.gologger // Use our test logger as default
	
	tests := []struct {
		name           string
		gologgerFunc   func()
		slogFunc       func()
	}{
		{
			name: "Info",
			gologgerFunc: func() { Info().Msg("test info message") },
			slogFunc:     func() { dl.slogLogger.Info("test info message") },
		},
		{
			name: "Warning", 
			gologgerFunc: func() { Warning().Msg("test warning message") },
			slogFunc:     func() { dl.slogLogger.Warn("test warning message") },
		},
		{
			name: "Error",
			gologgerFunc: func() { Error().Msg("test error message") },
			slogFunc:     func() { dl.slogLogger.Error("test error message") },
		},
		{
			name: "Debug",
			gologgerFunc: func() { Debug().Msg("test debug message") },
			slogFunc:     func() { dl.slogLogger.Debug("test debug message") },
		},
		{
			name: "Verbose", 
			gologgerFunc: func() { Verbose().Msg("test verbose message") },
			slogFunc:     func() { dl.slogLogger.LogAttrs(context.Background(), LevelVerbose, "test verbose message") },
		},
		{
			name: "Print",
			gologgerFunc: func() { Print().Msg("test print message") },
			slogFunc:     func() { dl.slogLogger.LogAttrs(context.Background(), LevelSilent, "test print message") },
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl.reset()
			
			// Execute both logging approaches
			test.gologgerFunc()
			test.slogFunc()
			
			// Compare outputs
			dl.assertOutputsEqualClean(t, test.name)
		})
	}
}

// TestPackageLevelFunctionsWithMetadataClean tests package functions with metadata/attributes
func TestPackageLevelFunctionsWithMetadataClean(t *testing.T) {
	// Setup package-level gologger to use our test buffer
	originalLogger := DefaultLogger
	defer func() { DefaultLogger = originalLogger }()
	
	dl := setupDualLoggersClean()
	DefaultLogger = dl.gologger
	
	tests := []struct {
		name           string
		gologgerFunc   func()
		slogFunc       func()
	}{
		{
			name: "Info with single attribute",
			gologgerFunc: func() { Info().Str("key", "value").Msg("test message") },
			slogFunc:     func() { dl.slogLogger.Info("test message", slog.String("key", "value")) },
		},
		{
			name: "Error with single attribute",
			gologgerFunc: func() {
				Error().Str("component", "auth").Msg("authentication failed")
			},
			slogFunc: func() {
				dl.slogLogger.Error("authentication failed",
					slog.String("component", "auth"))
			},
		},
		{
			name: "Debug with single attribute",
			gologgerFunc: func() { Debug().Str("service", "api").Msg("debug info") },
			slogFunc:     func() { dl.slogLogger.Debug("debug info", slog.String("service", "api")) },
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dl.reset()
			
			test.gologgerFunc()
			test.slogFunc()
			
			dl.assertOutputsEqualClean(t, test.name)
		})
	}
}