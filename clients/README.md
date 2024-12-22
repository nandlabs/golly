# Clients

The `clients` package provides utility features for building robust and resilient client implementations in Go. This package includes support for retry handling and circuit breaker patterns, which are essential for creating fault-tolerant applications.

## Index

- [Installation](#installation)
- [Features](#features)
  - [RetryHandler](#retryhandler)
    - [Usage](#usage)
  - [CircuitBreaker](#circuitbreaker)
    - [Usage](#usage-1)
    - [CircuitBreaker States](#circuitbreaker-states)
    - [Configuration Parameters](#configuration-parameters)
- [License](#license)

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/clients
```

## Features

### RetryHandler

The `RetryHandler` feature allows you to configure retry logic for your client operations. This is useful for handling transient errors and ensuring that your client can recover from temporary failures.

#### Usage

To use the `RetryHandler`, you need to define the retry configuration using the `RetryInfo` struct.

```go
package main

import (
    "fmt"
    "time"
    "github.com/nandlabs/golly/clients"
)

func main() {
    retryInfo := clients.RetryInfo{
        MaxRetries: 3,
        Wait:       1000, // Wait time in milliseconds
    }

    for i := 0; i < retryInfo.MaxRetries; i++ {
        err := performOperation()
        if err == nil {
            fmt.Println("Operation succeeded")
            break
        }
        fmt.Printf("Operation failed: %v. Retrying...\n", err)
        time.Sleep(time.Duration(retryInfo.Wait) * time.Millisecond)
    }
}

func performOperation() error {
    // Simulate an operation that may fail
    return fmt.Errorf("simulated error")
}
```

### CircuitBreaker

The `CircuitBreaker` feature helps you to prevent cascading failures and improve the resilience of your client by stopping requests to a failing service. It transitions between different states (closed, open, half-open) based on the success or failure of requests.

#### Usage

To use the `CircuitBreaker`, you need to create an instance of the `CircuitBreaker` struct with the desired configuration.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/clients"
)

func main() {
    breakerInfo := &clients.BreakerInfo{
        FailureThreshold: 3,
        SuccessThreshold: 3,
        MaxHalfOpen:      5,
        Timeout:          300, // Timeout in seconds
    }

    cb := clients.NewCB(breakerInfo)

    for i := 0; i < 10; i++ {
        err := cb.CanExecute()
        if err != nil {
            fmt.Println("Circuit breaker is open. Cannot execute operation.")
            continue
        }

        err = performOperation()
        cb.OnExecution(err == nil)
        if err != nil {
            fmt.Printf("Operation failed: %v\n", err)
        } else {
            fmt.Println("Operation succeeded")
        }
    }
}

func performOperation() error {
    // Simulate an operation that may fail
    return fmt.Errorf("simulated error")
}
```

#### CircuitBreaker States

- `circuitClosed`: The circuit is closed, and requests can flow through.
- `circuitHalfOpen`: The circuit is partially open and allows limited requests for testing.
- `circuitOpen`: The circuit is open, and requests are blocked.

#### Configuration Parameters

- `FailureThreshold`: Number of consecutive failures required to open the circuit.
- `SuccessThreshold`: Number of consecutive successes required to close the circuit.
- `MaxHalfOpen`: Maximum number of requests allowed in the half-open state.
- `Timeout`: Timeout duration for the circuit to transition from open to half-open state.
