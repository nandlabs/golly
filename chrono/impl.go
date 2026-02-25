package chrono

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// jobEntry holds the non-serializable parts of a job that stay local to this
// scheduler instance. The serializable state (status, counts, timestamps) is
// managed by Storage.
type jobEntry struct {
	fn       JobFunc
	schedule Schedule
	config   jobConfig
	running  int32 // atomic flag to prevent overlapping executions on this instance
}

// defaultScheduler is the default implementation of the Scheduler interface.
// It delegates state persistence and distributed locking to a Storage backend.
//
// Execution is driven by a precise timer that sleeps until the next job is due,
// combined with a slower background poll to pick up changes made by other scheduler
// instances sharing the same Storage (hybrid approach).
type defaultScheduler struct {
	mu                  sync.RWMutex
	entries             map[string]*jobEntry // local function + schedule registry
	storage             Storage
	running             bool
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	checkInterval       time.Duration // kept for backward compatibility (used as storagePollInterval default)
	storagePollInterval time.Duration // how often to poll storage for changes from other instances
	lockTTL             time.Duration
	instanceID          string
	wake                chan struct{} // signal to recalculate timer when jobs are added/removed/resumed
}

// AddJob adds a job with the given schedule.
func (s *defaultScheduler) AddJob(id, name string, fn JobFunc, schedule Schedule, opts ...JobOption) error {
	if id == "" {
		return ErrEmptyJobID
	}
	if fn == nil {
		return ErrNilJobFunc
	}

	cfg := jobConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already registered locally
	if _, exists := s.entries[id]; exists {
		logger.WarnF("AddJob: job %q already exists", id)
		return ErrJobAlreadyExists
	}

	// Check if record exists in storage (e.g., from a previous run or another instance)
	_, err := s.storage.GetJob(context.Background(), id)
	if err != nil {
		if !errors.Is(err, ErrJobNotFound) {
			logger.ErrorF("AddJob: storage error for job %q: %v", id, err)
			return err
		}
		// Not found in storage — create a new record
		rec := &JobRecord{
			ID:      id,
			Name:    name,
			Status:  JobStatusPending,
			NextRun: schedule.Next(time.Now()),
		}
		if saveErr := s.storage.SaveJob(context.Background(), rec); saveErr != nil {
			logger.ErrorF("AddJob: failed to save job %q to storage: %v", id, saveErr)
			return saveErr
		}
	} else {
		logger.DebugF("AddJob: job %q already exists in storage, registering function locally", id)
	}

	s.entries[id] = &jobEntry{
		fn:       fn,
		schedule: schedule,
		config:   cfg,
	}
	logger.InfoF("AddJob: registered job %q (%s)", id, name)
	s.signalWake()
	return nil
}

// AddCronJob adds a job that runs on a cron schedule.
func (s *defaultScheduler) AddCronJob(id, name string, fn JobFunc, cronExpr string, opts ...JobOption) error {
	sched, err := NewCronSchedule(cronExpr)
	if err != nil {
		return err
	}
	return s.AddJob(id, name, fn, sched, opts...)
}

// AddIntervalJob adds a job that runs at a fixed interval.
func (s *defaultScheduler) AddIntervalJob(id, name string, fn JobFunc, interval time.Duration, opts ...JobOption) error {
	sched, err := NewIntervalSchedule(interval)
	if err != nil {
		return err
	}
	return s.AddJob(id, name, fn, sched, opts...)
}

// AddOneShotJob adds a job that runs once after the specified delay.
func (s *defaultScheduler) AddOneShotJob(id, name string, fn JobFunc, delay time.Duration, opts ...JobOption) error {
	sched, err := NewOneShotSchedule(delay)
	if err != nil {
		return err
	}
	return s.AddJob(id, name, fn, sched, opts...)
}

// RemoveJob removes a scheduled job by ID.
func (s *defaultScheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[id]; !exists {
		return ErrJobNotFound
	}

	// Delete from storage
	if err := s.storage.DeleteJob(context.Background(), id); err != nil && !errors.Is(err, ErrJobNotFound) {
		logger.ErrorF("RemoveJob: failed to delete job %q from storage: %v", id, err)
		return err
	}

	delete(s.entries, id)
	logger.InfoF("RemoveJob: removed job %q", id)
	s.signalWake()
	return nil
}

// PauseJob pauses a scheduled job.
func (s *defaultScheduler) PauseJob(id string) error {
	s.mu.RLock()
	_, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	rec, err := s.storage.GetJob(context.Background(), id)
	if err != nil {
		return err
	}

	rec.Paused = true
	if err = s.storage.SaveJob(context.Background(), rec); err != nil {
		logger.ErrorF("PauseJob: failed to save paused state for job %q: %v", id, err)
		return err
	}
	logger.InfoF("PauseJob: paused job %q", id)
	return nil
}

// ResumeJob resumes a paused job.
func (s *defaultScheduler) ResumeJob(id string) error {
	s.mu.RLock()
	entry, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	rec, err := s.storage.GetJob(context.Background(), id)
	if err != nil {
		return err
	}

	rec.Paused = false
	// Recompute next run time from now using the local schedule
	rec.NextRun = entry.schedule.Next(time.Now())
	if err = s.storage.SaveJob(context.Background(), rec); err != nil {
		logger.ErrorF("ResumeJob: failed to save resumed state for job %q: %v", id, err)
		return err
	}
	logger.InfoF("ResumeJob: resumed job %q, next run at %s", id, rec.NextRun)
	s.signalWake()
	return nil
}

// GetJob returns information about a scheduled job.
func (s *defaultScheduler) GetJob(id string) (*JobInfo, error) {
	rec, err := s.storage.GetJob(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return recordToInfo(rec), nil
}

// ListJobs returns information about all scheduled jobs.
func (s *defaultScheduler) ListJobs() []*JobInfo {
	records, err := s.storage.ListJobs(context.Background())
	if err != nil {
		return nil
	}

	infos := make([]*JobInfo, 0, len(records))
	for _, rec := range records {
		infos = append(infos, recordToInfo(rec))
	}
	return infos
}

// Start starts the scheduler.
func (s *defaultScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrSchedulerRunning
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true

	s.wg.Add(1)
	go s.run()

	logger.InfoF("Scheduler started (instance=%s, storagePollInterval=%s, lockTTL=%s)", s.instanceID, s.storagePollInterval, s.lockTTL)
	return nil
}

// Stop stops the scheduler gracefully, waiting for running jobs to complete.
func (s *defaultScheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return ErrSchedulerStopped
	}
	s.cancel()
	s.running = false
	s.mu.Unlock()

	logger.Info("Scheduler stopping, waiting for running jobs to complete...")
	// Wait for the run loop and all job goroutines to finish
	s.wg.Wait()
	logger.Info("Scheduler stopped")
	return nil
}

// IsRunning returns true if the scheduler is currently running.
func (s *defaultScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// signalWake sends a non-blocking signal to the run loop to recalculate the
// next wake time. This is called after any mutation that may affect scheduling
// (AddJob, RemoveJob, ResumeJob).
func (s *defaultScheduler) signalWake() {
	select {
	case s.wake <- struct{}{}:
	default:
		// Already signaled; the run loop will pick it up.
	}
}

// nextWakeDuration computes how long to sleep until the earliest due job.
// It scans all local entries' schedules and returns the shortest duration.
// The caller must hold at least s.mu.RLock.
func (s *defaultScheduler) nextWakeDuration() time.Duration {
	now := time.Now()
	var earliest time.Time

	for _, entry := range s.entries {
		next := entry.schedule.Next(now)
		if next.IsZero() {
			continue
		}
		if earliest.IsZero() || next.Before(earliest) {
			earliest = next
		}
	}

	if earliest.IsZero() {
		// No jobs or all one-shot jobs already fired — sleep until the next
		// storage poll catches something.
		return s.storagePollInterval
	}

	d := earliest.Sub(now)
	if d <= 0 {
		return 0 // already due
	}
	return d
}

// run is the main scheduler loop. It uses a precise timer that wakes exactly
// when the next local job is due, combined with a background ticker that polls
// storage periodically to discover changes from other scheduler instances.
func (s *defaultScheduler) run() {
	defer s.wg.Done()

	// Storage poll ticker — slower cadence, catches external mutations
	storageTicker := time.NewTicker(s.storagePollInterval)
	defer storageTicker.Stop()

	// Precise timer — fires at the exact next-job time
	s.mu.RLock()
	d := s.nextWakeDuration()
	s.mu.RUnlock()
	timer := time.NewTimer(d)
	defer timer.Stop()

	resetTimer := func() {
		s.mu.RLock()
		next := s.nextWakeDuration()
		s.mu.RUnlock()
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(next)
	}

	for {
		select {
		case <-s.ctx.Done():
			return

		case now := <-timer.C:
			// Precise wake — a local job is due
			s.checkAndExecute(now)
			resetTimer()

		case now := <-storageTicker.C:
			// Background poll — catch jobs added by other instances
			s.checkAndExecute(now)
			resetTimer()

		case <-s.wake:
			// A local mutation happened — recalculate the timer without
			// executing anything; the timer will fire when due.
			logger.DebugF("run: wake signal received, recalculating timer")
			resetTimer()
		}
	}
}

// checkAndExecute checks storage for due jobs and executes them.
func (s *defaultScheduler) checkAndExecute(now time.Time) {
	// Query storage for due jobs
	records, err := s.storage.GetDueJobs(s.ctx, now)
	if err != nil {
		logger.ErrorF("checkAndExecute: failed to query due jobs: %v", err)
		return
	}
	if len(records) > 0 {
		logger.DebugF("checkAndExecute: found %d due job(s)", len(records))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rec := range records {
		// Check if we have the function registered locally
		entry, exists := s.entries[rec.ID]
		if !exists {
			continue
		}

		// Try to acquire the distributed lock
		locked, lockErr := s.storage.AcquireLock(s.ctx, rec.ID, s.instanceID, s.lockTTL)
		if lockErr != nil {
			logger.ErrorF("checkAndExecute: failed to acquire lock for job %q: %v", rec.ID, lockErr)
			continue
		}
		if !locked {
			logger.DebugF("checkAndExecute: lock not acquired for job %q (held by another instance)", rec.ID)
			continue
		}

		// Prevent overlapping execution on this instance
		if !atomic.CompareAndSwapInt32(&entry.running, 0, 1) {
			logger.DebugF("checkAndExecute: job %q already running on this instance, skipping", rec.ID)
			_ = s.storage.ReleaseLock(s.ctx, rec.ID, s.instanceID)
			continue
		}

		// Mark as running in storage
		rec.Status = JobStatusRunning
		_ = s.storage.SaveJob(s.ctx, rec)

		logger.DebugF("checkAndExecute: executing job %q", rec.ID)
		// Execute the job in a goroutine
		s.wg.Add(1)
		go s.executeJob(entry, rec)
	}
}

// executeJob executes a single job with retry and timeout support.
// It updates the job state in storage and releases the lock when done.
func (s *defaultScheduler) executeJob(entry *jobEntry, rec *JobRecord) {
	defer s.wg.Done()
	defer atomic.StoreInt32(&entry.running, 0)
	defer func() {
		_ = s.storage.ReleaseLock(context.Background(), rec.ID, s.instanceID)
	}()

	var jobErr error
	maxAttempts := 1 + entry.config.maxRetries

	for range maxAttempts {
		// Create job context (with optional timeout)
		var jobCtx context.Context
		var jobCancel context.CancelFunc

		if entry.config.timeout > 0 {
			jobCtx, jobCancel = context.WithTimeout(s.ctx, entry.config.timeout)
		} else {
			jobCtx, jobCancel = context.WithCancel(s.ctx)
		}

		jobErr = entry.fn(jobCtx)
		jobCancel()

		if jobErr == nil {
			break
		}

		// Check if context was canceled (scheduler stopping)
		select {
		case <-s.ctx.Done():
			logger.WarnF("executeJob: job %q canceled due to scheduler shutdown", rec.ID)
			rec.Status = JobStatusCancelled
			rec.LastRun = time.Now()
			_ = s.storage.SaveJob(context.Background(), rec)
			return
		default:
			logger.DebugF("executeJob: job %q failed (attempt), will retry: %v", rec.ID, jobErr)
		}
	}

	now := time.Now()
	rec.LastRun = now
	rec.RunCount++

	if jobErr != nil {
		rec.Status = JobStatusFailed
		rec.ErrorCount++
		rec.LastError = jobErr.Error()
		logger.ErrorF("executeJob: job %q failed after %d attempt(s): %v", rec.ID, maxAttempts, jobErr)
	} else {
		rec.Status = JobStatusCompleted
		rec.LastError = ""
		logger.DebugF("executeJob: job %q completed successfully (run #%d)", rec.ID, rec.RunCount)
	}

	// Compute next run time using the local schedule
	rec.NextRun = entry.schedule.Next(now)

	// Persist updated state to storage
	_ = s.storage.SaveJob(context.Background(), rec)

	// Invoke callbacks (outside storage operations)
	if jobErr != nil {
		if entry.config.onError != nil {
			entry.config.onError(rec.ID, jobErr)
		}
	} else {
		if entry.config.onSuccess != nil {
			entry.config.onSuccess(rec.ID)
		}
	}
}
