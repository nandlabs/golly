package vfs

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"
)

// --- Fallback path (FS does NOT implement VFileSystemCtx) ---

func TestOpenCtx_FallbackSucceeds(t *testing.T) {
	fs := &fakeFS{}
	_, err := OpenCtx(context.Background(), fs, mustURL(t, "file:///tmp/x"))
	if err != nil {
		t.Fatalf("OpenCtx: %v", err)
	}
	if !fs.openCalled {
		t.Fatalf("fallback should have invoked fs.Open")
	}
}

func TestOpenCtx_FallbackCancellation(t *testing.T) {
	fs := &fakeFS{openDelay: 200 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := OpenCtx(ctx, fs, mustURL(t, "file:///slow"))
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected ctx deadline error; got %v", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("OpenCtx didn't return promptly on cancel; took %v", elapsed)
	}
}

func TestDeleteCtx_FallbackPropagatesError(t *testing.T) {
	want := errors.New("nope")
	fs := &fakeFS{deleteErr: want}
	got := DeleteCtx(context.Background(), fs, mustURL(t, "file:///x"))
	if !errors.Is(got, want) {
		t.Fatalf("DeleteCtx err = %v, want %v", got, want)
	}
}

// --- Delegated path (FS implements VFileSystemCtx) ---

func TestOpenCtx_DelegatesToCtxImpl(t *testing.T) {
	fs := &ctxFS{}
	_, err := OpenCtx(context.Background(), fs, mustURL(t, "file:///x"))
	if err != nil {
		t.Fatalf("OpenCtx: %v", err)
	}
	if !fs.openCtxCalled {
		t.Fatalf("expected OpenCtx to delegate to ctxFS.OpenCtx")
	}
}

func TestAllHelpers_DelegateWhenAvailable(t *testing.T) {
	fs := &ctxFS{}
	ctx := context.Background()
	u := mustURL(t, "file:///x")
	_, _ = OpenCtx(ctx, fs, u)
	_, _ = CreateCtx(ctx, fs, u)
	_ = DeleteCtx(ctx, fs, u)
	_, _ = ListCtx(ctx, fs, u)
	_, _ = MkdirAllCtx(ctx, fs, u)
	_ = CopyCtx(ctx, fs, u, u)
	_ = MoveCtx(ctx, fs, u, u)
	_ = WalkCtx(ctx, fs, u, func(VFile) error { return nil })

	checks := []struct {
		name string
		hit  bool
	}{
		{"OpenCtx", fs.openCtxCalled},
		{"CreateCtx", fs.createCtxCalled},
		{"DeleteCtx", fs.deleteCtxCalled},
		{"ListCtx", fs.listCtxCalled},
		{"MkdirAllCtx", fs.mkdirAllCtxCalled},
		{"CopyCtx", fs.copyCtxCalled},
		{"MoveCtx", fs.moveCtxCalled},
		{"WalkCtx", fs.walkCtxCalled},
	}
	for _, c := range checks {
		if !c.hit {
			t.Errorf("%s did not delegate to the VFileSystemCtx impl", c.name)
		}
	}
}

// --- Helpers ---

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u
}

// fakeFS implements VFileSystem (NOT VFileSystemCtx). Returned VFiles are
// nil — these tests only assert call dispatch / cancellation timing, not
// VFile behavior.
type fakeFS struct {
	openCalled bool
	openDelay  time.Duration
	deleteErr  error
}

func (f *fakeFS) Copy(src, dst *url.URL) error      { return nil }
func (f *fakeFS) CopyRaw(src, dst string) error     { return nil }
func (f *fakeFS) Create(u *url.URL) (VFile, error)  { return nil, nil }
func (f *fakeFS) CreateRaw(string) (VFile, error)   { return nil, nil }
func (f *fakeFS) Delete(*url.URL) error             { return f.deleteErr }
func (f *fakeFS) DeleteRaw(string) error            { return f.deleteErr }
func (f *fakeFS) List(*url.URL) ([]VFile, error)    { return nil, nil }
func (f *fakeFS) ListRaw(string) ([]VFile, error)   { return nil, nil }
func (f *fakeFS) Mkdir(*url.URL) (VFile, error)     { return nil, nil }
func (f *fakeFS) MkdirRaw(string) (VFile, error)    { return nil, nil }
func (f *fakeFS) MkdirAll(*url.URL) (VFile, error)  { return nil, nil }
func (f *fakeFS) MkdirAllRaw(string) (VFile, error) { return nil, nil }
func (f *fakeFS) Move(src, dst *url.URL) error      { return nil }
func (f *fakeFS) MoveRaw(src, dst string) error     { return nil }
func (f *fakeFS) Open(u *url.URL) (VFile, error) {
	f.openCalled = true
	if f.openDelay > 0 {
		time.Sleep(f.openDelay)
	}
	return nil, nil
}
func (f *fakeFS) OpenRaw(string) (VFile, error)              { return nil, nil }
func (f *fakeFS) Schemes() []string                          { return []string{"file"} }
func (f *fakeFS) Walk(*url.URL, WalkFn) error                { return nil }
func (f *fakeFS) Find(*url.URL, FileFilter) ([]VFile, error) { return nil, nil }
func (f *fakeFS) WalkRaw(string, WalkFn) error               { return nil }
func (f *fakeFS) DeleteMatching(*url.URL, FileFilter) error  { return nil }

// ctxFS implements both VFileSystem and VFileSystemCtx; records which Ctx
// methods were called so tests can assert delegation.
type ctxFS struct {
	fakeFS
	openCtxCalled, createCtxCalled, deleteCtxCalled, listCtxCalled bool
	mkdirAllCtxCalled, copyCtxCalled, moveCtxCalled, walkCtxCalled bool
}

func (c *ctxFS) OpenCtx(context.Context, *url.URL) (VFile, error) {
	c.openCtxCalled = true
	return nil, nil
}
func (c *ctxFS) CreateCtx(context.Context, *url.URL) (VFile, error) {
	c.createCtxCalled = true
	return nil, nil
}
func (c *ctxFS) DeleteCtx(context.Context, *url.URL) error {
	c.deleteCtxCalled = true
	return nil
}
func (c *ctxFS) ListCtx(context.Context, *url.URL) ([]VFile, error) {
	c.listCtxCalled = true
	return nil, nil
}
func (c *ctxFS) MkdirAllCtx(context.Context, *url.URL) (VFile, error) {
	c.mkdirAllCtxCalled = true
	return nil, nil
}
func (c *ctxFS) CopyCtx(context.Context, *url.URL, *url.URL) error {
	c.copyCtxCalled = true
	return nil
}
func (c *ctxFS) MoveCtx(context.Context, *url.URL, *url.URL) error {
	c.moveCtxCalled = true
	return nil
}
func (c *ctxFS) WalkCtx(context.Context, *url.URL, WalkFn) error {
	c.walkCtxCalled = true
	return nil
}
