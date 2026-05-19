// Example: using gologger as a log/slog handler.
//
// gologger.DefaultLogger implements slog.Handler, so you can route the
// standard library's structured logger through gologger's CLI output
// without giving up the gologger.* fluent API. This example walks
// through the main slog integration points:
//
//   1. Drop-in handler via slog.SetDefault
//   2. Custom gologger levels usable from slog.Log(ctx, ...)
//   3. All slog attribute kinds
//   4. WithGroup flattening and WithAttrs persistence
//   5. Mixing slog.* and gologger.* in the same program
//   6. TrimGologgerLevels for clean level names in JSON/Text handlers
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

func main() {
	ctx := context.Background()

	gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	slog.SetDefault(slog.New(gologger.DefaultLogger))

	section("1. Quick start - slog routed through gologger")
	slog.Info("hello from slog")
	slog.Warn("low disk space", slog.Int("free_mb", 128))
	slog.Error("operation failed", slog.Any("err", errors.New("connection refused")))

	section("2. Custom gologger levels")
	// These map onto gologger's full level hierarchy. slog.Log lets you
	// log at any custom level - gologger picks the right label.
	slog.Log(ctx, gologger.LevelTrace, "trace: very detailed step")
	slog.Log(ctx, gologger.LevelVerbose, "verbose: noisy diagnostic")
	slog.Log(ctx, slog.LevelDebug, "debug: planner running")
	slog.Log(ctx, slog.LevelInfo, "info: standard message")
	slog.Log(ctx, gologger.LevelSilent, "silent: raw line, no [LBL]")
	slog.Log(ctx, slog.LevelWarn, "warn: something looks off")
	slog.Log(ctx, slog.LevelError, "error: request failed")

	section("3. Attribute kinds")
	slog.Info("user activity",
		slog.String("name", "john"),
		slog.Int("age", 30),
		slog.Uint64("visits", 42),
		slog.Bool("active", true),
		slog.Float64("score", 95.5),
		slog.Duration("latency", 100*time.Millisecond),
		slog.Time("login", time.Now()),
		slog.Any("custom", map[string]int{"visits": 42}),
	)

	section("4. Groups - flattened into dotted keys")
	api := slog.Default().WithGroup("api").WithGroup("request")
	api.Info("processing",
		slog.String("method", "POST"),
		slog.String("path", "/users"),
	)

	section("5. WithAttrs - attributes persist across calls")
	svc := slog.Default().With(
		slog.String("service", "user-api"),
		slog.String("version", "1.0.0"),
	)
	svc.Info("request processed", slog.String("endpoint", "/login"))
	svc.Info("request processed", slog.String("endpoint", "/logout"))

	section("6. Interop - mix slog and native gologger")
	// Both write through the same writer/formatter/level config.
	gologger.Info().Str("component", "legacy").Msg("fluent gologger call")
	slog.Info("structured slog call", slog.String("component", "new"))

	section("7. Timestamps - shared between both APIs")
	gologger.DefaultLogger.SetTimestamp(true, levels.LevelInfo)
	slog.Info("timestamped via slog")
	gologger.Info().Msg("timestamped via gologger")
	gologger.DefaultLogger.SetTimestamp(false, levels.LevelInfo)

	section("8. TrimGologgerLevels - clean labels through a JSON handler")
	// Useful when you want a JSON/Text handler (not gologger) but still
	// log at gologger's custom levels without seeing raw DEBUG-4 etc.
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, gologger.TrimGologgerLevels()))
	jsonLogger.Log(ctx, gologger.LevelTrace, "trace message")
	jsonLogger.Log(ctx, gologger.LevelVerbose, "verbose message")
	jsonLogger.Log(ctx, slog.LevelInfo, "info message")
	jsonLogger.Log(ctx, gologger.LevelSilent, "silent message (no level field)")
	jsonLogger.Log(ctx, gologger.LevelFatal, "fatal message (does not exit through this handler)")

	section("9. Fatal exits the process")
	// Logging at Fatal through the gologger slog handler calls os.Exit(1).
	// Keep this last.
	slog.Log(ctx, gologger.LevelFatal, "bye bye")
}

func section(title string) {
	gologger.Print().Msgf("\n--- %s ---", title)
}
