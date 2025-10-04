package gologger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
)

// testWriter captures output for testing
type testWriter struct {
	buf *bytes.Buffer
}

func (tw *testWriter) Write(data []byte, level levels.Level) {
	tw.buf.Write(data)
}

func TestSlogHandlerInterface(t *testing.T) {
	// Test that Logger implements slog.Handler
	var _ slog.Handler = (*Logger)(nil)
}

func TestSlogCompatibility(t *testing.T) {
	// Create a test logger with captured output
	buf := &bytes.Buffer{}
	testLogger := &Logger{}
	testLogger.SetMaxLevel(levels.LevelDebug)
	testLogger.SetFormatter(formatter.NewCLI(false))
	testLogger.SetWriter(&testWriter{buf: buf})

	// Create slog logger using gologger as handler
	slogLogger := slog.New(testLogger)

	// Test slog logging
	slogLogger.Info("slog info message")
	slogLogger.Debug("slog debug message")
	slogLogger.Warn("slog warn message")
	slogLogger.Error("slog error message")

	output := buf.String()
	
	// Verify output contains expected messages
	if output == "" {
		t.Fatal("Expected output from slog logging, got empty string")
	}
	
	// Reset buffer for backward compatibility test
	buf.Reset()

	// Test backward compatibility - existing gologger API should still work
	testLogger.Info().Msg("gologger info message")
	testLogger.Debug().Msg("gologger debug message")
	testLogger.Warning().Msg("gologger warning message") 
	testLogger.Error().Msg("gologger error message")

	backwardOutput := buf.String()
	
	if backwardOutput == "" {
		t.Fatal("Expected output from gologger logging, got empty string")
	}
}

func TestLevelMapping(t *testing.T) {
	tests := []struct {
		slogLevel     slog.Level
		expectedLevel levels.Level
	}{
		{slog.LevelDebug, levels.LevelDebug},
		{slog.LevelInfo, levels.LevelInfo},
		{slog.LevelWarn, levels.LevelWarning},
		{slog.LevelError, levels.LevelError},
	}

	for _, test := range tests {
		result := slogLevelToGologgerLevel(test.slogLevel)
		if result != test.expectedLevel {
			t.Errorf("slogLevelToGologgerLevel(%v) = %v, want %v", 
				test.slogLevel, result, test.expectedLevel)
		}
	}
}

func TestSlogHandlerMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelDebug)
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	ctx := context.Background()

	// Test Enabled
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Expected Info level to be enabled")
	}

	// Test Handle
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	record.AddAttrs(slog.String("key", "value"))
	
	err := logger.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected output from Handle method")
	}

	// Test WithAttrs and WithGroup return new handler instances
	newHandler := logger.WithAttrs([]slog.Attr{slog.String("test", "attr")})
	if newHandler == logger {
		t.Error("WithAttrs should return a new handler instance")
	}

	groupHandler := logger.WithGroup("testgroup")
	if groupHandler == logger {
		t.Error("WithGroup should return a new handler instance")
	}
}


func TestOutputIdentity(t *testing.T) {
	// Test that gologger.Info().Msg() and slog.Info() produce identical output
	
	// Capture gologger output
	gologgerBuf := &bytes.Buffer{}
	gologgerLogger := &Logger{}
	gologgerLogger.SetMaxLevel(levels.LevelDebug)
	gologgerLogger.SetFormatter(formatter.NewCLI(false)) // no colors for comparison
	gologgerLogger.SetWriter(&testWriter{buf: gologgerBuf})
	
	// Capture slog output using same gologger as handler  
	slogBuf := &bytes.Buffer{}
	slogLogger := &Logger{}
	slogLogger.SetMaxLevel(levels.LevelDebug)
	slogLogger.SetFormatter(formatter.NewCLI(false)) // no colors for comparison
	slogLogger.SetWriter(&testWriter{buf: slogBuf})
	
	slogInstance := slog.New(slogLogger)
	
	// Test identical messages
	gologgerLogger.Info().Msg("test message")
	slogInstance.Info("test message")
	
	gologgerOutput := gologgerBuf.String()
	slogOutput := slogBuf.String()
	
	if gologgerOutput != slogOutput {
		t.Errorf("Output mismatch:\nGologger: %q\nSlog: %q", gologgerOutput, slogOutput)
	}
	
	// Reset buffers
	gologgerBuf.Reset()
	slogBuf.Reset()
	
	// Test with metadata/attributes  
	gologgerLogger.Info().Str("key", "value").Msg("test with metadata")
	slogInstance.Info("test with metadata", slog.String("key", "value"))
	
	gologgerOutput = gologgerBuf.String()
	slogOutput = slogBuf.String()
	
	if gologgerOutput != slogOutput {
		t.Errorf("Output with metadata mismatch:\nGologger: %q\nSlog: %q", gologgerOutput, slogOutput)
	}
}

func TestCustomSlogLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelVerbose) // Enable all custom levels
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	slogLogger := slog.New(logger)
	ctx := context.Background()

	tests := []struct {
		name        string
		level       slog.Level
		expectedMsg string
	}{
		{"LevelTrace", LevelTrace, "trace message"},
		{"LevelVerbose", LevelVerbose, "verbose message"},
		{"LevelSilent", LevelSilent, "silent message"},
		// Skip Fatal level test as it would exit the process
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			slogLogger.Log(ctx, test.level, test.expectedMsg)
			
			output := buf.String()
			if output == "" {
				t.Errorf("Expected output for %s level, got empty string", test.name)
			}
			
			// Verify message content
			if !strings.Contains(output, test.expectedMsg) {
				t.Errorf("Expected output to contain %q, got %q", test.expectedMsg, output)
			}
		})
	}
}

func TestSlogGroups(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelInfo)
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	// Test simple group
	groupLogger := logger.WithGroup("api")
	slogLogger := slog.New(groupLogger)
	
	buf.Reset()
	slogLogger.Info("test message", slog.String("method", "POST"))
	output := buf.String()
	
	// Should contain grouped attribute (allowing for ANSI codes)
	if !strings.Contains(output, "api.method") || !strings.Contains(output, "POST") {
		t.Errorf("Expected grouped attribute 'api.method=POST', got: %q", output)
	}

	// Test nested groups
	nestedLogger := groupLogger.WithGroup("request")
	nestedSlogLogger := slog.New(nestedLogger)
	
	buf.Reset()
	nestedSlogLogger.Info("nested test", slog.String("path", "/users"))
	output = buf.String()
	
	// Should contain nested grouped attribute (allowing for ANSI codes)
	if !strings.Contains(output, "api.request.path") || !strings.Contains(output, "/users") {
		t.Errorf("Expected nested grouped attribute 'api.request.path=/users', got: %q", output)
	}
}

func TestSlogWithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelInfo)
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	// Test WithAttrs
	persistedLogger := logger.WithAttrs([]slog.Attr{
		slog.String("service", "api"),
		slog.String("version", "1.0"),
	})
	slogLogger := slog.New(persistedLogger)
	
	buf.Reset()
	slogLogger.Info("test message", slog.String("endpoint", "/login"))
	output := buf.String()
	
	// Should contain persisted and current attributes (allowing for ANSI codes)
	if !strings.Contains(output, "service") || !strings.Contains(output, "api") {
		t.Errorf("Expected persisted attribute 'service=api', got: %q", output)
	}
	if !strings.Contains(output, "version") || !strings.Contains(output, "1.0") {
		t.Errorf("Expected persisted attribute 'version=1.0', got: %q", output)
	}
	if !strings.Contains(output, "endpoint") || !strings.Contains(output, "/login") {
		t.Errorf("Expected current attribute 'endpoint=/login', got: %q", output)
	}
}

func TestAttributeTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelInfo)
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	slogLogger := slog.New(logger)
	
	buf.Reset()
	slogLogger.Info("attribute test",
		slog.String("str", "value"),
		slog.Int("int", 42),
		slog.Bool("bool", true),
		slog.Float64("float", 3.14),
		slog.Duration("dur", 100*time.Millisecond),
		slog.Time("time", time.Date(2023, 12, 7, 10, 30, 0, 0, time.UTC)),
		slog.Any("any", map[string]int{"count": 5}),
	)
	
	output := buf.String()
	
	// Verify all attribute types are formatted correctly (allowing for ANSI codes)
	expectedValues := []string{
		"value",
		"42", 
		"true",
		"3.14",
		"100ms",
		"2023-12-07T10:30:00Z",
		"map[count:5]",
	}
	expectedKeys := []string{
		"str",
		"int",
		"bool", 
		"float",
		"dur",
		"time",
		"any",
	}
	
	for i, key := range expectedKeys {
		if !strings.Contains(output, key) || !strings.Contains(output, expectedValues[i]) {
			t.Errorf("Expected output to contain key %q and value %q, got: %q", key, expectedValues[i], output)
		}
	}
}

func TestFatalLevelOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelFatal)
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	// Test that we can capture Fatal level output before exit
	// We can't test the actual exit behavior in a unit test easily,
	// but we can verify the message gets logged
	
	// Create a subprocess test for actual exit behavior would be better
	// but for now just test that the message is formatted correctly
	event := &Event{
		logger:   logger,
		level:    levels.LevelFatal,
		message:  "fatal error",
		metadata: make(map[string]string),
	}
	event.setLevelMetadata(levels.LevelFatal)
	
	// Manually call formatter and writer (bypassing the Log method that would exit)
	data, err := logger.formatter.Format(&formatter.LogEvent{
		Message:  event.message,
		Level:    event.level,
		Metadata: event.metadata,
	})
	if err != nil {
		t.Fatalf("Formatter error: %v", err)
	}
	logger.writer.Write(data, event.level)
	
	output := buf.String()
	if !strings.Contains(output, "fatal error") {
		t.Errorf("Expected fatal message in output, got: %q", output)
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("EmptyGroupName", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := &Logger{}
		logger.SetMaxLevel(levels.LevelInfo)
		logger.SetFormatter(formatter.NewCLI(false))
		logger.SetWriter(&testWriter{buf: buf})

		// Test empty group name (should still work but create awkward keys)
		emptyGroupLogger := logger.WithGroup("")
		slogLogger := slog.New(emptyGroupLogger)
		
		buf.Reset()
		slogLogger.Info("test", slog.String("key", "value"))
		output := buf.String()
		
		// Should contain the attribute with empty group prefix
		if !strings.Contains(output, ".key") || !strings.Contains(output, "value") {
			t.Errorf("Expected empty group to create '.key=value', got: %q", output)
		}
	})

	t.Run("DeepGroupNesting", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := &Logger{}
		logger.SetMaxLevel(levels.LevelInfo)
		logger.SetFormatter(formatter.NewCLI(false))
		logger.SetWriter(&testWriter{buf: buf})

		// Test deep group nesting
		deepLogger := logger.WithGroup("level1").WithGroup("level2").WithGroup("level3")
		slogLogger := slog.New(deepLogger)
		
		buf.Reset()
		slogLogger.Info("nested message", slog.String("key", "deep"))
		output := buf.String()
		
		// Should contain deeply nested attribute key
		if !strings.Contains(output, "level1.level2.level3.key") || !strings.Contains(output, "deep") {
			t.Errorf("Expected deeply nested key 'level1.level2.level3.key=deep', got: %q", output)
		}
	})

	t.Run("MixedGroupsAndAttrs", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := &Logger{}
		logger.SetMaxLevel(levels.LevelInfo)
		logger.SetFormatter(formatter.NewCLI(false))
		logger.SetWriter(&testWriter{buf: buf})

		// Test mixing groups and persistent attributes
		baseLogger := logger.WithAttrs([]slog.Attr{slog.String("base", "value")})
		groupLogger := baseLogger.WithGroup("api")
		finalLogger := groupLogger.WithAttrs([]slog.Attr{slog.String("service", "auth")})
		
		slogLogger := slog.New(finalLogger)
		
		buf.Reset()
		slogLogger.Info("mixed test", slog.String("request", "login"))
		output := buf.String()
		
		// Should contain all attributes: base (no group), service (grouped), request (grouped)
		if !strings.Contains(output, "base") || !strings.Contains(output, "value") {
			t.Errorf("Expected base attribute, got: %q", output)
		}
		if !strings.Contains(output, "api.service") || !strings.Contains(output, "auth") {
			t.Errorf("Expected grouped service attribute, got: %q", output)
		}
		if !strings.Contains(output, "api.request") || !strings.Contains(output, "login") {
			t.Errorf("Expected grouped request attribute, got: %q", output)
		}
	})
}

func TestConcurrentUsage(t *testing.T) {
	// Test that concurrent usage of logger doesn't cause data races or corruption
	buf := &bytes.Buffer{}
	logger := &Logger{}
	logger.SetMaxLevel(levels.LevelInfo) 
	logger.SetFormatter(formatter.NewCLI(false))
	logger.SetWriter(&testWriter{buf: buf})

	slogLogger := slog.New(logger)
	
	// Run multiple goroutines logging concurrently
	const numGoroutines = 10
	const messagesPerGoroutine = 10
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				slogLogger.Info("concurrent message", 
					slog.Int("goroutine", id), 
					slog.Int("message", j))
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	output := buf.String()
	
	// Should contain messages from all goroutines (exact count may vary due to concurrency)
	if len(output) == 0 {
		t.Error("Expected output from concurrent logging, got empty string")
	}
	
	// Check that we have messages from different goroutines
	foundGoroutines := make(map[int]bool)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for i := 0; i < numGoroutines; i++ {
			if strings.Contains(line, "goroutine") && strings.Contains(line, fmt.Sprintf("%d", i)) {
				foundGoroutines[i] = true
			}
		}
	}
	
	if len(foundGoroutines) < numGoroutines/2 {
		t.Errorf("Expected messages from at least %d goroutines, found %d", numGoroutines/2, len(foundGoroutines))
	}
}