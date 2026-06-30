package vfs

import (
	"context"
	"errors"
	"io"
	"net/url"
)

// ErrSkipDir is returned by a WalkFn callback to skip the rest of the
// current directory's contents without aborting the walk. Walk
// implementations that honor the sentinel SHOULD continue with the
// next sibling. Mirrors io/fs.SkipDir for consumer convenience.
var ErrSkipDir = errors.New("vfs: skip directory")

// ErrSkipAll is returned by a WalkFn callback to terminate the walk
// cleanly without surfacing as an error from Walk / WalkCtx. Walk
// implementations that honor the sentinel SHOULD treat it as success
// and stop iterating. Mirrors io/fs.SkipAll for consumer convenience.
var ErrSkipAll = errors.New("vfs: skip all")

// FileIterator yields VFiles one at a time without materializing the
// whole result set. Returned by Lister.ListIter — call Next until it
// returns io.EOF, then Close to release backend resources.
//
// FileIterator is NOT safe for concurrent use; one goroutine should
// own a given iterator.
type FileIterator interface {
	// Next returns the next VFile or io.EOF when no more remain.
	// Returns ctx.Err() if the context is canceled mid-iteration.
	Next(ctx context.Context) (VFile, error)
	// Close releases backend resources (HTTP connections, paging
	// state, etc.). Idempotent. Safe to call without consuming the
	// iterator to completion.
	Close() error
}

// Lister is an optional capability a VFileSystem may implement to
// support paginated, cancellable listing of large directories /
// prefixes. Backends without it fall back to the eager List() return
// in vfs.ListIter (the package helper).
//
// Cloud backends listing object stores (S3, GCS) SHOULD implement
// Lister so callers don't materialize million-key prefixes into one
// slice.
type Lister interface {
	ListIter(ctx context.Context, u *url.URL) (FileIterator, error)
}

// ListIter returns a paginated iterator over the entries at u. If fs
// implements Lister it is used; otherwise the helper falls back to the
// eager List() result wrapped as a slice-backed iterator.
//
// Always Close the returned iterator (even on early break) to release
// backend resources.
func ListIter(ctx context.Context, sys VFileSystem, u *url.URL) (FileIterator, error) {
	if l, ok := sys.(Lister); ok {
		return l.ListIter(ctx, u)
	}
	files, err := sys.List(u)
	if err != nil {
		return nil, err
	}
	return &sliceIterator{files: files}, nil
}

// sliceIterator wraps a fully-materialized []VFile as a FileIterator so
// the fallback path of ListIter has the same signature as a real
// paginated iterator.
type sliceIterator struct {
	files []VFile
	idx   int
	done  bool
}

// Next returns the next VFile in the slice or io.EOF when exhausted.
// Honors ctx cancellation between elements (cheap; the work is already
// done up-front).
func (s *sliceIterator) Next(ctx context.Context) (VFile, error) {
	if s.done {
		return nil, io.EOF
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.idx >= len(s.files) {
		s.done = true
		return nil, io.EOF
	}
	f := s.files[s.idx]
	s.idx++
	return f, nil
}

// Close marks the iterator finished. Idempotent.
func (s *sliceIterator) Close() error {
	s.done = true
	return nil
}
