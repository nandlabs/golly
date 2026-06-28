package chrono

import (
	"context"
	"sync"
	"time"
)

// MisfirePolicy controls behavior when the scheduler discovers a job whose
// scheduled time has already passed (e.g. after a downtime). It does not
// override the per-run lock; it only governs WHICH missed instances are
// resurrected when the scheduler catches up.
type MisfirePolicy int

const (
	// MisfireSkip silently advances NextRun to the next future instance.
	// Best for tasks where stale data is worse than missed work
	// (e.g. metrics snapshots).
	MisfireSkip MisfirePolicy = iota

	// MisfireFireOnce coalesces all missed instances into one immediate
	// run, then resumes the normal cadence. Good default for idempotent
	// jobs (cleanups, syncs).
	MisfireFireOnce

	// MisfireFireAll fires every missed instance in order, back-to-back.
	// Use only when each instance is semantically distinct
	// (e.g. per-hour reports that must each be emitted).
	MisfireFireAll
)

// JobResult is delivered to a WithResult callback after a job execution
// completes (or fails). It complements WithOnSuccess / WithOnError by
// surfacing timing data and the return error in a single hook.
type JobResult struct {
	JobID      string
	Started    time.Time
	Finished   time.Time
	Err        error // nil on success
	RetryCount int   // number of retries actually performed (0 if first try succeeded)
}

// Duration returns how long the job took to execute end-to-end.
func (r *JobResult) Duration() time.Duration {
	if r == nil {
		return 0
	}
	return r.Finished.Sub(r.Started)
}

// WithMisfire sets the misfire policy for missed scheduled runs.
//
// The policy is recorded on the job's config; the scheduler implementation
// is responsible for honoring it when catching up. Default is
// MisfireFireOnce (the safest behavior for most jobs).
func WithMisfire(p MisfirePolicy) JobOption {
	return func(c *jobConfig) {
		c.misfire = p
		c.misfireSet = true
	}
}

// WithResult attaches a single callback that receives full execution detail
// (start/finish times, error, retry count) for every run. Pairs well with —
// or replaces — WithOnSuccess / WithOnError when you want timing data.
func WithResult(fn func(JobResult)) JobOption {
	return func(c *jobConfig) {
		c.onResult = fn
	}
}

// LeaderElector decides whether THIS scheduler instance is currently
// permitted to drive scheduled work. When the elector reports
// false, the scheduler should skip execution and (typically) rely on the
// existing lock-TTL primitive for safety.
//
// The chrono package ships a built-in always-leader implementation
// (NewAlwaysLeader) suitable for single-instance deployments and as a
// default. Multi-node deployments wire in a real elector backed by their
// coordination service (Postgres advisory locks, etcd, Consul, …).
type LeaderElector interface {
	// IsLeader returns true when this instance is currently the leader.
	// It MUST return quickly — typically a local atomic.Bool read after a
	// background lease loop maintains the state.
	IsLeader(ctx context.Context) bool
	// Resign voluntarily releases leadership; called on Scheduler.Stop().
	// May be a no-op for stateless electors.
	Resign(ctx context.Context) error
}

// alwaysLeader is the trivial elector that always claims leadership.
// Suitable for single-instance deployments. Multi-instance setups should
// supply a coordinated elector via WithLeaderElector.
type alwaysLeader struct{}

// NewAlwaysLeader returns a LeaderElector that always reports true.
func NewAlwaysLeader() LeaderElector { return &alwaysLeader{} }

func (alwaysLeader) IsLeader(context.Context) bool { return true }
func (alwaysLeader) Resign(context.Context) error  { return nil }

// MemoryLeader is a goroutine-safe leader elector whose state can be
// toggled at runtime — useful for tests and for app code that drives
// leadership externally (e.g. via a Kubernetes Lease watcher).
type MemoryLeader struct {
	mu     sync.RWMutex
	leader bool
}

// NewMemoryLeader returns a MemoryLeader in the given initial state.
func NewMemoryLeader(initial bool) *MemoryLeader {
	return &MemoryLeader{leader: initial}
}

// IsLeader returns the current state.
func (m *MemoryLeader) IsLeader(context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.leader
}

// Set updates the leadership state. Safe for concurrent use.
func (m *MemoryLeader) Set(v bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leader = v
}

// Resign clears the leader flag.
func (m *MemoryLeader) Resign(context.Context) error {
	m.Set(false)
	return nil
}

// WithLeaderElector attaches a LeaderElector to the scheduler. When the
// elector reports false, due-job dispatching is skipped (existing lock-TTL
// behavior remains as the safety net).
//
// Default when this option is not provided: NewAlwaysLeader() (single-
// instance mode — the existing distributed-lock behavior is unchanged).
func WithLeaderElector(le LeaderElector) Option {
	return func(s *defaultScheduler) {
		if le != nil {
			s.leader = le
		}
	}
}
