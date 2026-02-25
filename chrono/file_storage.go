package chrono

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/fsutils"
)

// fileState is the top-level structure persisted to the file.
type fileState struct {
	Jobs  []*JobRecord `json:"jobs" xml:"jobs" yaml:"jobs"`
	Locks []*fileLock  `json:"locks,omitempty" xml:"locks,omitempty" yaml:"locks,omitempty"`
}

// fileLock is the serializable representation of a lock entry.
type fileLock struct {
	JobID   string    `json:"jobId" xml:"jobId" yaml:"jobId"`
	Owner   string    `json:"owner" xml:"owner" yaml:"owner"`
	Expires time.Time `json:"expires" xml:"expires" yaml:"expires"`
}

// FileStorage is a file-based implementation of the Storage interface.
// It persists all job records and lock state to a single file using golly's
// codec package. The serialization format (YAML, JSON, or XML) is automatically
// determined from the file extension using fsutils.LookupContentType.
//
// Supported extensions: .yaml, .yml, .json, .xml
//
// All reads and writes are serialized through a mutex to ensure consistency.
// The entire state is rewritten on each mutation (append-replace strategy).
type FileStorage struct {
	mu   sync.Mutex
	path string
	c    codec.Codec
}

// NewFileStorage creates a new FileStorage that persists state to the given file path.
// The serialization format is determined by the file extension using
// fsutils.LookupContentType. Supported content types are YAML, JSON, and XML.
//
// The directory is created if it does not exist. If the file already exists, its
// contents are loaded on first access; otherwise an empty state file is created.
func NewFileStorage(path string) (Storage, error) {
	contentType := fsutils.LookupContentType(path)

	c, err := codec.GetDefault(contentType)
	if err != nil {
		return nil, fmt.Errorf("chrono: unsupported file type %q for %s: %w", contentType, filepath.Base(path), err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	fs := &FileStorage{
		path: path,
		c:    c,
	}

	// If the file doesn't exist yet, write an empty state so subsequent
	// reads don't fail.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.DebugF("FileStorage: creating initial state file %s", path)
		if writeErr := fs.writeState(&fileState{}); writeErr != nil {
			logger.ErrorF("FileStorage: failed to create initial state file %s: %v", path, writeErr)
			return nil, writeErr
		}
	}

	logger.InfoF("FileStorage: initialized with path=%s contentType=%s", path, contentType)
	return fs, nil
}

// readState loads the full state from the file.
func (fs *FileStorage) readState() (*fileState, error) {
	f, err := os.Open(fs.path)
	if err != nil {
		logger.ErrorF("FileStorage: failed to open state file %s: %v", fs.path, err)
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var state fileState
	if err := fs.c.Read(f, &state); err != nil {
		logger.ErrorF("FileStorage: failed to decode state file %s: %v", fs.path, err)
		return nil, err
	}
	return &state, nil
}

// writeState persists the full state to the file atomically.
// It writes to a temp file first, then renames to prevent corruption.
func (fs *FileStorage) writeState(state *fileState) error {
	tmp := fs.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		logger.ErrorF("FileStorage: failed to create temp file %s: %v", tmp, err)
		return err
	}

	if writeErr := fs.c.Write(state, f); writeErr != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		logger.ErrorF("FileStorage: failed to encode state to %s: %v", tmp, writeErr)
		return writeErr
	}
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}

	return os.Rename(tmp, fs.path)
}

// findJob returns the index and pointer to the job with the given ID, or -1, nil.
func (fs *FileStorage) findJob(state *fileState, id string) (int, *JobRecord) {
	for i, rec := range state.Jobs {
		if rec.ID == id {
			return i, rec
		}
	}
	return -1, nil
}

// findLock returns the index and pointer to the lock for the given jobID, or -1, nil.
func (fs *FileStorage) findLock(state *fileState, jobID string) (int, *fileLock) {
	for i, lk := range state.Locks {
		if lk.JobID == jobID {
			return i, lk
		}
	}
	return -1, nil
}

// SaveJob persists a job record (upsert).
func (fs *FileStorage) SaveJob(_ context.Context, record *JobRecord) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return err
	}

	cp := *record
	if idx, _ := fs.findJob(state, record.ID); idx >= 0 {
		state.Jobs[idx] = &cp
	} else {
		state.Jobs = append(state.Jobs, &cp)
	}

	return fs.writeState(state)
}

// GetJob retrieves a job record by ID.
func (fs *FileStorage) GetJob(_ context.Context, id string) (*JobRecord, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return nil, err
	}

	_, rec := fs.findJob(state, id)
	if rec == nil {
		return nil, ErrJobNotFound
	}

	cp := *rec
	return &cp, nil
}

// DeleteJob removes a job record by ID.
func (fs *FileStorage) DeleteJob(_ context.Context, id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return err
	}

	idx, _ := fs.findJob(state, id)
	if idx < 0 {
		return ErrJobNotFound
	}

	// Remove job
	state.Jobs = append(state.Jobs[:idx], state.Jobs[idx+1:]...)

	// Remove associated lock if any
	if li, _ := fs.findLock(state, id); li >= 0 {
		state.Locks = append(state.Locks[:li], state.Locks[li+1:]...)
	}

	return fs.writeState(state)
}

// ListJobs returns all stored job records.
func (fs *FileStorage) ListJobs(_ context.Context) ([]*JobRecord, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return nil, err
	}

	records := make([]*JobRecord, len(state.Jobs))
	for i, rec := range state.Jobs {
		cp := *rec
		records[i] = &cp
	}
	return records, nil
}

// GetDueJobs returns jobs that are due for execution.
func (fs *FileStorage) GetDueJobs(_ context.Context, now time.Time) ([]*JobRecord, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return nil, err
	}

	var due []*JobRecord
	for _, rec := range state.Jobs {
		if rec.Paused || rec.NextRun.IsZero() || now.Before(rec.NextRun) {
			continue
		}
		cp := *rec
		due = append(due, &cp)
	}
	return due, nil
}

// AcquireLock attempts to acquire an execution lock for a job.
func (fs *FileStorage) AcquireLock(_ context.Context, jobID string, ownerID string, ttl time.Duration) (bool, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return false, err
	}

	now := time.Now()
	idx, lk := fs.findLock(state, jobID)

	if lk != nil {
		// Lock held by another owner and not yet expired
		if lk.Owner != ownerID && now.Before(lk.Expires) {
			return false, nil
		}
		// Update existing lock
		state.Locks[idx] = &fileLock{
			JobID:   jobID,
			Owner:   ownerID,
			Expires: now.Add(ttl),
		}
	} else {
		state.Locks = append(state.Locks, &fileLock{
			JobID:   jobID,
			Owner:   ownerID,
			Expires: now.Add(ttl),
		})
	}

	return true, fs.writeState(state)
}

// ReleaseLock releases the execution lock for a job.
func (fs *FileStorage) ReleaseLock(_ context.Context, jobID string, ownerID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state, err := fs.readState()
	if err != nil {
		return err
	}

	idx, lk := fs.findLock(state, jobID)
	if lk == nil {
		return nil
	}

	// Only the owner can release the lock
	if lk.Owner == ownerID {
		state.Locks = append(state.Locks[:idx], state.Locks[idx+1:]...)
		return fs.writeState(state)
	}

	return nil
}

// Close is a no-op for file storage — the file is opened and closed on each operation.
func (fs *FileStorage) Close() error {
	return nil
}
