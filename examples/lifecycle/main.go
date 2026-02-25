// Package main demonstrates the lifecycle package for component management.
package main

import (
	"fmt"
	"log"
	"time"

	"oss.nandlabs.io/golly/lifecycle"
)

func main() {
	// Create a component manager
	manager := lifecycle.NewSimpleComponentManager()

	// Create components with lifecycle hooks
	db := &lifecycle.SimpleComponent{
		CompId: "database",
		StartFunc: func() error {
			fmt.Println("[database] Connecting to database...")
			time.Sleep(100 * time.Millisecond)
			fmt.Println("[database] Connected.")
			return nil
		},
		StopFunc: func() error {
			fmt.Println("[database] Disconnecting...")
			return nil
		},
		BeforeStart: func() {
			fmt.Println("[database] Preparing to start...")
		},
		AfterStart: func(err error) {
			if err != nil {
				fmt.Println("[database] Start failed:", err)
			} else {
				fmt.Println("[database] Start complete.")
			}
		},
	}

	cache := &lifecycle.SimpleComponent{
		CompId: "cache",
		StartFunc: func() error {
			fmt.Println("[cache] Initializing cache...")
			time.Sleep(50 * time.Millisecond)
			fmt.Println("[cache] Ready.")
			return nil
		},
		StopFunc: func() error {
			fmt.Println("[cache] Flushing and closing cache...")
			return nil
		},
	}

	server := &lifecycle.SimpleComponent{
		CompId: "http-server",
		StartFunc: func() error {
			fmt.Println("[http-server] Starting HTTP server...")
			return nil
		},
		StopFunc: func() error {
			fmt.Println("[http-server] Shutting down HTTP server...")
			return nil
		},
	}

	// Register components
	manager.Register(db)
	manager.Register(cache)
	manager.Register(server)

	// Add dependencies: server depends on database and cache
	if err := manager.AddDependency("http-server", "database"); err != nil {
		log.Fatal(err)
	}
	if err := manager.AddDependency("http-server", "cache"); err != nil {
		log.Fatal(err)
	}
	// Cache depends on database
	if err := manager.AddDependency("cache", "database"); err != nil {
		log.Fatal(err)
	}

	// Track state changes
	manager.OnChange("database", func(prev, next lifecycle.ComponentState) {
		fmt.Printf("  [state change] database: %d -> %d\n", prev, next)
	})

	// Start all components (respects dependency order)
	fmt.Println("\n--- Starting all components ---")
	if err := manager.StartAll(); err != nil {
		log.Fatal("StartAll failed:", err)
	}

	// List all components and their states
	fmt.Println("\n--- Component States ---")
	for _, comp := range manager.List() {
		fmt.Printf("  %s: state=%d\n", comp.Id(), comp.State())
	}

	// Stop all components
	fmt.Println("\n--- Stopping all components ---")
	if err := manager.StopAll(); err != nil {
		log.Fatal("StopAll failed:", err)
	}

	fmt.Println("\nDone.")
}
