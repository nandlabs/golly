# golly/lifecycle

[![Go Reference](https://pkg.go.dev/badge/oss.nandlabs.io/golly/lifecycle.svg)](https://pkg.go.dev/oss.nandlabs.io/golly/lifecycle)

## Overview

The `lifecycle` package provides a component lifecycle management framework for Go applications. It lets you register, start, stop, and monitor application components in a controlled manner -- including dependency ordering, state tracking, timeout enforcement, and graceful shutdown on OS signals.

## Installation

```sh
go get oss.nandlabs.io/golly/lifecycle
```

## Key Concepts

| Concept | Description |
|---|---|
| **Component** | An entity with `Start()` / `Stop()` methods and observable state |
| **ComponentManager** | Orchestrates multiple components: registration, ordering, lifecycle operations |
| **SimpleComponent** | Ready-made `Component` implementation backed by function callbacks |
| **Dependencies** | Declare ordering constraints so dependents start after, and stop before, their dependencies |
| **Timeout Methods** | Run any lifecycle operation with a deadline -- returns `ErrTimeout` on expiry |

## Component States

```
Unknown -> Starting -> Running -> Stopping -> Stopped
                  \      Error      /
```

| State | Value | Meaning |
|---|---|---|
| `Unknown` | 0 | Initial / unset |
| `Error` | 1 | Start or stop returned an error |
| `Stopped` | 2 | Successfully stopped |
| `Stopping` | 3 | Stop in progress |
| `Running` | 4 | Successfully started |
| `Starting` | 5 | Start in progress |

## Quick Start

```go
package main

import (
    "fmt"
    "oss.nandlabs.io/golly/lifecycle"
)

func main() {
    // 1. Create a manager
    manager := lifecycle.NewSimpleComponentManager()

    // 2. Register components
    manager.Register(&lifecycle.SimpleComponent{
        CompId: "db",
        StartFunc: func() error {
            fmt.Println("DB connected")
            return nil
        },
        StopFunc: func() error {
            fmt.Println("DB disconnected")
            return nil
        },
    })

    manager.Register(&lifecycle.SimpleComponent{
        CompId: "server",
        StartFunc: func() error {
            fmt.Println("Server listening")
            return nil
        },
        StopFunc: func() error {
            fmt.Println("Server shut down")
            return nil
        },
    })

    // 3. Declare dependencies (server depends on db)
    manager.AddDependency("server", "db")

    // 4. Start all and wait for SIGINT/SIGTERM
    manager.StartAndWait()
}
```

## API Reference

### Component Interface

```go
type Component interface {
    Id() string
    Start() error
    Stop() error
    State() ComponentState
    OnChange(f func(prevState, newState ComponentState))
}
```

### SimpleComponent

A concrete `Component` backed by callback functions:

```go
comp := &lifecycle.SimpleComponent{
    CompId:    "my-service",
    StartFunc: func() error { /* start logic */ return nil },
    StopFunc:  func() error { /* stop logic */ return nil },
}
```

### ComponentManager Interface

| Method | Description |
|---|---|
| `Register(component)` | Register a component |
| `Unregister(id)` | Unregister and stop a component |
| `AddDependency(id, dependsOn)` | Declare that `id` depends on `dependsOn` |
| `Start(id)` | Start a single component (starts dependencies first) |
| `StartWithTimeout(id, timeout)` | Start with a deadline; returns `ErrTimeout` on expiry |
| `StartAll()` | Start all registered components in dependency order |
| `StartAllWithTimeout(timeout)` | Start all with a deadline |
| `StartAndWait()` | Start all and block until `StopAll` or OS signal |
| `Stop(id)` | Stop a single component (stops dependents first) |
| `StopWithTimeout(id, timeout)` | Stop with a deadline; returns `ErrTimeout` on expiry |
| `StopAll()` | Stop all components in reverse order |
| `StopAllWithTimeout(timeout)` | Stop all with a deadline |
| `Wait()` | Block until all components are stopped |
| `GetState(id)` | Return a component's current state |
| `List()` | Return all registered components |
| `OnChange(id, f)` | Register a state-change callback |

### Timeout Methods

The timeout variants wrap the corresponding operation in a goroutine and apply `time.After`. If the operation does not complete within the specified duration, `ErrTimeout` is returned.

```go
import "time"

// Start with a 5-second deadline
err := manager.StartWithTimeout("db", 5*time.Second)
if errors.Is(err, lifecycle.ErrTimeout) {
    log.Fatal("db took too long to start")
}

// Stop all with a 10-second deadline
err = manager.StopAllWithTimeout(10 * time.Second)
```

### Sentinel Errors

| Error | Meaning |
|---|---|
| `ErrCompNotFound` | Component ID not registered |
| `ErrCompAlreadyStarted` | Component is already running |
| `ErrCompAlreadyStopped` | Component is already stopped |
| `ErrInvalidComponentState` | Invalid state transition |
| `ErrCyclicDependency` | Dependency graph has a cycle |
| `ErrTimeout` | Operation exceeded its timeout |

## Dependencies

Components can declare dependencies so that:

- On **Start**, dependencies are started first (with a `sync.WaitGroup`).
- On **Stop**, dependents are stopped before the component itself.

```go
manager.AddDependency("server", "db")    // server depends on db
manager.AddDependency("server", "cache") // server also depends on cache

manager.StartAll() // starts db, cache first, then server
manager.StopAll()  // stops server first, then db and cache
```

Cyclic dependencies are detected and return `ErrCyclicDependency`.

## Graceful Shutdown

`NewSimpleComponentManager()` automatically listens for `SIGINT` and `SIGTERM`. On receipt, it calls `StopAll()` to shut down components in reverse order. Use `StartAndWait()` to block the main goroutine until shutdown completes.

## Custom Components

Implement the `Component` interface directly for full control:

```go
type MyServer struct {
    state lifecycle.ComponentState
    mu    sync.RWMutex
}

func (s *MyServer) Id() string { return "my-server" }

func (s *MyServer) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ... start server ...
    s.state = lifecycle.Running
    return nil
}

func (s *MyServer) Stop() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ... stop server ...
    s.state = lifecycle.Stopped
    return nil
}

func (s *MyServer) State() lifecycle.ComponentState {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.state
}

func (s *MyServer) OnChange(f func(prev, next lifecycle.ComponentState)) {
    // optional: store and call on state transitions
}
```

For more information, refer to the [GoDoc](https://pkg.go.dev/oss.nandlabs.io/golly/lifecycle) documentation.
