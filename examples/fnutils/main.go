// Package main demonstrates the fnutils package.
package main

import (
	"fmt"
	"time"

	"oss.nandlabs.io/golly/fnutils"
)

func main() {
	// Execute a function after a delay (in seconds)
	fmt.Println("Scheduling task to run after 2 seconds...")
	start := time.Now()
	err := fnutils.ExecuteAfterSecs(func() {
		fmt.Printf("Task executed after %v\n", time.Since(start).Round(time.Second))
	}, 2)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Execute after milliseconds
	fmt.Println("Scheduling task to run after 500ms...")
	start = time.Now()
	err = fnutils.ExecuteAfterMs(func() {
		fmt.Printf("Task executed after %v\n", time.Since(start).Round(100*time.Millisecond))
	}, 500)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
