# Pool Package

The `pool` package provides a generic, thread-safe object pool implementation with configurable capacity, automatic lifecycle management, and wait-based checkout.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **Generic**: Pool any type using Go generics (`Pool[T]`)
- **Configurable capacity**: Set minimum and maximum pool size
- **Lifecycle management**: User-supplied creator and destroyer functions
- **Wait-based checkout**: Configurable max wait time when pool is exhausted
- **Thread-safe**: Safe for concurrent use from multiple goroutines
- **Metrics**: Track current size, high water mark, and pool state

## Usage

```go
import "oss.nandlabs.io/golly/pool"

// Create a pool with creator and destroyer functions
p, err := pool.NewPool(
    func() (*Connection, error) { return NewConnection(), nil },  // creator
    func(c *Connection) error { return c.Close() },                // destroyer
    2,   // min: pre-create 2 objects
    10,  // max: allow up to 10 objects
    30,  // maxWait: wait up to 30 seconds for an object
)

// Start the pool (pre-creates min objects)
p.Start()

// Checkout an object
conn, err := p.Checkout()

// Use the object...

// Return to pool
p.Checkin(conn)

// Pool metrics
fmt.Println(p.Current())       // current pool size
fmt.Println(p.HighWaterMark()) // peak pool size

// Close the pool (destroys all objects)
p.Close()
```
