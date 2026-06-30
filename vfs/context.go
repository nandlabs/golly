package vfs

import (
	"context"
	"net/url"
)

// VFileSystemCtx is an optional capability a VFileSystem may implement to
// support context-aware variants of the I/O-heavy operations. Providers
// signal support by satisfying this interface in addition to VFileSystem.
//
// Callers should prefer the package-level helpers (OpenCtx, CreateCtx,
// CopyCtx, …) rather than type-asserting directly — the helpers fall back
// to the non-ctx methods with best-effort cancellation when a backend
// does not implement VFileSystemCtx.
//
// Network-backed implementations (S3, GCS, Azure Blob, etc.) should
// implement VFileSystemCtx so the context propagates into the underlying
// SDK call for genuine cancellation and deadline support. Local-filesystem
// implementations gain little from implementing it directly, since the
// underlying os package does not accept a context — the fallback path
// already provides wait-level cancellation.
type VFileSystemCtx interface {
	OpenCtx(ctx context.Context, u *url.URL) (VFile, error)
	CreateCtx(ctx context.Context, u *url.URL) (VFile, error)
	DeleteCtx(ctx context.Context, u *url.URL) error
	ListCtx(ctx context.Context, u *url.URL) ([]VFile, error)
	MkdirAllCtx(ctx context.Context, u *url.URL) (VFile, error)
	CopyCtx(ctx context.Context, src, dst *url.URL) error
	MoveCtx(ctx context.Context, src, dst *url.URL) error
	WalkCtx(ctx context.Context, u *url.URL, fn WalkFn) error
}

// runCtx runs op in a goroutine and races it against ctx cancellation.
// On ctx cancellation, returns ctx.Err() — op may still complete in the
// background (the underlying syscall is not cancellable). This is
// best-effort cancellation: callers regain control immediately, but
// running work cannot always be terminated. Network backends that fully
// support context should implement VFileSystemCtx directly instead.
func runCtx[T any](ctx context.Context, op func() (T, error)) (T, error) {
	type result struct {
		v   T
		err error
	}
	done := make(chan result, 1)
	go func() {
		v, err := op()
		done <- result{v, err}
	}()
	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// runCtxErr is the void-returning variant of runCtx for ops that only
// return error.
func runCtxErr(ctx context.Context, op func() error) error {
	_, err := runCtx(ctx, func() (struct{}, error) {
		return struct{}{}, op()
	})
	return err
}

// OpenCtx opens u with the given context. If fs implements VFileSystemCtx
// the call is delegated; otherwise it runs Open in a goroutine and races
// it against ctx.
func OpenCtx(ctx context.Context, fs VFileSystem, u *url.URL) (VFile, error) {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.OpenCtx(ctx, u)
	}
	return runCtx(ctx, func() (VFile, error) { return fs.Open(u) })
}

// CreateCtx mirrors VFileSystem.Create with context support.
func CreateCtx(ctx context.Context, fs VFileSystem, u *url.URL) (VFile, error) {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.CreateCtx(ctx, u)
	}
	return runCtx(ctx, func() (VFile, error) { return fs.Create(u) })
}

// DeleteCtx mirrors VFileSystem.Delete with context support.
func DeleteCtx(ctx context.Context, fs VFileSystem, u *url.URL) error {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.DeleteCtx(ctx, u)
	}
	return runCtxErr(ctx, func() error { return fs.Delete(u) })
}

// ListCtx mirrors VFileSystem.List with context support.
func ListCtx(ctx context.Context, fs VFileSystem, u *url.URL) ([]VFile, error) {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.ListCtx(ctx, u)
	}
	return runCtx(ctx, func() ([]VFile, error) { return fs.List(u) })
}

// MkdirAllCtx mirrors VFileSystem.MkdirAll with context support.
func MkdirAllCtx(ctx context.Context, fs VFileSystem, u *url.URL) (VFile, error) {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.MkdirAllCtx(ctx, u)
	}
	return runCtx(ctx, func() (VFile, error) { return fs.MkdirAll(u) })
}

// CopyCtx mirrors VFileSystem.Copy with context support.
func CopyCtx(ctx context.Context, fs VFileSystem, src, dst *url.URL) error {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.CopyCtx(ctx, src, dst)
	}
	return runCtxErr(ctx, func() error { return fs.Copy(src, dst) })
}

// MoveCtx mirrors VFileSystem.Move with context support.
func MoveCtx(ctx context.Context, fs VFileSystem, src, dst *url.URL) error {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.MoveCtx(ctx, src, dst)
	}
	return runCtxErr(ctx, func() error { return fs.Move(src, dst) })
}

// WalkCtx mirrors VFileSystem.Walk with context support. The fn callback
// is invoked synchronously from the same goroutine running Walk, so it
// must not block indefinitely; callers wanting per-file ctx checking
// should test ctx.Err() inside fn.
func WalkCtx(ctx context.Context, fs VFileSystem, u *url.URL, fn WalkFn) error {
	if v, ok := fs.(VFileSystemCtx); ok {
		return v.WalkCtx(ctx, u, fn)
	}
	return runCtxErr(ctx, func() error { return fs.Walk(u, fn) })
}
