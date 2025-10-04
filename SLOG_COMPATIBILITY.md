# gologger slog Compatibility Guide

gologger fully supports Go's standard `log/slog` package as a drop-in replacement while preserving your existing CLI output format. This guide demonstrates how both APIs produce identical output and how to migrate incrementally.

## Table of Contents

- [Quick Start](#quick-start)
- [Side-by-Side Comparisons](#side-by-side-comparisons)
- [Attribute Types and Translation](#attribute-types-and-translation)  
- [Groups and Nested Attributes](#groups-and-nested-attributes)
- [Custom Levels](#custom-levels)
- [Migration Guide](#migration-guide)
- [Advanced Features](#advanced-features)

## Quick Start

```go
package main

import (
    "log/slog"
    "github.com/projectdiscovery/gologger"
    "github.com/projectdiscovery/gologger/levels"
)

func main() {
    // Setup gologger as slog handler (one-time setup)
    gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
    slog.SetDefault(slog.New(gologger.DefaultLogger))
    
    // Now both APIs produce identical output:
    gologger.Info().Msg("Hello from gologger")
    slog.Info("Hello from slog")
    
    // Both output: [INF] Hello from gologger
    //              [INF] Hello from slog
}
```

## Side-by-Side Comparisons

All examples below produce **identical output** when using gologger as the slog handler.

### Basic Logging

**gologger API:**
```go
gologger.Info().Msg("Service started")
gologger.Debug().Msg("Debug information") 
gologger.Warning().Msg("Something unusual")
gologger.Error().Msg("An error occurred")
```

**slog API:**
```go  
slog.Info("Service started")
slog.Debug("Debug information")
slog.Warn("Something unusual") 
slog.Error("An error occurred")
```

**Output:**
```
[INF] Service started
[DBG] Debug information
[WRN] Something unusual
[ERR] An error occurred
```

### With Attributes/Metadata

**gologger API:**
```go
gologger.Info().Str("user", "john").Str("action", "login").Msg("User logged in")
gologger.Error().Str("component", "database").Str("error", "timeout").Msg("Connection failed")
```

**slog API:**
```go
slog.Info("User logged in", slog.String("user", "john"), slog.String("action", "login"))
slog.Error("Connection failed", slog.String("component", "database"), slog.String("error", "timeout"))
```

**Output:**
```
[INF] User logged in [user=john,action=login]
[ERR] Connection failed [component=database,error=timeout]
```

### Timestamps

**gologger API:**
```go
gologger.Info().TimeStamp().Msg("Timestamped message")
// or enable automatic timestamps:
gologger.DefaultLogger.SetTimestamp(true, levels.LevelDebug)
gologger.Debug().Msg("Auto timestamped")
```

**slog API:**
```go
slog.Info("Timestamped message", slog.Time("timestamp", time.Now()))
// timestamps are automatic in slog when using standard handlers
```

**Output:**
```
[INF] Timestamped message [timestamp=2023-12-07T10:30:00Z]  
[DBG] Auto timestamped [timestamp=2023-12-07T10:30:01Z]
```

## Attribute Types and Translation

gologger automatically converts all slog attribute types to appropriate string representations:

### Supported Types

**slog API:**
```go
slog.Info("User activity",
    slog.String("name", "john"),           // String
    slog.Int("age", 30),                   // Integer  
    slog.Bool("active", true),             // Boolean
    slog.Float64("score", 95.5),           // Float
    slog.Duration("latency", 100*time.Millisecond), // Duration
    slog.Time("login", time.Now()),        // Time
    slog.Any("custom", map[string]int{"visits": 42}), // Any type
)
```

**Output:**
```
[INF] User activity [name=john,age=30,active=true,score=95.5,latency=100ms,login=2023-12-07T10:30:00Z,custom=map[visits:42]]
```

**Equivalent gologger API:**
```go
gologger.Info().
    Str("name", "john").
    Str("age", "30").
    Str("active", "true").
    Str("score", "95.5").
    Str("latency", "100ms").
    Str("login", time.Now().Format(time.RFC3339)).
    Str("custom", fmt.Sprintf("%+v", map[string]int{"visits": 42})).
    Msg("User activity")
```

## Groups and Nested Attributes

slog groups are flattened to gologger's metadata format using dotted notation:

**slog API:**
```go
// Nested groups
logger := slog.New(gologger.DefaultLogger)
apiLogger := logger.WithGroup("api")
requestLogger := apiLogger.WithGroup("request")

requestLogger.Info("Processing request",
    slog.String("method", "POST"),
    slog.String("path", "/users"),
    slog.Int("id", 123),
)
```

**Output:**
```
[INF] Processing request [api.request.method=POST,api.request.path=/users,api.request.id=123]
```

**Equivalent gologger API:**
```go
gologger.Info().
    Str("api.request.method", "POST").
    Str("api.request.path", "/users").
    Str("api.request.id", "123").
    Msg("Processing request")
```

### WithAttrs for Persistent Attributes

**slog API:**
```go
baseLogger := slog.New(gologger.DefaultLogger)
serviceLogger := baseLogger.WithAttrs([]slog.Attr{
    slog.String("service", "user-api"),
    slog.String("version", "1.0.0"),
})

serviceLogger.Info("Request processed", slog.String("endpoint", "/login"))
serviceLogger.Error("Database error", slog.String("table", "users"))
```

**Output:**
```
[INF] Request processed [service=user-api,version=1.0.0,endpoint=/login]
[ERR] Database error [service=user-api,version=1.0.0,table=users]
```

## Custom Levels

gologger provides additional levels beyond standard slog levels:

### Verbose Level (More detailed than Debug)

**gologger API:**
```go
gologger.Verbose().Msg("Very detailed information")
```

**slog API:**
```go
slog.Log(context.Background(), gologger.LevelVerbose, "Very detailed information")
```

**Output:**
```
[VER] Very detailed information
```

### Trace Level (Maps to Verbose)

Note: `gologger.LevelTrace` is provided for slog compatibility but maps internally to Verbose level.

**slog API:**
```go
slog.Log(context.Background(), gologger.LevelTrace, "Trace level information")
```

**Output:**
```
[VER] Trace level information
```

### Silent Level (No label output)

**gologger API:**
```go
gologger.Print().Msg("Clean output without labels")
gologger.Silent().Msg("Another clean message")
```

**slog API:**
```go
slog.Log(context.Background(), gologger.LevelSilent, "Clean output without labels")
```

**Output:**
```
Clean output without labels
Another clean message
```

### Fatal Level (Exits after logging)

**gologger API:**
```go
gologger.Fatal().Msg("Critical error - exiting")
```

**slog API:**  
```go
slog.Log(context.Background(), gologger.LevelFatal, "Critical error - exiting")
```

**Output:**
```
[FTL] Critical error - exiting
// Program exits with os.Exit(1)
```

## Migration Guide

### Phase 1: Setup (Zero Code Changes)

1. Configure gologger as your slog handler:

```go
import (
    "log/slog"
    "github.com/projectdiscovery/gologger" 
    "github.com/projectdiscovery/gologger/levels"
)

func init() {
    // Configure gologger
    gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
    
    // Set as default slog handler  
    slog.SetDefault(slog.New(gologger.DefaultLogger))
}
```

2. **All existing gologger code continues working unchanged**
3. **All new slog code produces identical output**

### Phase 2: Incremental Migration

Migrate function by function, file by file:

**Before:**
```go
func authenticateUser(username string) error {
    gologger.Info().Str("user", username).Msg("Authentication attempt")
    
    if err := validateCredentials(username); err != nil {
        gologger.Error().Str("user", username).Str("error", err.Error()).Msg("Authentication failed")
        return err
    }
    
    gologger.Info().Str("user", username).Msg("Authentication successful")
    return nil
}
```

**After:**
```go
func authenticateUser(username string) error {
    slog.Info("Authentication attempt", slog.String("user", username))
    
    if err := validateCredentials(username); err != nil {
        slog.Error("Authentication failed", 
            slog.String("user", username), 
            slog.String("error", err.Error()))
        return err
    }
    
    slog.Info("Authentication successful", slog.String("user", username))
    return nil
}
```

**Output is identical in both cases!**

### Phase 3: Leverage slog Ecosystem

Once migrated, you can:

- Use structured logging libraries built for slog
- Switch to different handlers (JSON, custom) while keeping same logging calls
- Use advanced features like sampling, filtering, etc.

## Advanced Features

### Clean Level Names for JSON/Text Output  

When using standard slog handlers (JSON/Text), use `TrimGologgerLevels()` to get clean level names:

```go
// Instead of "DEBUG-4", "ERROR+4" - get "TRACE", "FATAL"
handler := slog.NewJSONHandler(os.Stdout, gologger.TrimGologgerLevels())
logger := slog.New(handler)

logger.Log(context.Background(), gologger.LevelTrace, "Trace message")
logger.Log(context.Background(), gologger.LevelFatal, "Fatal message")
logger.Log(context.Background(), gologger.LevelSilent, "Silent message")
```

**JSON Output:**
```json
{"time":"2023-12-07T10:30:00Z","level":"TRACE","msg":"Trace message"}
{"time":"2023-12-07T10:30:00Z","level":"FATAL","msg":"Fatal message"}
{"time":"2023-12-07T10:30:00Z","msg":"Silent message"}
```

**Note:** Silent level messages have no level field, as intended for clean output.


### Mixed Usage Patterns

You can use both APIs in the same application:

```go
func main() {
    // Setup
    slog.SetDefault(slog.New(gologger.DefaultLogger))
    
    // Legacy functions continue using gologger  
    legacyFunction()
    
    // New functions use slog
    newFunction()
    
    // Both produce identical CLI output!
}

func legacyFunction() {
    gologger.Info().Str("component", "legacy").Msg("Legacy system active")
}

func newFunction() {
    slog.Info("New system active", slog.String("component", "new"))
}
```

### Context Integration

slog's context integration works seamlessly:

```go
ctx := context.WithValue(context.Background(), "requestID", "req-123")

// Context is available to custom handlers, middleware, etc.
slog.InfoContext(ctx, "Request processed", slog.String("endpoint", "/api/users"))
```

## Summary

- **100% Backward Compatible**: All existing gologger code continues working
- **Identical Output**: slog produces the exact same CLI format when using gologger as handler
- **Incremental Migration**: Migrate at your own pace, function by function
- **Best of Both Worlds**: Keep gologger's beautiful CLI output, gain slog's ecosystem
- **Zero Breaking Changes**: No disruption to existing applications
- **Full Feature Parity**: All gologger features (groups, attributes, custom levels) work in slog

The migration gives you access to Go's standard logging ecosystem while preserving the CLI aesthetics that make gologger special.

