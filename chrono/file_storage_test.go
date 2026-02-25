package chrono

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testFormats defines extensions to exercise for every FileStorage test.
var testFormats = []string{".yaml", ".json", ".xml"}

func tempFilePathExt(t *testing.T, ext string) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "chrono"+ext)
}

func tempFilePath(t *testing.T) string {
	return tempFilePathExt(t, ".yaml")
}

// runForAllFormats runs the given sub-test function once per format.
func runForAllFormats(t *testing.T, fn func(t *testing.T, ext string)) {
	t.Helper()
	for _, ext := range testFormats {
		t.Run(ext, func(t *testing.T) {
			fn(t, ext)
		})
	}
}

func TestNewFileStorage_CreatesFile(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		path := tempFilePathExt(t, ext)
		fs, err := NewFileStorage(path)
		if err != nil {
			t.Fatalf("NewFileStorage error: %v", err)
		}
		defer fs.Close()

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("expected file to be created")
		}
	})
}

func TestNewFileStorage_CreatesDir(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sub", "deep", "chrono"+ext)
		fs, err := NewFileStorage(path)
		if err != nil {
			t.Fatalf("NewFileStorage error: %v", err)
		}
		defer fs.Close()

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("expected file to be created in nested dir")
		}
	})
}

func TestNewFileStorage_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chrono.csv")
	_, err := NewFileStorage(path)
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewFileStorage_YmlExtension(t *testing.T) {
	path := tempFilePathExt(t, ".yml")
	fs, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage error for .yml: %v", err)
	}
	defer fs.Close()
	ctx := context.Background()

	_ = fs.SaveJob(ctx, &JobRecord{ID: "ymltest", Name: "YML Test"})
	got, err := fs.GetJob(ctx, "ymltest")
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	if got.Name != "YML Test" {
		t.Fatalf("expected 'YML Test', got %q", got.Name)
	}
}

func TestFileStorage_SaveAndGetJob(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		rec := &JobRecord{
			ID:      "job1",
			Name:    "Test Job",
			Status:  JobStatusPending,
			NextRun: time.Now().Add(time.Hour).Truncate(time.Second),
		}

		if err := fs.SaveJob(ctx, rec); err != nil {
			t.Fatalf("SaveJob error: %v", err)
		}

		got, err := fs.GetJob(ctx, "job1")
		if err != nil {
			t.Fatalf("GetJob error: %v", err)
		}
		if got.ID != "job1" || got.Name != "Test Job" || got.Status != JobStatusPending {
			t.Fatalf("unexpected job: %+v", got)
		}
	})
}

func TestFileStorage_GetJob_NotFound(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()

		_, err := fs.GetJob(context.Background(), "nonexistent")
		if err != ErrJobNotFound {
			t.Fatalf("expected ErrJobNotFound, got: %v", err)
		}
	})
}

func TestFileStorage_SaveJob_Upsert(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		rec := &JobRecord{ID: "job1", Name: "V1", Status: JobStatusPending}
		_ = fs.SaveJob(ctx, rec)

		rec.Name = "V2"
		rec.Status = JobStatusCompleted
		rec.RunCount = 5
		_ = fs.SaveJob(ctx, rec)

		got, _ := fs.GetJob(ctx, "job1")
		if got.Name != "V2" || got.RunCount != 5 || got.Status != JobStatusCompleted {
			t.Fatalf("upsert failed: %+v", got)
		}

		jobs, _ := fs.ListJobs(ctx)
		if len(jobs) != 1 {
			t.Fatalf("expected 1 job after upsert, got %d", len(jobs))
		}
	})
}

func TestFileStorage_DeleteJob(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		_ = fs.SaveJob(ctx, &JobRecord{ID: "job1", Name: "J1"})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "job2", Name: "J2"})

		if err := fs.DeleteJob(ctx, "job1"); err != nil {
			t.Fatalf("DeleteJob error: %v", err)
		}

		_, err := fs.GetJob(ctx, "job1")
		if err != ErrJobNotFound {
			t.Fatalf("expected ErrJobNotFound after delete, got: %v", err)
		}

		jobs, _ := fs.ListJobs(ctx)
		if len(jobs) != 1 || jobs[0].ID != "job2" {
			t.Fatalf("unexpected jobs after delete: %+v", jobs)
		}
	})
}

func TestFileStorage_DeleteJob_NotFound(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()

		err := fs.DeleteJob(context.Background(), "nonexistent")
		if err != ErrJobNotFound {
			t.Fatalf("expected ErrJobNotFound, got: %v", err)
		}
	})
}

func TestFileStorage_ListJobs(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		_ = fs.SaveJob(ctx, &JobRecord{ID: "a", Name: "A"})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "b", Name: "B"})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "c", Name: "C"})

		jobs, err := fs.ListJobs(ctx)
		if err != nil {
			t.Fatalf("ListJobs error: %v", err)
		}
		if len(jobs) != 3 {
			t.Fatalf("expected 3 jobs, got %d", len(jobs))
		}
	})
}

func TestFileStorage_GetDueJobs(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		now := time.Now()

		_ = fs.SaveJob(ctx, &JobRecord{ID: "due", Name: "Due", NextRun: now.Add(-time.Minute)})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "future", Name: "Future", NextRun: now.Add(time.Hour)})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "paused", Name: "Paused", NextRun: now.Add(-time.Minute), Paused: true})
		_ = fs.SaveJob(ctx, &JobRecord{ID: "done", Name: "Done"})

		due, err := fs.GetDueJobs(ctx, now)
		if err != nil {
			t.Fatalf("GetDueJobs error: %v", err)
		}
		if len(due) != 1 || due[0].ID != "due" {
			t.Fatalf("expected 1 due job 'due', got %d: %+v", len(due), due)
		}
	})
}

func TestFileStorage_AcquireReleaseLock(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		fs, _ := NewFileStorage(tempFilePathExt(t, ext))
		defer fs.Close()
		ctx := context.Background()

		ok, err := fs.AcquireLock(ctx, "job1", "owner-a", 5*time.Minute)
		if err != nil || !ok {
			t.Fatalf("expected lock acquired, got ok=%v err=%v", ok, err)
		}

		ok, err = fs.AcquireLock(ctx, "job1", "owner-a", 5*time.Minute)
		if err != nil || !ok {
			t.Fatalf("expected lock re-acquired by same owner, got ok=%v err=%v", ok, err)
		}

		ok, err = fs.AcquireLock(ctx, "job1", "owner-b", 5*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected lock denied for different owner")
		}

		if err := fs.ReleaseLock(ctx, "job1", "owner-a"); err != nil {
			t.Fatalf("ReleaseLock error: %v", err)
		}

		ok, err = fs.AcquireLock(ctx, "job1", "owner-b", 5*time.Minute)
		if err != nil || !ok {
			t.Fatalf("expected lock acquired by owner-b after release, got ok=%v err=%v", ok, err)
		}
	})
}

func TestFileStorage_LockExpiry(t *testing.T) {
	fs, _ := NewFileStorage(tempFilePath(t))
	defer fs.Close()
	ctx := context.Background()

	ok, _ := fs.AcquireLock(ctx, "job1", "owner-a", 1*time.Millisecond)
	if !ok {
		t.Fatal("expected lock acquired")
	}

	time.Sleep(5 * time.Millisecond)

	ok, err := fs.AcquireLock(ctx, "job1", "owner-b", 5*time.Minute)
	if err != nil || !ok {
		t.Fatalf("expected lock acquired after expiry, got ok=%v err=%v", ok, err)
	}
}

func TestFileStorage_ReleaseLock_WrongOwner(t *testing.T) {
	fs, _ := NewFileStorage(tempFilePath(t))
	defer fs.Close()
	ctx := context.Background()

	_, _ = fs.AcquireLock(ctx, "job1", "owner-a", 5*time.Minute)
	_ = fs.ReleaseLock(ctx, "job1", "owner-b")

	ok, _ := fs.AcquireLock(ctx, "job1", "owner-b", 5*time.Minute)
	if ok {
		t.Fatal("lock should still be held by owner-a")
	}
}

func TestFileStorage_DeleteJob_RemovesLock(t *testing.T) {
	fs, _ := NewFileStorage(tempFilePath(t))
	defer fs.Close()
	ctx := context.Background()

	_ = fs.SaveJob(ctx, &JobRecord{ID: "job1", Name: "J1"})
	_, _ = fs.AcquireLock(ctx, "job1", "owner-a", 5*time.Minute)

	_ = fs.DeleteJob(ctx, "job1")

	ok, _ := fs.AcquireLock(ctx, "job1", "owner-b", 5*time.Minute)
	if !ok {
		t.Fatal("expected lock acquirable after job deletion")
	}
}

func TestFileStorage_Persistence(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		path := tempFilePathExt(t, ext)
		ctx := context.Background()

		fs1, _ := NewFileStorage(path)
		_ = fs1.SaveJob(ctx, &JobRecord{ID: "persist", Name: "Persist Test", RunCount: 42})
		_ = fs1.Close()

		fs2, err := NewFileStorage(path)
		if err != nil {
			t.Fatalf("NewFileStorage error on re-open: %v", err)
		}
		defer fs2.Close()

		got, err := fs2.GetJob(ctx, "persist")
		if err != nil {
			t.Fatalf("GetJob error after re-open: %v", err)
		}
		if got.Name != "Persist Test" || got.RunCount != 42 {
			t.Fatalf("persistence failed: %+v", got)
		}
	})
}

func TestFileStorage_WithScheduler(t *testing.T) {
	runForAllFormats(t, func(t *testing.T, ext string) {
		path := tempFilePathExt(t, ext)
		store, _ := NewFileStorage(path)
		defer store.Close()

		s := New(
			WithStorage(store),
			WithCheckInterval(50*time.Millisecond),
		)

		var counter int32
		_ = s.AddOneShotJob("fs-test", "File Storage Test", func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		}, 50*time.Millisecond)

		_ = s.Start()
		time.Sleep(300 * time.Millisecond)
		_ = s.Stop()

		if atomic.LoadInt32(&counter) < 1 {
			t.Fatal("expected at least 1 execution with FileStorage")
		}

		rec, err := store.GetJob(context.Background(), "fs-test")
		if err != nil {
			t.Fatalf("GetJob error: %v", err)
		}
		if rec.RunCount < 1 {
			t.Fatalf("expected RunCount >= 1, got %d", rec.RunCount)
		}
	})
}

func TestFileStorage_FileContents(t *testing.T) {
	// Verify the produced file contains expected format markers.
	tests := []struct {
		ext      string
		contains string
	}{
		{".yaml", "id: content-check"},
		{".json", "\"id\":\"content-check\""},
		{".xml", "<id>content-check</id>"},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			path := tempFilePathExt(t, tt.ext)
			fs, _ := NewFileStorage(path)
			ctx := context.Background()

			_ = fs.SaveJob(ctx, &JobRecord{ID: "content-check", Name: "Check"})
			_ = fs.Close()

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile error: %v", err)
			}
			content := string(data)
			if !strings.Contains(content, tt.contains) {
				t.Fatalf("expected file to contain %q, got:\n%s", tt.contains, content)
			}
		})
	}
}

func TestFileStorage_AllFieldsPersist(t *testing.T) {
	// Verify all JobRecord fields round-trip correctly for each format.
	runForAllFormats(t, func(t *testing.T, ext string) {
		path := tempFilePathExt(t, ext)
		ctx := context.Background()
		now := time.Now().Truncate(time.Second) // truncate for JSON/XML round-trip

		original := &JobRecord{
			ID:         "full",
			Name:       "Full Record",
			Status:     JobStatusFailed,
			Paused:     true,
			LastRun:    now.Add(-time.Hour),
			NextRun:    now.Add(time.Hour),
			RunCount:   17,
			ErrorCount: 3,
			LastError:  "something went wrong",
		}

		fs1, _ := NewFileStorage(path)
		_ = fs1.SaveJob(ctx, original)
		_ = fs1.Close()

		fs2, _ := NewFileStorage(path)
		defer fs2.Close()
		got, err := fs2.GetJob(ctx, "full")
		if err != nil {
			t.Fatalf("GetJob error: %v", err)
		}

		if got.ID != original.ID {
			t.Fatalf("ID: want %q, got %q", original.ID, got.ID)
		}
		if got.Name != original.Name {
			t.Fatalf("Name: want %q, got %q", original.Name, got.Name)
		}
		if got.Status != original.Status {
			t.Fatalf("Status: want %v, got %v", original.Status, got.Status)
		}
		if got.Paused != original.Paused {
			t.Fatalf("Paused: want %v, got %v", original.Paused, got.Paused)
		}
		if !got.LastRun.Equal(original.LastRun) {
			t.Fatalf("LastRun: want %v, got %v", original.LastRun, got.LastRun)
		}
		if !got.NextRun.Equal(original.NextRun) {
			t.Fatalf("NextRun: want %v, got %v", original.NextRun, got.NextRun)
		}
		if got.RunCount != original.RunCount {
			t.Fatalf("RunCount: want %d, got %d", original.RunCount, got.RunCount)
		}
		if got.ErrorCount != original.ErrorCount {
			t.Fatalf("ErrorCount: want %d, got %d", original.ErrorCount, got.ErrorCount)
		}
		if got.LastError != original.LastError {
			t.Fatalf("LastError: want %q, got %q", original.LastError, got.LastError)
		}
	})
}
