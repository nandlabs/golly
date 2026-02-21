// Package main demonstrates the clients package: auth, retry, and circuit breaker.
package main

import (
	"fmt"
	"time"

	"oss.nandlabs.io/golly/clients"
)

func main() {
	// --- Basic Auth ---
	fmt.Println("=== Basic Auth ===")
	basicAuth := clients.NewBasicAuth("admin", "secret123")
	fmt.Println("Type:", basicAuth.Type())
	fmt.Println("Token:", basicAuth.Token())

	// --- Bearer Auth ---
	fmt.Println("\n=== Bearer Auth ===")
	bearerAuth := clients.NewBearerAuth("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example")
	fmt.Println("Type:", bearerAuth.Type())
	fmt.Println("Token:", bearerAuth.Token()[:30], "...")

	// --- Retry with exponential backoff ---
	fmt.Println("\n=== RetryInfo (Exponential Backoff) ===")
	retry := &clients.RetryInfo{
		MaxRetries:  5,
		Wait:        100, // 100ms base wait
		Exponential: true,
		Multiplier:  2.0,
		MaxWait:     5000, // 5s max
		Jitter:      true,
	}

	for i := 0; i < retry.MaxRetries; i++ {
		wait := retry.WaitTime(i)
		fmt.Printf("  Retry %d: wait %v\n", i+1, wait)
	}

	// --- Circuit Breaker ---
	fmt.Println("\n=== Circuit Breaker ===")
	cb := clients.NewCircuitBreaker(&clients.BreakerInfo{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		MaxHalfOpen:      1,
		Timeout:          5,
	})

	// Simulate requests
	fmt.Println("Simulating requests through circuit breaker:")
	for i := 0; i < 6; i++ {
		err := cb.CanExecute()
		if err != nil {
			fmt.Printf("  Request %d: blocked (%v)\n", i+1, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if i < 3 {
			// Simulate failures
			cb.OnExecution(false)
			fmt.Printf("  Request %d: failed (circuit notified)\n", i+1)
		} else {
			// Simulate success
			cb.OnExecution(true)
			fmt.Printf("  Request %d: success\n", i+1)
		}
	}
}
