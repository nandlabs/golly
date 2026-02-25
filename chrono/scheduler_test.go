package chrono

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCronSchedule_Valid(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"every minute", "* * * * *"},
		{"every 5 minutes", "*/5 * * * *"},
		{"hourly", "0 * * * *"},
		{"daily at midnight", "0 0 * * *"},
		{"weekdays at 9am", "0 9 * * 1-5"},
		{"specific minutes", "0,15,30,45 * * * *"},
		{"specific day and time", "30 14 1 * *"},
		{"range with step", "0-30/10 * * * *"},
		{"complex", "5,10,15 1-3 1,15 1-6 0,6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, err := NewCronSchedule(tt.expr)
			if err != nil {
				t.Fatalf("NewCronSchedule(%q) returned error: %v", tt.expr, err)
			}
			if cs == nil {
				t.Fatal("NewCronSchedule returned nil")
			}
		})
	}
}

func TestNewCronSchedule_Macros(t *testing.T) {
	macros := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly"}
	for _, m := range macros {
		t.Run(m, func(t *testing.T) {
			cs, err := NewCronSchedule(m)
			if err != nil {
				t.Fatalf("NewCronSchedule(%q) returned error: %v", m, err)
			}
			if cs == nil {
				t.Fatal("NewCronSchedule returned nil")
			}
		})
	}
}

func TestNewCronSchedule_Invalid(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * *"},
		{"invalid minute", "60 * * * *"},
		{"invalid hour", "* 24 * * *"},
		{"invalid day", "* * 32 * *"},
		{"invalid month", "* * * 13 *"},
		{"invalid dow", "* * * * 7"},
		{"invalid range", "* * 5-3 * *"},
		{"invalid step", "*/0 * * * *"},
		{"non-numeric", "abc * * * *"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCronSchedule(tt.expr)
			if err == nil {
				t.Fatalf("NewCronSchedule(%q) expected error, got nil", tt.expr)
			}
			if !errors.Is(err, ErrInvalidCronExpr) {
				t.Fatalf("expected ErrInvalidCronExpr, got: %v", err)
			}
		})
	}
}

func TestCronSchedule_Next(t *testing.T) {
	cs, _ := NewCronSchedule("* * * * *")
	from := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	next := cs.Next(from)
	expected := time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCronSchedule_NextEvery5Min(t *testing.T) {
	cs, _ := NewCronSchedule("*/5 * * * *")
	from := time.Date(2024, 1, 15, 10, 7, 0, 0, time.UTC)
	next := cs.Next(from)
	expected := time.Date(2024, 1, 15, 10, 10, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCronSchedule_NextHourly(t *testing.T) {
	cs, _ := NewCronSchedule("@hourly")
	from := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	next := cs.Next(from)
	expected := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCronSchedule_NextWeekday(t *testing.T) {
	cs, _ := NewCronSchedule("0 9 * * 1-5")
	from := time.Date(2024, 1, 13, 10, 0, 0, 0, time.UTC)
	next := cs.Next(from)
	expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCronSchedule_NextSpecificMonths(t *testing.T) {
	cs, _ := NewCronSchedule("0 0 1 1,4,7,10 *")
	from := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)
	next := cs.Next(from)
	expected := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCronSchedule_String(t *testing.T) {
	cs, _ := NewCronSchedule("*/5 * * * *")
	if cs.String() != "*/5 * * * *" {
		t.Fatalf("expected '*/5 * * * *', got '%s'", cs.String())
	}
}

func TestNewIntervalSchedule_Valid(t *testing.T) {
	s, err := NewIntervalSchedule(5 * time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Interval() != 5*time.Second {
		t.Fatalf("expected 5s interval, got %v", s.Interval())
	}
}

func TestNewIntervalSchedule_Invalid(t *testing.T) {
	_, err := NewIntervalSchedule(0)
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval, got: %v", err)
	}
	_, err = NewIntervalSchedule(-1 * time.Second)
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval, got: %v", err)
	}
}

func TestIntervalSchedule_Next(t *testing.T) {
	s, _ := NewIntervalSchedule(30 * time.Second)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	next := s.Next(from)
	expected := from.Add(30 * time.Second)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestNewOneShotSchedule_Valid(t *testing.T) {
	s, err := NewOneShotSchedule(5 * time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.RunAt().IsZero() {
		t.Fatal("RunAt should not be zero")
	}
}

func TestNewOneShotSchedule_ZeroDelay(t *testing.T) {
	s, err := NewOneShotSchedule(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.RunAt().IsZero() {
		t.Fatal("RunAt should not be zero")
	}
}

func TestNewOneShotSchedule_NegativeDelay(t *testing.T) {
	_, err := NewOneShotSchedule(-1 * time.Second)
	if !errors.Is(err, ErrInvalidDelay) {
		t.Fatalf("expected ErrInvalidDelay, got: %v", err)
	}
}

func TestOneShotSchedule_Next(t *testing.T) {
	target := time.Now().Add(1 * time.Hour)
	s := NewOneShotScheduleAt(target)
	next := s.Next(time.Now())
	if !next.Equal(target) {
		t.Fatalf("expected %v, got %v", target, next)
	}
	next = s.Next(target.Add(time.Minute))
	if !next.IsZero() {
		t.Fatalf("expected zero time after target, got %v", next)
	}
}

func TestNewOneShotScheduleAt(t *testing.T) {
	target := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	s := NewOneShotScheduleAt(target)
	if !s.RunAt().Equal(target) {
		t.Fatalf("expected RunAt %v, got %v", target, s.RunAt())
	}
}

func TestScheduler_AddJob_Errors(t *testing.T) {
	s := New()
	err := s.AddJob("", "test", func(ctx context.Context) error { return nil }, &IntervalSchedule{interval: time.Second})
	if !errors.Is(err, ErrEmptyJobID) {
		t.Fatalf("expected ErrEmptyJobID, got: %v", err)
	}
	err = s.AddJob("test", "test", nil, &IntervalSchedule{interval: time.Second})
	if !errors.Is(err, ErrNilJobFunc) {
		t.Fatalf("expected ErrNilJobFunc, got: %v", err)
	}
	err = s.AddJob("job1", "Job 1", func(ctx context.Context) error { return nil }, &IntervalSchedule{interval: time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = s.AddJob("job1", "Job 1 dup", func(ctx context.Context) error { return nil }, &IntervalSchedule{interval: time.Second})
	if !errors.Is(err, ErrJobAlreadyExists) {
		t.Fatalf("expected ErrJobAlreadyExists, got: %v", err)
	}
}

func TestScheduler_AddCronJob_InvalidExpr(t *testing.T) {
	s := New()
	err := s.AddCronJob("test", "test", func(ctx context.Context) error { return nil }, "bad expr")
	if !errors.Is(err, ErrInvalidCronExpr) {
		t.Fatalf("expected ErrInvalidCronExpr, got: %v", err)
	}
}

func TestScheduler_AddIntervalJob_InvalidInterval(t *testing.T) {
	s := New()
	err := s.AddIntervalJob("test", "test", func(ctx context.Context) error { return nil }, 0)
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval, got: %v", err)
	}
}

func TestScheduler_AddOneShotJob_InvalidDelay(t *testing.T) {
	s := New()
	err := s.AddOneShotJob("test", "test", func(ctx context.Context) error { return nil }, -1*time.Second)
	if !errors.Is(err, ErrInvalidDelay) {
		t.Fatalf("expected ErrInvalidDelay, got: %v", err)
	}
}

func TestScheduler_RemoveJob(t *testing.T) {
	s := New()
	_ = s.AddIntervalJob("job1", "Job 1", func(ctx context.Context) error { return nil }, time.Second)
	err := s.RemoveJob("job1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = s.RemoveJob("nonexistent")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
}

func TestScheduler_GetJob(t *testing.T) {
	s := New()
	_ = s.AddIntervalJob("job1", "Job 1", func(ctx context.Context) error { return nil }, time.Second)
	info, err := s.GetJob("job1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ID != "job1" || info.Name != "Job 1" {
		t.Fatalf("unexpected job info: %+v", info)
	}
	if info.Status != JobStatusPending {
		t.Fatalf("expected pending status, got: %s", info.Status)
	}
	_, err = s.GetJob("nonexistent")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
}

func TestScheduler_ListJobs(t *testing.T) {
	s := New()
	_ = s.AddIntervalJob("job1", "Job 1", func(ctx context.Context) error { return nil }, time.Second)
	_ = s.AddIntervalJob("job2", "Job 2", func(ctx context.Context) error { return nil }, time.Second)
	jobs := s.ListJobs()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s := New()
	if s.IsRunning() {
		t.Fatal("scheduler should not be running before Start()")
	}
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if !s.IsRunning() {
		t.Fatal("scheduler should be running after Start()")
	}
	err = s.Start()
	if !errors.Is(err, ErrSchedulerRunning) {
		t.Fatalf("expected ErrSchedulerRunning, got: %v", err)
	}
	err = s.Stop()
	if err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
	if s.IsRunning() {
		t.Fatal("scheduler should not be running after Stop()")
	}
	err = s.Stop()
	if !errors.Is(err, ErrSchedulerStopped) {
		t.Fatalf("expected ErrSchedulerStopped, got: %v", err)
	}
}

func TestScheduler_IntervalJobExecution(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var counter int32
	_ = s.AddIntervalJob("counter", "Counter", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 100*time.Millisecond)
	_ = s.Start()
	time.Sleep(350 * time.Millisecond)
	_ = s.Stop()
	count := atomic.LoadInt32(&counter)
	if count < 2 {
		t.Fatalf("expected at least 2 executions, got %d", count)
	}
}

func TestScheduler_OneShotJobExecution(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var counter int32
	_ = s.AddOneShotJob("once", "Run Once", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 100*time.Millisecond)
	_ = s.Start()
	time.Sleep(500 * time.Millisecond)
	_ = s.Stop()
	count := atomic.LoadInt32(&counter)
	if count != 1 {
		t.Fatalf("expected exactly 1 execution, got %d", count)
	}
}

func TestScheduler_JobWithTimeout(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var timedOut int32
	_ = s.AddIntervalJob("slow", "Slow Job", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			atomic.StoreInt32(&timedOut, 1)
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	}, 100*time.Millisecond, WithTimeout(50*time.Millisecond))
	_ = s.Start()
	time.Sleep(300 * time.Millisecond)
	_ = s.Stop()
	if atomic.LoadInt32(&timedOut) != 1 {
		t.Fatal("expected job to be timed out")
	}
}

func TestScheduler_JobWithRetries(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var attempts int32
	_ = s.AddOneShotJob("retry", "Retry Job", func(ctx context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return fmt.Errorf("intentional failure")
	}, 50*time.Millisecond, WithMaxRetries(2))
	_ = s.Start()
	time.Sleep(400 * time.Millisecond)
	_ = s.Stop()
	count := atomic.LoadInt32(&attempts)
	if count != 3 {
		t.Fatalf("expected 3 attempts (1 + 2 retries), got %d", count)
	}
	info, _ := s.GetJob("retry")
	if info.Status != JobStatusFailed {
		t.Fatalf("expected failed status, got: %s", info.Status)
	}
	if info.ErrorCount != 1 {
		t.Fatalf("expected 1 error count, got %d", info.ErrorCount)
	}
}

func TestScheduler_JobCallbacks(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var successCalled int32
	var errorCalled int32
	_ = s.AddOneShotJob("success", "Success Job", func(ctx context.Context) error {
		return nil
	}, 50*time.Millisecond,
		WithOnSuccess(func(id string) { atomic.StoreInt32(&successCalled, 1) }),
		WithOnError(func(id string, err error) { atomic.StoreInt32(&errorCalled, 1) }),
	)
	_ = s.Start()
	time.Sleep(300 * time.Millisecond)
	_ = s.Stop()
	if atomic.LoadInt32(&successCalled) != 1 {
		t.Fatal("expected success callback to be called")
	}
	if atomic.LoadInt32(&errorCalled) != 0 {
		t.Fatal("error callback should not be called")
	}
}

func TestScheduler_ErrorCallback(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var errorCalled int32
	var errorID string
	var errorMsg string
	_ = s.AddOneShotJob("fail", "Fail Job", func(ctx context.Context) error {
		return fmt.Errorf("test error")
	}, 50*time.Millisecond,
		WithOnError(func(id string, err error) {
			atomic.StoreInt32(&errorCalled, 1)
			errorID = id
			errorMsg = err.Error()
		}),
	)
	_ = s.Start()
	time.Sleep(300 * time.Millisecond)
	_ = s.Stop()
	if atomic.LoadInt32(&errorCalled) != 1 {
		t.Fatal("expected error callback to be called")
	}
	if errorID != "fail" {
		t.Fatalf("expected error job ID 'fail', got '%s'", errorID)
	}
	if errorMsg != "test error" {
		t.Fatalf("expected error message 'test error', got '%s'", errorMsg)
	}
}

func TestScheduler_PauseResumeJob(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var counter int32
	_ = s.AddIntervalJob("pausable", "Pausable", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 100*time.Millisecond)
	_ = s.Start()
	time.Sleep(350 * time.Millisecond)
	_ = s.PauseJob("pausable")
	// Allow any in-flight execution and storage sync to complete
	time.Sleep(250 * time.Millisecond)
	countAtPause := atomic.LoadInt32(&counter)
	time.Sleep(400 * time.Millisecond)
	countAfterPause := atomic.LoadInt32(&counter)
	if countAfterPause != countAtPause {
		t.Fatalf("job executed %d times while paused (was %d at pause)", countAfterPause, countAtPause)
	}
	_ = s.ResumeJob("pausable")
	time.Sleep(300 * time.Millisecond)
	_ = s.Stop()
	countAfterResume := atomic.LoadInt32(&counter)
	if countAfterResume <= countAtPause {
		t.Fatal("job should have executed after resume")
	}
}

func TestScheduler_PauseResume_NotFound(t *testing.T) {
	s := New()
	err := s.PauseJob("nonexistent")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
	err = s.ResumeJob("nonexistent")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got: %v", err)
	}
}

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status   JobStatus
		expected string
	}{
		{JobStatusPending, "pending"},
		{JobStatusRunning, "running"},
		{JobStatusCompleted, "completed"},
		{JobStatusFailed, "failed"},
		{JobStatusCancelled, "canceled"},
		{JobStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, tt.status.String())
			}
		})
	}
}

func TestScheduler_AddJobWhileRunning(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	_ = s.Start()
	var counter int32
	err := s.AddIntervalJob("dynamic", "Dynamic Job", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error adding job while running: %v", err)
	}
	time.Sleep(350 * time.Millisecond)
	_ = s.Stop()
	count := atomic.LoadInt32(&counter)
	if count < 1 {
		t.Fatalf("expected at least 1 execution, got %d", count)
	}
}

func TestScheduler_RemoveJobWhileRunning(t *testing.T) {
	s := New(WithCheckInterval(50 * time.Millisecond))
	var counter int32
	_ = s.AddIntervalJob("removable", "Removable", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 100*time.Millisecond)
	_ = s.Start()
	time.Sleep(250 * time.Millisecond)
	_ = s.RemoveJob("removable")
	// Allow any in-flight execution to complete
	time.Sleep(100 * time.Millisecond)
	countAfterRemove := atomic.LoadInt32(&counter)
	time.Sleep(250 * time.Millisecond)
	_ = s.Stop()
	countFinal := atomic.LoadInt32(&counter)
	// No new executions should happen after the in-flight one completes
	if countFinal != countAfterRemove {
		t.Fatalf("job executed after removal settled: settled=%d, final=%d", countAfterRemove, countFinal)
	}
}

func TestMakeRange(t *testing.T) {
	tests := []struct {
		start, end, step int
		expected         []int
	}{
		{0, 5, 1, []int{0, 1, 2, 3, 4, 5}},
		{0, 10, 3, []int{0, 3, 6, 9}},
		{1, 1, 1, []int{1}},
		{0, 59, 15, []int{0, 15, 30, 45}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d-%d/%d", tt.start, tt.end, tt.step), func(t *testing.T) {
			result := makeRange(tt.start, tt.end, tt.step)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Fatalf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestIntSliceContains(t *testing.T) {
	slice := []int{0, 5, 10, 15, 20}
	if !intSliceContains(slice, 0) {
		t.Fatal("should contain 0")
	}
	if !intSliceContains(slice, 15) {
		t.Fatal("should contain 15")
	}
	if intSliceContains(slice, 7) {
		t.Fatal("should not contain 7")
	}
	if intSliceContains(slice, -1) {
		t.Fatal("should not contain -1")
	}
}

func TestUniqueInts(t *testing.T) {
	input := []int{1, 2, 3, 2, 1, 4, 3}
	result := uniqueInts(input)
	if len(result) != 4 {
		t.Fatalf("expected 4 unique values, got %d: %v", len(result), result)
	}
}
