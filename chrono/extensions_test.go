package chrono

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---- LeaderElector implementations ----

func TestAlwaysLeader(t *testing.T) {
	l := NewAlwaysLeader()
	if !l.IsLeader(context.Background()) {
		t.Error("NewAlwaysLeader should always report true")
	}
	if err := l.Resign(context.Background()); err != nil {
		t.Errorf("Resign should be no-op; got %v", err)
	}
}

func TestMemoryLeader_Toggle(t *testing.T) {
	m := NewMemoryLeader(false)
	if m.IsLeader(context.Background()) {
		t.Error("initial state should be false")
	}
	m.Set(true)
	if !m.IsLeader(context.Background()) {
		t.Error("after Set(true) should be leader")
	}
	_ = m.Resign(context.Background())
	if m.IsLeader(context.Background()) {
		t.Error("Resign should clear leader flag")
	}
}

// ---- WithResult callback fires ----

func TestWithResult_Success(t *testing.T) {
	sched := New(WithCheckInterval(50 * time.Millisecond))
	defer sched.Stop()

	var got JobResult
	var fired sync.WaitGroup
	fired.Add(1)
	err := sched.AddOneShotJob("ok", "", func(_ context.Context) error { return nil }, 10*time.Millisecond,
		WithResult(func(r JobResult) {
			got = r
			fired.Done()
		}))
	if err != nil {
		t.Fatal(err)
	}
	_ = sched.Start()
	waitWithTimeout(t, &fired, 2*time.Second)

	if got.JobID != "ok" || got.Err != nil {
		t.Errorf("result wrong: %+v", got)
	}
	if got.RetryCount != 0 {
		t.Errorf("RetryCount = %d, want 0", got.RetryCount)
	}
	if got.Duration() < 0 {
		t.Errorf("Duration should be non-negative; got %v", got.Duration())
	}
}

func TestWithResult_FailureCountsRetries(t *testing.T) {
	sched := New(WithCheckInterval(50 * time.Millisecond))
	defer sched.Stop()

	var got JobResult
	var fired sync.WaitGroup
	fired.Add(1)
	var attempts int32
	boom := errors.New("boom")
	err := sched.AddOneShotJob("bad", "", func(_ context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return boom
	}, 10*time.Millisecond,
		WithMaxRetries(2),
		WithResult(func(r JobResult) { got = r; fired.Done() }))
	if err != nil {
		t.Fatal(err)
	}
	_ = sched.Start()
	waitWithTimeout(t, &fired, 2*time.Second)

	if got.Err == nil {
		t.Errorf("expected error result; got %+v", got)
	}
	if got.RetryCount != 2 {
		t.Errorf("RetryCount = %d, want 2 (1 initial + 2 retries = 3 attempts → 2 retries)", got.RetryCount)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

// ---- LeaderElector gates dispatch ----

func TestWithLeaderElector_NonLeaderSkipsDispatch(t *testing.T) {
	leader := NewMemoryLeader(false) // start non-leader
	sched := New(WithCheckInterval(50*time.Millisecond), WithLeaderElector(leader))
	defer sched.Stop()

	var ran int32
	_ = sched.AddIntervalJob("x", "", func(_ context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}, 100*time.Millisecond)
	_ = sched.Start()
	// Wait long enough that, were the elector reporting true, the job would
	// have fired multiple times.
	time.Sleep(400 * time.Millisecond)
	if atomic.LoadInt32(&ran) != 0 {
		t.Errorf("non-leader instance should not dispatch; ran = %d", ran)
	}
	// Become leader; job should start running.
	leader.Set(true)
	time.Sleep(400 * time.Millisecond)
	if atomic.LoadInt32(&ran) == 0 {
		t.Errorf("after becoming leader, job should have run at least once")
	}
}

// ---- Misfire policy is recorded (semantic enforcement is impl follow-up) ----

func TestWithMisfire_RecordsOnConfig(t *testing.T) {
	cfg := jobConfig{}
	WithMisfire(MisfireFireAll)(&cfg)
	if cfg.misfire != MisfireFireAll {
		t.Errorf("misfire policy not recorded; got %v", cfg.misfire)
	}
	if !cfg.misfireSet {
		t.Errorf("misfireSet should be true after WithMisfire")
	}
}

// ---- helper: wait on WaitGroup with a hard timeout ----

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, d time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("timed out waiting (%v)", d)
	}
}
