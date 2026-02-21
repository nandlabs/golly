// Package main demonstrates the pool package for generic object pooling.
package main

import (
	"fmt"
	"log"
	"sync/atomic"

	"oss.nandlabs.io/golly/pool"
)

func main() {
	var counter atomic.Int64

	// Create a pool of database-like connections (simulated with strings)
	p, err := pool.NewPool(
		// Creator: creates new objects
		func() (string, error) {
			id := counter.Add(1)
			conn := fmt.Sprintf("connection-%d", id)
			fmt.Println("  Created:", conn)
			return conn, nil
		},
		// Destroyer: cleans up objects
		func(conn string) error {
			fmt.Println("  Destroyed:", conn)
			return nil
		},
		2,  // min: pre-create 2 objects
		5,  // max: allow up to 5 objects
		10, // maxWait: wait up to 10 seconds for an object
	)
	if err != nil {
		log.Fatal("Failed to create pool:", err)
	}

	// Start the pool (pre-creates min objects)
	fmt.Println("Starting pool...")
	if err := p.Start(); err != nil {
		log.Fatal("Failed to start pool:", err)
	}
	fmt.Printf("Pool started: current=%d, min=%d, max=%d\n\n", p.Current(), p.Min(), p.Max())

	// Checkout objects
	fmt.Println("Checking out objects...")
	conn1, err := p.Checkout()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  Got:", conn1)

	conn2, err := p.Checkout()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  Got:", conn2)

	fmt.Printf("  Current: %d, HighWaterMark: %d\n\n", p.Current(), p.HighWaterMark())

	// Return objects to the pool
	fmt.Println("Checking in objects...")
	p.Checkin(conn1)
	p.Checkin(conn2)
	fmt.Printf("  Current: %d\n\n", p.Current())

	// Close the pool
	fmt.Println("Closing pool...")
	if err := p.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pool closed.")
}
