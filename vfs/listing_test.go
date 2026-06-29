package vfs

import (
	"context"
	"errors"
	"io"
	"net/url"
	"testing"
)

func TestListIter_FallbackOverEagerList(t *testing.T) {
	want := []VFile{&readableFakeFile{content: []byte("a")}, &readableFakeFile{content: []byte("b")}}
	fs := &listFakeFS{eager: want}
	it, err := ListIter(context.Background(), fs, mustURLForList(t, "file:///dir"))
	if err != nil {
		t.Fatalf("ListIter: %v", err)
	}
	defer it.Close()

	count := 0
	for {
		_, err := it.Next(context.Background())
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		count++
	}
	if count != len(want) {
		t.Fatalf("got %d items, want %d", count, len(want))
	}
}

func TestListIter_DelegatesToListerWhenAvailable(t *testing.T) {
	called := false
	fs := &listerFakeFS{
		hook: func(context.Context, *url.URL) (FileIterator, error) {
			called = true
			return &sliceIterator{}, nil
		},
	}
	_, err := ListIter(context.Background(), fs, mustURLForList(t, "file:///dir"))
	if err != nil {
		t.Fatalf("ListIter: %v", err)
	}
	if !called {
		t.Fatalf("expected Lister.ListIter to be called")
	}
}

func TestSliceIterator_RespectsCtxCancellation(t *testing.T) {
	it := &sliceIterator{files: []VFile{&readableFakeFile{}, &readableFakeFile{}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := it.Next(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected ctx.Canceled, got %v", err)
	}
}

func TestSliceIterator_CloseIsIdempotent(t *testing.T) {
	it := &sliceIterator{files: []VFile{&readableFakeFile{}}}
	if err := it.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := it.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := it.Next(context.Background()); err != io.EOF {
		t.Fatalf("Next after Close should be EOF; got %v", err)
	}
}

func TestWalkSkipSentinels_Defined(t *testing.T) {
	if ErrSkipDir == nil || ErrSkipAll == nil {
		t.Fatal("ErrSkipDir and ErrSkipAll must be non-nil")
	}
	if ErrSkipDir == ErrSkipAll {
		t.Fatal("ErrSkipDir and ErrSkipAll must be distinct sentinels")
	}
}

// --- helpers ---

func mustURLForList(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u
}

// listFakeFS satisfies VFileSystem but NOT Lister — exercises the eager
// List fallback path.
type listFakeFS struct {
	eager []VFile
}

func (f *listFakeFS) Copy(src, dst *url.URL) error               { return nil }
func (f *listFakeFS) CopyRaw(src, dst string) error              { return nil }
func (f *listFakeFS) Create(*url.URL) (VFile, error)             { return nil, nil }
func (f *listFakeFS) CreateRaw(string) (VFile, error)            { return nil, nil }
func (f *listFakeFS) Delete(*url.URL) error                      { return nil }
func (f *listFakeFS) DeleteRaw(string) error                     { return nil }
func (f *listFakeFS) List(*url.URL) ([]VFile, error)             { return f.eager, nil }
func (f *listFakeFS) ListRaw(string) ([]VFile, error)            { return f.eager, nil }
func (f *listFakeFS) Mkdir(*url.URL) (VFile, error)              { return nil, nil }
func (f *listFakeFS) MkdirRaw(string) (VFile, error)             { return nil, nil }
func (f *listFakeFS) MkdirAll(*url.URL) (VFile, error)           { return nil, nil }
func (f *listFakeFS) MkdirAllRaw(string) (VFile, error)          { return nil, nil }
func (f *listFakeFS) Move(src, dst *url.URL) error               { return nil }
func (f *listFakeFS) MoveRaw(src, dst string) error              { return nil }
func (f *listFakeFS) Open(*url.URL) (VFile, error)               { return nil, nil }
func (f *listFakeFS) OpenRaw(string) (VFile, error)              { return nil, nil }
func (f *listFakeFS) Schemes() []string                          { return []string{"file"} }
func (f *listFakeFS) Walk(*url.URL, WalkFn) error                { return nil }
func (f *listFakeFS) Find(*url.URL, FileFilter) ([]VFile, error) { return nil, nil }
func (f *listFakeFS) WalkRaw(string, WalkFn) error               { return nil }
func (f *listFakeFS) DeleteMatching(*url.URL, FileFilter) error  { return nil }

// listerFakeFS satisfies both VFileSystem and Lister.
type listerFakeFS struct {
	listFakeFS
	hook func(context.Context, *url.URL) (FileIterator, error)
}

func (f *listerFakeFS) ListIter(ctx context.Context, u *url.URL) (FileIterator, error) {
	return f.hook(ctx, u)
}
