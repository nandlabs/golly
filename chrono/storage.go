package chrono

import (
	"context"
	"time"
)

// Storage defines the interface for persisting scheduler state and coordinating
// job execution across instances. Implementations can provide in-memory,
// file-based, or database-backed storage for single-instance or multi-instance
// (clustered) deployments.
//
// For single-instance usage, use NewInMemoryStorage(). For multi-instance clusters,
// implement this interface backed by a shared data store (e.g., PostgreSQL, Redis)
// to enable distributed locking and state synchronization.
type Storage interface {
	// SaveJob persists a job record. If a record with the same ID already exists,
	// it is updated (upsert). Implementations should handle concurrent access safely.
	SaveJob(ctx context.Context, record *JobRecord) error

	// GetJob retrieves a job record by its unique ID.
	// Returns ErrJobNotFound if no record with the given ID exists.
	GetJob(ctx context.Context, id string) (*JobRecord, error)

	// DeleteJob removes a job record by its unique ID.
	// Returns ErrJobNotFound if no record with the given ID exists.
	DeleteJob(ctx context.Context, id string) error

	// ListJobs returns all stored job records.
	ListJobs(ctx context.Context) ([]*JobRecord, error)

	// GetDueJobs returns job records that are due for execution.
	// A job is due when its NextRun is at or before now, it is not paused,
	// and its NextRun is not zero (i.e., not a completed one-shot job).
	// Storage implementations can optimize this with efficient queries/indexes.
	GetDueJobs(ctx context.Context, now time.Time) ([]*JobRecord, error)

	// AcquireLock attempts to acquire an execution lock for the given job.
	// The ownerID identifies the scheduler instance attempting to acquire the lock.
	// The lock should auto-expire after the specified TTL to handle crashed instances.
	//
	// Returns true if the lock was successfully acquired. Returns false if another
	// owner holds a non-expired lock. Re-acquiring a lock by the same owner extends it.
	//
	// For in-memory storage (single instance), this always succeeds.
	// For distributed storage, this provides mutual exclusion across instances.
	AcquireLock(ctx context.Context, jobID string, ownerID string, ttl time.Duration) (bool, error)

	// ReleaseLock releases the execution lock for the given job.
	// Only the owner that acquired the lock can release it.
	ReleaseLock(ctx context.Context, jobID string, ownerID string) error

	// Close releases any resources held by the storage (connections, file handles, etc.).
	Close() error
}

// JobRecord is the serializable representation of a job's metadata and execution state.
// This is what gets persisted in Storage. The actual JobFunc is not stored — it is
// registered locally on each scheduler instance via AddJob/AddCronJob/etc.
type JobRecord struct {
	// ID is the unique identifier of the job.
	ID string `json:"id" xml:"id" yaml:"id"`
	// Name is the human-readable name of the job.
	Name string `json:"name" xml:"name" yaml:"name"`
	// Status is the current execution status.
	Status JobStatus `json:"status" xml:"status" yaml:"status"`
	// Paused indicates whether the job is paused.
	Paused bool `json:"paused" xml:"paused" yaml:"paused"`
	// LastRun is the time the job was last executed.
	LastRun time.Time `json:"lastRun" xml:"lastRun" yaml:"lastRun"`
	// NextRun is the scheduled time for the next execution.
	NextRun time.Time `json:"nextRun" xml:"nextRun" yaml:"nextRun"`
	// RunCount is the total number of times the job has been executed.
	RunCount int64 `json:"runCount" xml:"runCount" yaml:"runCount"`
	// ErrorCount is the total number of failed executions.
	ErrorCount int64 `json:"errorCount" xml:"errorCount" yaml:"errorCount"`
	// LastError is the error message from the most recent failed execution.
	LastError string `json:"lastError,omitempty" xml:"lastError,omitempty" yaml:"lastError,omitempty"`
}
