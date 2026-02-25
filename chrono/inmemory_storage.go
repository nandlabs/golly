package chrono

import (
	"context"
	"sync"
	"time"
)

// lockEntry represents a distributed lock held by a scheduler instance.
type lockEntry struct {
	owner   string
	expires time.Time
}

// InMemoryStorage is an in-memory implementation of the Storage interface.
// It is suitable for single-instance deployments where persistence across
// restarts is not required.
type InMemoryStorage struct {
	mu    sync.RWMutex
	jobs  map[string]*JobRecord
	locks map[string]*lockEntry
}

// NewInMemoryStorage creates a new InMemoryStorage instance.
func NewInMemoryStorage() Storage {
	return &InMemoryStorage{
		jobs:  make(map[string]*JobRecord),
		locks: make(map[string]*lockEntry),
	}
}

// SaveJob persists a job record in memory (upsert).
func (m *InMemoryStorage) SaveJob(_ context.Context, record *JobRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy to prevent external mutation
	cp := *record
	m.jobs[record.ID] = &cp
	return nil
}

// GetJob retrieves a job record by ID.
func (m *InMemoryStorage) GetJob(_ context.Context, id string) (*JobRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, exists := m.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}

	// Return a copy to prevent external mutation
	cp := *rec
	return &cp, nil
}

// DeleteJob removes a job record by ID.
func (m *InMemoryStorage) DeleteJob(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.jobs[id]; !exists {
		return ErrJobNotFound
	}

	delete(m.jobs, id)
	delete(m.locks, id)
	return nil
}

// ListJobs returns all stored job records.
func (m *InMemoryStorage) ListJobs(_ context.Context) ([]*JobRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]*JobRecord, 0, len(m.jobs))
	for _, rec := range m.jobs {
		cp := *rec
		records = append(records, &cp)
	}
	return records, nil
}

// GetDueJobs returns jobs that are due for execution.
func (m *InMemoryStorage) GetDueJobs(_ context.Context, now time.Time) ([]*JobRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var due []*JobRecord
	for _, rec := range m.jobs {
		if rec.Paused {
			continue
		}
		if rec.NextRun.IsZero() {
			continue
		}
		if now.Before(rec.NextRun) {
			continue
		}
		cp := *rec
		due = append(due, &cp)
	}
	return due, nil
}

// AcquireLock attempts to acquire an execution lock for a job.
// For in-memory storage, this implements per-job mutual exclusion using
// owner IDs and expiration times.
func (m *InMemoryStorage) AcquireLock(_ context.Context, jobID string, ownerID string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	if lock, exists := m.locks[jobID]; exists {
		// Lock held by another owner and not yet expired
		if lock.owner != ownerID && now.Before(lock.expires) {
			return false, nil
		}
	}

	// Acquire or extend the lock
	m.locks[jobID] = &lockEntry{
		owner:   ownerID,
		expires: now.Add(ttl),
	}
	return true, nil
}

// ReleaseLock releases the execution lock for a job.
func (m *InMemoryStorage) ReleaseLock(_ context.Context, jobID string, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[jobID]
	if !exists {
		return nil
	}

	// Only the owner can release the lock
	if lock.owner == ownerID {
		delete(m.locks, jobID)
	}

	return nil
}

// Close is a no-op for in-memory storage.
func (m *InMemoryStorage) Close() error {
	return nil
}
