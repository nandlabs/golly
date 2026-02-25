# l3 — Lightweight Levelled Logger

A fast, levelled logging package for Go with console and file writers, per-package log levels, async mode, and zero external dependencies.

---

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Logger Interface](#logger-interface)
- [Log Levels](#log-levels)
- [Configuration](#configuration)
  - [File-Based Configuration](#1-file-based-configuration)
  - [Environment Variables](#2-environment-variables)
  - [Programmatic Configuration](#3-programmatic-configuration)
- [Writers](#writers)
  - [Console Writer](#console-writer)
  - [File Writer](#file-writer)
- [Async Logging](#async-logging)
- [Thread Safety](#thread-safety)

---

## Features

- **Six log levels**: `OFF`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`
- **Console and file writers** — route levels to different destinations
- **Per-package log levels** — fine-grained control without global noise
- **Formatted logging** — `printf`-style methods (`ErrorF`, `InfoF`, etc.)
- **Async logging** — non-blocking writes via buffered channel
- **Configuration** via JSON file, environment variables, or runtime API
- **Thread-safe** — all writers are mutex-protected

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Quick Start

```go
package main

import (
    "errors"

    "oss.nandlabs.io/golly/l3"
)

// Package-level logger — call l3.Get() once per package.
var logger = l3.Get()

func main() {
    // Simple messages
    logger.Info("Server starting on port 8080")
    logger.Warn("Cache size exceeds 80% capacity")
    logger.Error("Failed to connect to database", errors.New("connection refused"))

    // Formatted messages (printf-style)
    logger.InfoF("Listening on %s:%d", "localhost", 8080)
    logger.DebugF("Request processed in %dms", 42)
    logger.ErrorF("HTTP %d: %s", 500, "internal server error")

    // Level checks (avoid expensive arg computation)
    logger.Trace("Detailed trace data: ", someExpensiveCall())
}

func someExpensiveCall() string {
    return "trace-payload"
}
```

**Default output** (text format, `INFO` level, to stdout/stderr):

```
2026-02-21T10:00:00+11:00 INFO Server starting on port 8080
2026-02-21T10:00:00+11:00 WARN Cache size exceeds 80% capacity
2026-02-21T10:00:00+11:00 ERROR Failed to connect to database connection refused
2026-02-21T10:00:00+11:00 INFO Listening on localhost:8080
```

## Logger Interface

```go
type Logger interface {
    Error(a ...interface{})
    ErrorF(f string, a ...interface{})
    Warn(a ...interface{})
    WarnF(f string, a ...interface{})
    Info(a ...interface{})
    InfoF(f string, a ...interface{})
    Debug(a ...interface{})
    DebugF(f string, a ...interface{})
    Trace(a ...interface{})
    TraceF(f string, a ...interface{})
}
```

Get a logger instance with `l3.Get()`. Each call returns a logger scoped to the calling package — log levels can be configured per package.

## Log Levels

Levels are ordered by severity. Setting a level enables that level and all levels above it.

| Level   | Value | Includes                        |
| ------- | ----- | ------------------------------- |
| `OFF`   | 0     | Nothing                         |
| `ERROR` | 1     | Error                           |
| `WARN`  | 2     | Error, Warn                     |
| `INFO`  | 3     | Error, Warn, Info               |
| `DEBUG` | 4     | Error, Warn, Info, Debug        |
| `TRACE` | 5     | Error, Warn, Info, Debug, Trace |

## Configuration

### 1. File-Based Configuration

Place a `log-config.json` file in the application directory, or set `GC_LOG_CONFIG_FILE` to a custom path.

```json
{
  "format": "text",
  "async": false,
  "defaultLvl": "INFO",
  "datePattern": "2006-01-02T15:04:05Z07:00",
  "includeFunction": true,
  "includeLineNum": true,
  "pkgConfigs": [
    {
      "pkgName": "main",
      "level": "DEBUG"
    },
    {
      "pkgName": "server",
      "level": "WARN"
    }
  ],
  "writers": [
    {
      "console": {
        "errToStdOut": false,
        "warnToStdOut": false
      }
    },
    {
      "file": {
        "defaultPath": "/var/log/app/app.log",
        "errorPath": "/var/log/app/error.log"
      }
    }
  ]
}
```

**Configuration fields:**

| Field             | Type    | Description                                                                                           | Default        |
| ----------------- | ------- | ----------------------------------------------------------------------------------------------------- | -------------- |
| `format`          | String  | Output format: `"text"` or `"json"`                                                                   | `"text"`       |
| `async`           | Boolean | Write log messages asynchronously via a buffered channel                                              | `false`        |
| `queueSize`       | Integer | Channel buffer size when `async` is `true`                                                            | `4096`         |
| `defaultLvl`      | String  | Global default log level: `OFF`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`                            | `"INFO"`       |
| `datePattern`     | String  | Timestamp format (Go `time.Format` layout)                                                            | `time.RFC3339` |
| `includeFunction` | Boolean | Include the calling function name in log entries                                                      | `false`        |
| `includeLineNum`  | Boolean | Include line number (only works when `includeFunction` is `true`)                                     | `false`        |
| `pkgConfigs`      | Array   | Per-package level overrides: `[{"pkgName": "pkg", "level": "DEBUG"}]`                                 | `null`         |
| `writers`         | Array   | Array of writer configs — each entry is either a `console` or `file` writer (see [Writers](#writers)) | Console only   |

### 2. Environment Variables

When no config file is found, the framework loads a default console configuration. These environment variables override the defaults:

| Variable             | Type    | Description                               | Default             |
| -------------------- | ------- | ----------------------------------------- | ------------------- |
| `GC_LOG_CONFIG_FILE` | String  | Path to the JSON config file              | `./log-config.json` |
| `GC_LOG_ASYNC`       | Boolean | Enable async logging                      | `false`             |
| `GC_LOG_FMT`         | String  | Output format: `"text"` or `"json"`       | `"text"`            |
| `GC_LOG_DEF_LEVEL`   | String  | Default log level                         | `"INFO"`            |
| `GC_LOG_TIME_FMT`    | String  | Timestamp format                          | `time.RFC3339`      |
| `GC_LOG_ERR_STDOUT`  | Boolean | Write `ERROR` to stdout instead of stderr | `false`             |
| `GC_LOG_WARN_STDOUT` | Boolean | Write `WARN` to stdout instead of stderr  | `false`             |

### 3. Programmatic Configuration

Use `l3.Configure()` to set the log configuration at runtime:

```go
package main

import "oss.nandlabs.io/golly/l3"

func main() {
    l3.Configure(&l3.LogConfig{
        Format:          "json",
        Async:           true,
        QueueSize:       8192,
        DefaultLvl:      "DEBUG",
        IncludeFunction: true,
        IncludeLineNum:  true,
        PkgConfigs: []*l3.PackageConfig{
            {PackageName: "main", Level: "TRACE"},
            {PackageName: "db", Level: "WARN"},
        },
        Writers: []*l3.WriterConfig{
            {Console: &l3.ConsoleConfig{
                WriteErrToStdOut:  false,
                WriteWarnToStdOut: false,
            }},
        },
    })

    logger := l3.Get()
    logger.InfoF("Logging configured at %s level", "DEBUG")
}
```

## Writers

### Console Writer

Routes log messages to `os.Stdout` / `os.Stderr`:

- `INFO`, `DEBUG`, `TRACE` → `os.Stdout`
- `ERROR`, `WARN` → `os.Stderr` (configurable via `errToStdOut` / `warnToStdOut`)

```json
{
  "console": {
    "errToStdOut": false,
    "warnToStdOut": false
  }
}
```

### File Writer

Routes log messages to files. Each level can have its own file, or share a `defaultPath`:

```json
{
  "file": {
    "defaultPath": "/var/log/app/app.log",
    "errorPath": "/var/log/app/error.log",
    "warnPath": "/var/log/app/warn.log",
    "infoPath": "/var/log/app/info.log",
    "debugPath": "/var/log/app/debug.log",
    "tracePath": "/var/log/app/trace.log"
  }
}
```

If a level-specific path is omitted, that level falls back to `defaultPath`. Files are opened with `O_RDWR|O_APPEND|O_CREATE`.

## Async Logging

When `async` is `true`, log messages are dispatched to a buffered channel and written by a background goroutine. This avoids blocking the calling goroutine on I/O:

```go
l3.Configure(&l3.LogConfig{
    Async:      true,
    QueueSize:  4096, // channel buffer size
    DefaultLvl: "INFO",
    Writers: []*l3.WriterConfig{
        {Console: &l3.ConsoleConfig{}},
    },
})
```

> **Note**: If the channel is full, `handleLog` will block until space is available.

## Thread Safety

All writers (`ConsoleWriter`, `FileWriter`) are protected by `sync.Mutex`. The global log configuration is accessed under a package-level mutex. It is safe to log from multiple goroutines concurrently.
