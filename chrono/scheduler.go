package chrono

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// Error sentinels for common scheduler errors.
var (
	// ErrSchedulerRunning is returned when attempting to start an already running scheduler.
	ErrSchedulerRunning = errors.New("chrono: already running")
	// ErrSchedulerStopped is returned when attempting to operate on a stopped scheduler.
	ErrSchedulerStopped = errors.New("chrono: not running")
	// ErrJobNotFound is returned when a job with the given ID does not exist.
	ErrJobNotFound = errors.New("chrono: job not found")
	// ErrJobAlreadyExists is returned when a job with the given ID already exists.
	ErrJobAlreadyExists = errors.New("chrono: job already exists")
	// ErrInvalidCronExpr is returned when a cron expression is malformed.
	ErrInvalidCronExpr = errors.New("chrono: invalid cron expression")
	// ErrInvalidInterval is returned when an interval duration is invalid.
	ErrInvalidInterval = errors.New("chrono: invalid interval")
	// ErrInvalidDelay is returned when a delay duration is invalid.
	ErrInvalidDelay = errors.New("chrono: invalid delay")
	// ErrNilJobFunc is returned when a nil function is provided.
	ErrNilJobFunc = errors.New("chrono: job function cannot be nil")
	// ErrEmptyJobID is returned when an empty job ID is provided.
	ErrEmptyJobID = errors.New("chrono: job ID cannot be empty")
)

// JobFunc is the function signature for scheduled jobs.
// The context is canceled when the scheduler is stopped or when the job times out.
type JobFunc func(ctx context.Context) error

// JobStatus represents the current execution status of a job.
type JobStatus int

const (
	// JobStatusPending indicates the job is waiting to be executed.
	JobStatusPending JobStatus = iota
	// JobStatusRunning indicates the job is currently executing.
	JobStatusRunning
	// JobStatusCompleted indicates the job has completed its last execution successfully.
	JobStatusCompleted
	// JobStatusFailed indicates the job has failed its last execution.
	JobStatusFailed
	// JobStatusCancelled indicates the job has been canceled.
	JobStatusCancelled
)

// String returns the string representation of a JobStatus.
func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusRunning:
		return "running"
	case JobStatusCompleted:
		return "completed"
	case JobStatusFailed:
		return "failed"
	case JobStatusCancelled:
		return "canceled"
	default:
		return "unknown"
	}
}

// Schedule defines when a job should be executed.
type Schedule interface {
	// Next returns the next activation time after the given time.
	// It returns the zero time if there are no more activations.
	Next(from time.Time) time.Time
}

// JobOption is a functional option for configuring a job.
type JobOption func(*jobConfig)

// jobConfig holds optional job configuration.
type jobConfig struct {
	maxRetries int
	timeout    time.Duration
	onSuccess  func(jobID string)
	onError    func(jobID string, err error)
}

// WithMaxRetries sets the maximum number of retries for a failed job execution.
// If a job fails, it will be retried up to n times before being marked as failed.
func WithMaxRetries(n int) JobOption {
	return func(c *jobConfig) {
		if n > 0 {
			c.maxRetries = n
		}
	}
}

// WithTimeout sets the maximum execution time for a single job run.
// If the job does not complete within the timeout, its context is canceled.
func WithTimeout(d time.Duration) JobOption {
	return func(c *jobConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithOnSuccess sets a callback function that is invoked when the job completes successfully.
func WithOnSuccess(fn func(jobID string)) JobOption {
	return func(c *jobConfig) {
		c.onSuccess = fn
	}
}

// WithOnError sets a callback function that is invoked when the job fails.
func WithOnError(fn func(jobID string, err error)) JobOption {
	return func(c *jobConfig) {
		c.onError = fn
	}
}

// JobInfo provides read-only information about a scheduled job.
type JobInfo struct {
	// ID is the unique identifier of the job.
	ID string
	// Name is the human-readable name of the job.
	Name string
	// Status is the current execution status.
	Status JobStatus
	// LastRun is the time the job was last executed.
	LastRun time.Time
	// NextRun is the scheduled time for the next execution.
	NextRun time.Time
	// RunCount is the total number of times the job has been executed.
	RunCount int64
	// ErrorCount is the total number of failed executions.
	ErrorCount int64
	// LastError is the error from the most recent failed execution.
	LastError error
}

// Scheduler manages the scheduling and execution of jobs.
type Scheduler interface {
	// AddJob adds a job with the given schedule.
	AddJob(id, name string, fn JobFunc, schedule Schedule, opts ...JobOption) error
	// AddCronJob adds a job that runs on a cron schedule.
	// The cron expression must be a standard 5-field expression:
	// minute (0-59), hour (0-23), day-of-month (1-31), month (1-12), day-of-week (0-6, 0=Sunday).
	// Predefined macros (@yearly, @monthly, @weekly, @daily, @hourly) are also supported.
	AddCronJob(id, name string, fn JobFunc, cronExpr string, opts ...JobOption) error
	// AddIntervalJob adds a job that runs at a fixed interval.
	AddIntervalJob(id, name string, fn JobFunc, interval time.Duration, opts ...JobOption) error
	// AddOneShotJob adds a job that runs once after the specified delay.
	AddOneShotJob(id, name string, fn JobFunc, delay time.Duration, opts ...JobOption) error
	// RemoveJob removes a scheduled job by ID.
	RemoveJob(id string) error
	// PauseJob pauses a scheduled job. The job will not be executed until resumed.
	PauseJob(id string) error
	// ResumeJob resumes a paused job.
	ResumeJob(id string) error
	// GetJob returns information about a scheduled job.
	GetJob(id string) (*JobInfo, error)
	// ListJobs returns information about all scheduled jobs.
	ListJobs() []*JobInfo
	// Start starts the scheduler. It begins monitoring and executing due jobs.
	Start() error
	// Stop stops the scheduler gracefully, waiting for running jobs to complete.
	Stop() error
	// IsRunning returns true if the scheduler is currently running.
	IsRunning() bool
}

// Option is a functional option for configuring the scheduler itself.
type Option func(*defaultScheduler)

// WithCheckInterval sets the interval at which the scheduler checks for due jobs.
// This value is used as the storage poll interval for the background ticker that
// discovers jobs added by other scheduler instances. The default is 1 second.
// For precise local scheduling, the scheduler uses an event-driven timer that
// wakes exactly when the next job is due — this interval only affects distributed
// change detection.
func WithCheckInterval(d time.Duration) Option {
	return func(s *defaultScheduler) {
		if d > 0 {
			s.checkInterval = d
		}
	}
}

// WithStoragePollInterval sets the interval at which the scheduler polls the
// storage backend for changes made by other scheduler instances (e.g., new jobs,
// removed jobs, resumed jobs). This is the "slow" poll in the hybrid approach.
// The precise timer handles locally-known schedules with zero-latency wake-ups;
// this poll catches external mutations. The default is 30 seconds.
func WithStoragePollInterval(d time.Duration) Option {
	return func(s *defaultScheduler) {
		if d > 0 {
			s.storagePollInterval = d
		}
	}
}

// WithStorage sets the storage backend for the scheduler.
// If not set, NewInMemoryStorage() is used by default.
// Use a shared storage implementation (e.g., database-backed) for multi-instance
// cluster deployments where jobs must be coordinated across instances.
func WithStorage(store Storage) Option {
	return func(s *defaultScheduler) {
		s.storage = store
	}
}

// WithInstanceID sets a unique identifier for this scheduler instance.
// This ID is used for distributed lock ownership. In a cluster, each instance
// must have a unique ID. If not set, a default is generated from hostname and PID.
func WithInstanceID(id string) Option {
	return func(s *defaultScheduler) {
		if id != "" {
			s.instanceID = id
		}
	}
}

// WithLockTTL sets the time-to-live for job execution locks.
// If a scheduler instance crashes while holding a lock, the lock will expire
// after this duration, allowing another instance to pick up the job.
// The default is 5 minutes. Set this to a value longer than your longest-running job.
func WithLockTTL(d time.Duration) Option {
	return func(s *defaultScheduler) {
		if d > 0 {
			s.lockTTL = d
		}
	}
}

// New creates a new Scheduler with default settings.
// By default, uses in-memory storage. Use WithStorage() to provide a
// persistent or distributed storage backend.
func New(opts ...Option) Scheduler {
	s := &defaultScheduler{
		entries:             make(map[string]*jobEntry),
		checkInterval:       time.Second,
		storagePollInterval: 30 * time.Second,
		lockTTL:             5 * time.Minute,
		instanceID:          defaultInstanceID(),
		wake:                make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(s)
	}
	// Backward compatibility: if checkInterval was customized but storagePollInterval
	// was not explicitly set via WithStoragePollInterval, use checkInterval as the
	// storage poll interval. This preserves behavior for existing users of
	// WithCheckInterval.
	if s.storagePollInterval == 30*time.Second && s.checkInterval != time.Second {
		s.storagePollInterval = s.checkInterval
	}
	if s.storage == nil {
		s.storage = NewInMemoryStorage()
	}
	logger.InfoF("Scheduler created (instance=%s, storagePollInterval=%s, lockTTL=%s)", s.instanceID, s.storagePollInterval, s.lockTTL)
	return s
}

// defaultInstanceID generates a unique ID for this scheduler instance.
func defaultInstanceID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%d-%d", hostname, os.Getpid(), time.Now().UnixNano())
}

// recordToInfo converts a JobRecord from storage into a JobInfo for the public API.
func recordToInfo(rec *JobRecord) *JobInfo {
	var lastErr error
	if rec.LastError != "" {
		lastErr = errors.New(rec.LastError)
	}
	return &JobInfo{
		ID:         rec.ID,
		Name:       rec.Name,
		Status:     rec.Status,
		LastRun:    rec.LastRun,
		NextRun:    rec.NextRun,
		RunCount:   rec.RunCount,
		ErrorCount: rec.ErrorCount,
		LastError:  lastErr,
	}
}
