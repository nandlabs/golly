# FnUtils Package

The `fnutils` package provides utility functions for deferred and timed function execution in Go.

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

- Execute a function after a configurable delay
- Support for delays in seconds, milliseconds, minutes, or `time.Duration`

## Usage

```go
import "oss.nandlabs.io/golly/fnutils"

// Execute after 2 seconds
fnutils.ExecuteAfterSecs(func() {
    fmt.Println("Executed after 2 seconds")
}, 2)

// Execute after 500 milliseconds
fnutils.ExecuteAfterMs(func() {
    fmt.Println("Executed after 500ms")
}, 500)

// Execute after 1 minute
fnutils.ExecuteAfterMin(func() {
    fmt.Println("Executed after 1 minute")
}, 1)

// Execute after a custom duration
fnutils.ExecuteAfter(func() {
    fmt.Println("Custom delay")
}, 3 * time.Second)
```
