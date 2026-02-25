# Chrono

A full-featured task scheduler for Go applications. Supports cron-based scheduling, fixed-interval execution, and one-shot delayed tasks with comprehensive job management.

## Features

- **Cron Scheduling** — Standard 5-field cron expressions with wildcards, ranges, lists, steps, and predefined macros
- **Interval Scheduling** — Run tasks at fixed intervals (e.g., every 30 seconds)
- **One-Shot Tasks** — Execute a task once after a specified delay
- **Hybrid Event-Driven Architecture** — Precise timer wakes exactly when the next job is due, combined with a background poll to detect changes from other instances
- **Pluggable Storage** — `Storage` interface for job state persistence and distributed locking
- **Built-in Storage** — In-memory and file-based (YAML/JSON/XML) storage implementations included
- **Cluster Support** — Run multiple chrono instances with shared storage for high availability
- **Job Management** — Add, remove, pause, resume, and inspect jobs at runtime
- **Timeout Support** — Set maximum execution time per job
- **Retry Support** — Automatically retry failed jobs with configurable retry count
- **Callbacks** — Register success and error callbacks per job
- **Overlap Prevention** — Prevents concurrent execution of the same job (local + distributed locks)
- **Thread-Safe** — Safe for concurrent use from multiple goroutines
- **Graceful Shutdown** — Waits for running jobs to complete before stopping

## Installation

```bash
go get oss.nandlabs.io/golly
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    "oss.nandlabs.io/golly/chrono"
)

func main() {
    // Create a new scheduler
    s := chrono.New()

    // Add a cron job — runs every 5 minutes
    s.AddCronJob("cleanup", "Temp Cleanup", func(ctx context.Context) error {
        fmt.Println("Cleaning up temporary files...")
        return nil
    }, "*/5 * * * *")

    // Add an interval job — runs every 30 seconds
    s.AddIntervalJob("heartbeat", "Heartbeat", func(ctx context.Context) error {
        fmt.Println("Sending heartbeat...")
        return nil
    }, 30*time.Second)

    // Add a one-shot job — runs once after 5 seconds
    s.AddOneShotJob("init", "Initialize Cache", func(ctx context.Context) error {
        fmt.Println("Initializing cache...")
        return nil
    }, 5*time.Second)

    // Start the scheduler
    s.Start()

    // ... application runs ...

    // Stop gracefully
    s.Stop()
}
```

## Architecture

Chrono uses a **hybrid event-driven** execution model:

1. **Precise Timer** — A `time.Timer` that sleeps until exactly the next job is due, providing near-zero latency for locally-registered jobs.
2. **Background Poll** — A `time.Ticker` that periodically polls the storage backend to discover changes made by other scheduler instances (e.g., new jobs, removed jobs, resumed jobs).
3. **Wake Signal** — Mutations (AddJob, RemoveJob, ResumeJob) on the local instance immediately signal the run loop to recalculate the timer, so newly added jobs are picked up instantly without waiting for the next poll cycle.

This hybrid approach delivers efficient CPU usage (no unnecessary polling when idle) while maintaining cluster-level coordination through shared storage.

## Cron Expression Format

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6, 0 = Sunday)
│ │ │ │ │
* * * * *
```

### Supported Syntax

| Symbol  | Description       | Example              |
| ------- | ----------------- | -------------------- |
| `*`     | All values        | `* * * * *`          |
| `*/n`   | Every nth value   | `*/5 * * * *`        |
| `n`     | Specific value    | `30 * * * *`         |
| `n-m`   | Range (inclusive) | `0 9-17 * * *`       |
| `n-m/s` | Range with step   | `0-30/10 * * *`      |
| `n,m,o` | List of values    | `0,15,30,45 * * * *` |

### Predefined Macros

| Macro      | Equivalent  | Description           |
| ---------- | ----------- | --------------------- |
| `@yearly`  | `0 0 1 1 *` | Once a year (Jan 1)   |
| `@monthly` | `0 0 1 * *` | Once a month (1st)    |
| `@weekly`  | `0 0 * * 0` | Once a week (Sunday)  |
| `@daily`   | `0 0 * * *` | Once a day (midnight) |
| `@hourly`  | `0 * * * *` | Once an hour          |

## Job Options

```go
// Set a timeout for job execution
chrono.WithTimeout(30 * time.Second)

// Set maximum retry attempts on failure
chrono.WithMaxRetries(3)

// Register a success callback
chrono.WithOnSuccess(func(jobID string) {
    log.Printf("Job %s completed successfully", jobID)
})

// Register an error callback
chrono.WithOnError(func(jobID string, err error) {
    log.Printf("Job %s failed: %v", jobID, err)
})
```

### Example with Options

```go
s.AddCronJob("report", "Daily Report", generateReport, "0 8 * * 1-5",
    chrono.WithTimeout(5*time.Minute),
    chrono.WithMaxRetries(2),
    chrono.WithOnSuccess(func(id string) {
        log.Println("Report generated successfully")
    }),
    chrono.WithOnError(func(id string, err error) {
        alert.Send("Report generation failed: " + err.Error())
    }),
)
```

## Job Management

```go
// Pause a job — it will not be executed until resumed
s.PauseJob("cleanup")

// Resume a paused job — next run time is recomputed from now
s.ResumeJob("cleanup")

// Remove a job entirely
s.RemoveJob("cleanup")

// Get job information
info, err := s.GetJob("heartbeat")
if err == nil {
    fmt.Printf("Status: %s, Runs: %d, Errors: %d\n", info.Status, info.RunCount, info.ErrorCount)
}

// List all jobs
for _, job := range s.ListJobs() {
    fmt.Printf("%-15s %-12s Next: %s\n", job.ID, job.Status, job.NextRun)
}
```

## Scheduler Options

| Option                       | Description                                                    | Default           |
| ---------------------------- | -------------------------------------------------------------- | ----------------- |
| `WithCheckInterval(d)`       | Sets the storage poll interval (backward-compatible alias)     | `1s`              |
| `WithStoragePollInterval(d)` | Interval for polling storage to detect external changes        | `30s`             |
| `WithStorage(store)`         | Sets the storage backend                                       | `InMemoryStorage` |
| `WithInstanceID(id)`         | Unique identifier for this scheduler instance (for clustering) | auto-generated    |
| `WithLockTTL(d)`             | Time-to-live for job execution locks                           | `5m`              |

```go
s := chrono.New(
    chrono.WithStoragePollInterval(15 * time.Second),
    chrono.WithInstanceID("worker-1"),
    chrono.WithLockTTL(10 * time.Minute),
)
```

> **Note:** `WithCheckInterval` is kept for backward compatibility. When set, it also applies to the storage poll interval unless `WithStoragePollInterval` is explicitly provided.

## Storage

Chrono uses a `Storage` interface to persist job state and coordinate execution across instances. The scheduler separates **job functions** (registered locally) from **job metadata** (persisted in storage), enabling multi-instance coordination where each instance registers the same functions and the storage layer ensures only one instance executes each job at a time.

### Storage Interface

```go
type Storage interface {
    // SaveJob persists a job record (upsert).
    SaveJob(ctx context.Context, record *JobRecord) error

    // GetJob retrieves a job record by ID.
    // Returns ErrJobNotFound if the job does not exist.
    GetJob(ctx context.Context, id string) (*JobRecord, error)

    // DeleteJob removes a job record by ID.
    // Returns ErrJobNotFound if the job does not exist.
    DeleteJob(ctx context.Context, id string) error

    // ListJobs returns all stored job records.
    ListJobs(ctx context.Context) ([]*JobRecord, error)

    // GetDueJobs returns jobs where NextRun <= now, not paused, and NextRun is non-zero.
    GetDueJobs(ctx context.Context, now time.Time) ([]*JobRecord, error)

    // AcquireLock attempts to acquire a distributed execution lock for a job.
    // Returns true if the lock was acquired, false if held by another owner.
    // The lock auto-expires after the TTL to handle crashed instances.
    AcquireLock(ctx context.Context, jobID string, ownerID string, ttl time.Duration) (bool, error)

    // ReleaseLock releases the execution lock. Only the lock owner can release it.
    ReleaseLock(ctx context.Context, jobID string, ownerID string) error

    // Close releases any resources held by the storage.
    Close() error
}
```

### JobRecord

`JobRecord` is the serializable representation of a job's metadata and execution state. This is what gets persisted in storage:

| Field        | Type        | Description                                                    |
| ------------ | ----------- | -------------------------------------------------------------- |
| `ID`         | `string`    | Unique identifier of the job                                   |
| `Name`       | `string`    | Human-readable name                                            |
| `Status`     | `JobStatus` | Current status: pending, running, completed, failed, cancelled |
| `Paused`     | `bool`      | Whether the job is paused                                      |
| `LastRun`    | `time.Time` | Time the job was last executed                                 |
| `NextRun`    | `time.Time` | Scheduled time for the next execution                          |
| `RunCount`   | `int64`     | Total number of executions                                     |
| `ErrorCount` | `int64`     | Total number of failed executions                              |
| `LastError`  | `string`    | Error message from the most recent failure                     |

All fields have JSON, XML, and YAML struct tags for codec compatibility.

### Built-in Implementations

| Storage           | Constructor            | Use Case                                          |
| ----------------- | ---------------------- | ------------------------------------------------- |
| `InMemoryStorage` | `NewInMemoryStorage()` | Single-instance, no persistence required          |
| `FileStorage`     | `NewFileStorage(path)` | Single-instance, file persistence (YAML/JSON/XML) |

### In-Memory Storage

The default storage. Jobs and locks are held in memory. State is lost on restart. Ideal for single-instance deployments where persistence is not needed.

```go
// Explicitly using in-memory storage (this is the default)
s := chrono.New(chrono.WithStorage(chrono.NewInMemoryStorage()))

// Equivalent — in-memory is used when no storage is specified
s := chrono.New()
```

### File Storage

`FileStorage` persists all job state and lock information to a single file using golly's `codec` package. The serialization format is automatically determined from the file extension using `fsutils.LookupContentType`:

| Extension       | Format |
| --------------- | ------ |
| `.yaml`, `.yml` | YAML   |
| `.json`         | JSON   |
| `.xml`          | XML    |

```go
// YAML format
store, err := chrono.NewFileStorage("/var/lib/myapp/chrono.yaml")
if err != nil {
    log.Fatal(err)
}
s := chrono.New(chrono.WithStorage(store))
```

```go
// JSON format
store, err := chrono.NewFileStorage("/var/lib/myapp/chrono.json")
```

```go
// XML format
store, err := chrono.NewFileStorage("/var/lib/myapp/chrono.xml")
```

**Behavior details:**

- The directory is created automatically if it does not exist
- If the file already exists, its contents are loaded on first access
- All reads and writes are serialized through a mutex
- State is written atomically (write to temp file, then rename)

### Custom Storage

Implement the `Storage` interface to integrate with any persistence layer (PostgreSQL, Redis, MongoDB, etcd, etc.). Below is a complete skeleton for a custom storage implementation:

```go
package mystore

import (
    "context"
    "time"

    "oss.nandlabs.io/golly/chrono"
)

// RedisStorage is an example custom Storage backed by Redis.
type RedisStorage struct {
    // your Redis client, connection pool, etc.
}

// NewRedisStorage creates a new Redis-backed storage.
func NewRedisStorage(addr string) (*RedisStorage, error) {
    // Initialize connection...
    return &RedisStorage{}, nil
}

// SaveJob persists a job record (upsert).
// Use a Redis hash or serialized JSON value keyed by job ID.
func (r *RedisStorage) SaveJob(ctx context.Context, record *chrono.JobRecord) error {
    // SET chrono:job:<id> <serialized record>
    return nil
}

// GetJob retrieves a job record by ID.
// Return chrono.ErrJobNotFound if the key does not exist.
func (r *RedisStorage) GetJob(ctx context.Context, id string) (*chrono.JobRecord, error) {
    // GET chrono:job:<id>
    // if not found: return nil, chrono.ErrJobNotFound
    return nil, nil
}

// DeleteJob removes a job record by ID.
// Return chrono.ErrJobNotFound if the key does not exist.
func (r *RedisStorage) DeleteJob(ctx context.Context, id string) error {
    // DEL chrono:job:<id>
    return nil
}

// ListJobs returns all stored job records.
// Use SCAN or maintain a separate set of job IDs for efficient listing.
func (r *RedisStorage) ListJobs(ctx context.Context) ([]*chrono.JobRecord, error) {
    // SCAN for chrono:job:* keys, deserialize each
    return nil, nil
}

// GetDueJobs returns job records that are due for execution.
// Use a sorted set with NextRun as the score for efficient range queries:
//   ZRANGEBYSCORE chrono:due 0 <now_unix>
// Filter out paused jobs and zero NextRun values.
func (r *RedisStorage) GetDueJobs(ctx context.Context, now time.Time) ([]*chrono.JobRecord, error) {
    return nil, nil
}

// AcquireLock attempts to acquire a distributed lock for the given job.
// Use SET NX EX (Redis single-key lock) or Redlock for stronger guarantees.
// Return true if the lock was acquired, false if held by another owner.
// Re-acquiring by the same owner should extend the TTL.
func (r *RedisStorage) AcquireLock(ctx context.Context, jobID, ownerID string, ttl time.Duration) (bool, error) {
    // SET chrono:lock:<jobID> <ownerID> NX EX <ttl_seconds>
    return false, nil
}

// ReleaseLock releases the execution lock for the given job.
// Only release if the current owner matches (use a Lua script for atomicity):
//   if redis.call("get", key) == ownerID then redis.call("del", key) end
func (r *RedisStorage) ReleaseLock(ctx context.Context, jobID, ownerID string) error {
    return nil
}

// Close releases any resources (close the Redis connection pool).
func (r *RedisStorage) Close() error {
    return nil
}
```

#### Using Custom Storage

```go
store, err := mystore.NewRedisStorage("localhost:6379")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

s := chrono.New(
    chrono.WithStorage(store),
    chrono.WithInstanceID("worker-1"),
    chrono.WithLockTTL(10 * time.Minute),
    chrono.WithStoragePollInterval(10 * time.Second),
)
```

#### Implementation Guidelines

When building a custom `Storage`:

| Method                 | Key Considerations                                                                                    |
| ---------------------- | ----------------------------------------------------------------------------------------------------- |
| `SaveJob`              | Must be an upsert (insert or update). Handle concurrent writes safely.                                |
| `GetJob` / `DeleteJob` | Return `chrono.ErrJobNotFound` when the record does not exist.                                        |
| `GetDueJobs`           | Filter: `NextRun <= now AND NOT Paused AND NextRun != zero`. Use indexes/sorted sets for performance. |
| `AcquireLock`          | Must be atomic. Support TTL-based expiry. Same-owner re-acquisition should extend the lock.           |
| `ReleaseLock`          | Only the owning instance should be able to release. Use compare-and-delete.                           |
| `Close`                | Release connections, file handles, or other resources.                                                |

> **Important:** All methods must be safe for concurrent use from multiple goroutines. The scheduler calls storage methods from the run loop and from job execution goroutines concurrently.

### Cluster Deployment

For multi-instance deployments, use a shared storage backend. Chrono uses distributed locks to ensure each job is executed by only one instance at a time.

```go
// Instance 1
s1 := chrono.New(
    chrono.WithStorage(sharedStore),
    chrono.WithInstanceID("instance-1"),
)
s1.AddCronJob("cleanup", "Cleanup", cleanupFunc, "*/5 * * * *")
s1.Start()

// Instance 2 (same jobs registered, storage coordinates execution)
s2 := chrono.New(
    chrono.WithStorage(sharedStore),
    chrono.WithInstanceID("instance-2"),
)
s2.AddCronJob("cleanup", "Cleanup", cleanupFunc, "*/5 * * * *")
s2.Start()
```

**Key points for cluster usage:**

- Each instance must have a unique `instanceID` (auto-generated from hostname + PID if not set)
- All instances must register the same job functions locally (functions cannot be serialized)
- The storage backend handles state persistence and lock coordination
- Set `lockTTL` longer than your longest-running job to prevent duplicate execution
- The background storage poll (`WithStoragePollInterval`) detects jobs added or modified by other instances

## Schedule Types

You can also create schedules directly and use `AddJob`:

```go
// Cron schedule
cron, _ := chrono.NewCronSchedule("*/10 * * * *")

// Interval schedule
interval, _ := chrono.NewIntervalSchedule(5 * time.Minute)

// One-shot schedule (by delay from now)
oneshot, _ := chrono.NewOneShotSchedule(10 * time.Second)

// One-shot schedule (at a specific time)
at := chrono.NewOneShotScheduleAt(time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC))

// Add with any schedule
s.AddJob("my-job", "My Job", myFunc, cron)
```

## Error Handling

Chrono defines the following sentinel errors:

| Error                 | Description                                            |
| --------------------- | ------------------------------------------------------ |
| `ErrSchedulerRunning` | Returned when calling `Start()` on a running scheduler |
| `ErrSchedulerStopped` | Returned when calling `Stop()` on a stopped scheduler  |
| `ErrJobNotFound`      | Job with the given ID does not exist                   |
| `ErrJobAlreadyExists` | Job with the given ID is already registered            |
| `ErrInvalidCronExpr`  | Cron expression is malformed                           |
| `ErrInvalidInterval`  | Interval duration is zero or negative                  |
| `ErrInvalidDelay`     | Delay duration is zero or negative                     |
| `ErrNilJobFunc`       | A nil function was provided                            |
| `ErrEmptyJobID`       | An empty job ID was provided                           |

```go
err := s.AddCronJob("job1", "Job", fn, "bad cron")
if errors.Is(err, chrono.ErrInvalidCronExpr) {
    log.Println("Invalid cron expression")
}
```
