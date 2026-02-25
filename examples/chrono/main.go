// Package main demonstrates the chrono package for task scheduling.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"oss.nandlabs.io/golly/chrono"
)

func main() {
	fmt.Println("=== Chrono Scheduler Examples ===")
	fmt.Println()

	// --- Example 1: Basic Interval Job ---
	fmt.Println("1. Basic Interval Job")
	basicIntervalExample()

	// --- Example 2: One-Shot Job ---
	fmt.Println("2. One-Shot Job")
	oneShotExample()

	// --- Example 3: Job with Timeout and Retries ---
	fmt.Println("3. Job with Timeout and Retries")
	retriesExample()

	// --- Example 4: Job Management (Pause, Resume, Remove) ---
	fmt.Println("4. Job Management (Pause, Resume, Remove)")
	managementExample()

	// --- Example 5: File Storage ---
	fmt.Println("5. File Storage (YAML)")
	fileStorageExample()

	// --- Example 6: Cron Schedule ---
	fmt.Println("6. Cron Schedule (simulated)")
	cronExample()

	// --- Example 7: Graceful Shutdown ---
	fmt.Println("7. Graceful Shutdown (Ctrl+C to test)")
	fmt.Println("   Skipped in demo mode. See gracefulShutdownExample().")
	fmt.Println()

	fmt.Println("=== All examples completed ===")
}

// basicIntervalExample creates a scheduler with an interval job that runs every 200ms.
func basicIntervalExample() {
	s := chrono.New(chrono.WithCheckInterval(50 * time.Millisecond))

	var count atomic.Int32

	err := s.AddIntervalJob("ticker", "Tick Counter", func(ctx context.Context) error {
		n := count.Add(1)
		fmt.Printf("   tick #%d\n", n)
		return nil
	}, 200*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	s.Start()
	time.Sleep(650 * time.Millisecond)
	s.Stop()

	fmt.Printf("   Total ticks: %d\n\n", count.Load())
}

// oneShotExample creates a one-shot job that runs once after a short delay.
func oneShotExample() {
	s := chrono.New(chrono.WithCheckInterval(50 * time.Millisecond))

	done := make(chan struct{})

	s.AddOneShotJob("init", "Initialize", func(ctx context.Context) error {
		fmt.Println("   One-shot job executed!")
		close(done)
		return nil
	}, 100*time.Millisecond)

	s.Start()
	<-done
	s.Stop()

	info, _ := s.GetJob("init")
	fmt.Printf("   Status: %s, RunCount: %d\n\n", info.Status, info.RunCount)
}

// retriesExample demonstrates a job that fails and retries before succeeding.
func retriesExample() {
	s := chrono.New(chrono.WithCheckInterval(50 * time.Millisecond))

	var attempt atomic.Int32

	s.AddOneShotJob("flaky", "Flaky Task", func(ctx context.Context) error {
		n := attempt.Add(1)
		if n <= 2 {
			fmt.Printf("   Attempt %d: failed\n", n)
			return fmt.Errorf("transient error")
		}
		fmt.Printf("   Attempt %d: success!\n", n)
		return nil
	}, 100*time.Millisecond,
		chrono.WithMaxRetries(3),
		chrono.WithOnSuccess(func(id string) {
			fmt.Printf("   Callback: job %q succeeded\n", id)
		}),
		chrono.WithOnError(func(id string, err error) {
			fmt.Printf("   Callback: job %q failed: %v\n", id, err)
		}),
	)

	s.Start()
	time.Sleep(500 * time.Millisecond)
	s.Stop()

	info, _ := s.GetJob("flaky")
	fmt.Printf("   Final status: %s, runs: %d, errors: %d\n\n", info.Status, info.RunCount, info.ErrorCount)
}

// managementExample demonstrates pausing, resuming, listing, and removing jobs.
func managementExample() {
	s := chrono.New(chrono.WithCheckInterval(50 * time.Millisecond))

	var count atomic.Int32

	s.AddIntervalJob("counter", "Counter", func(ctx context.Context) error {
		count.Add(1)
		return nil
	}, 100*time.Millisecond)

	s.Start()
	time.Sleep(350 * time.Millisecond)
	before := count.Load()
	fmt.Printf("   Before pause: %d runs\n", before)

	// Pause the job
	s.PauseJob("counter")
	fmt.Println("   Job paused")
	time.Sleep(300 * time.Millisecond)
	afterPause := count.Load()
	fmt.Printf("   After 300ms paused: %d runs (should be same)\n", afterPause)

	// Resume
	s.ResumeJob("counter")
	fmt.Println("   Job resumed")
	time.Sleep(350 * time.Millisecond)
	afterResume := count.Load()
	fmt.Printf("   After resume: %d runs\n", afterResume)

	// List all jobs
	fmt.Println("   All jobs:")
	for _, job := range s.ListJobs() {
		fmt.Printf("     - %-12s status=%-10s runs=%d next=%s\n",
			job.ID, job.Status, job.RunCount, job.NextRun.Format("15:04:05.000"))
	}

	// Remove the job
	s.RemoveJob("counter")
	fmt.Println("   Job removed")

	s.Stop()
	fmt.Println()
}

// fileStorageExample persists scheduler state to a YAML file.
func fileStorageExample() {
	dir, _ := os.MkdirTemp("", "chrono-example-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "state.yaml")

	store, err := chrono.NewFileStorage(path)
	if err != nil {
		log.Fatal(err)
	}

	s := chrono.New(
		chrono.WithStorage(store),
		chrono.WithCheckInterval(50*time.Millisecond),
	)

	s.AddOneShotJob("persist-test", "Persisted Job", func(ctx context.Context) error {
		fmt.Println("   Persisted job executed!")
		return nil
	}, 100*time.Millisecond)

	s.Start()
	time.Sleep(400 * time.Millisecond)
	s.Stop()

	// Read back the file to show persistence
	data, _ := os.ReadFile(path)
	fmt.Println("   Saved state file:")
	fmt.Println("   ---")
	for _, line := range splitLines(string(data)) {
		fmt.Println("   " + line)
	}
	fmt.Println("   ---")
	fmt.Println()
}

// cronExample demonstrates creating a cron schedule (uses a near-future time for demo).
func cronExample() {
	// Create a cron schedule that fires every minute
	cron, err := chrono.NewCronSchedule("* * * * *")
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	next1 := cron.Next(now)
	next2 := cron.Next(next1)
	next3 := cron.Next(next2)

	fmt.Printf("   Now:    %s\n", now.Format("15:04:05"))
	fmt.Printf("   Next 1: %s\n", next1.Format("15:04:05"))
	fmt.Printf("   Next 2: %s\n", next2.Format("15:04:05"))
	fmt.Printf("   Next 3: %s\n", next3.Format("15:04:05"))

	// Also show custom schedules
	interval, _ := chrono.NewIntervalSchedule(30 * time.Second)
	fmt.Printf("   Interval (30s) next: %s\n", interval.Next(now).Format("15:04:05"))

	oneshot, _ := chrono.NewOneShotSchedule(5 * time.Second)
	fmt.Printf("   OneShot (5s) next:   %s\n", oneshot.Next(now).Format("15:04:05"))

	at := chrono.NewOneShotScheduleAt(time.Date(2026, 12, 31, 23, 59, 0, 0, time.Local))
	fmt.Printf("   OneShotAt (NYE):     %s\n", at.Next(now).Format("2006-01-02 15:04:05"))
	fmt.Println()
}

// gracefulShutdownExample demonstrates clean shutdown on SIGINT/SIGTERM.
// Not called in demo mode to avoid blocking.
func gracefulShutdownExample() {
	s := chrono.New()

	s.AddIntervalJob("worker", "Background Worker", func(ctx context.Context) error {
		fmt.Println("Working...")
		time.Sleep(2 * time.Second) // simulate work
		fmt.Println("Work done")
		return nil
	}, 5*time.Second)

	s.Start()
	fmt.Println("Scheduler running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\nShutting down gracefully...")
	s.Stop()
	fmt.Println("Shutdown complete")
}

// splitLines splits a string into lines, filtering out empty trailing lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
