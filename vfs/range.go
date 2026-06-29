package vfs

import (
	"context"
	"errors"
	"io"
)

// RangeReader is an optional capability a VFile may implement to expose
// HTTP-Range-style partial reads. Cloud backends (S3, GCS) SHOULD
// implement it and map ReadRange to a native ranged GET so callers can
// fetch a small window of a large object cheaply (file headers, media
// byte ranges, resumable downloads).
//
// Local-filesystem implementations gain little — Seek+Read is the same
// operation under the hood — but may implement it for consistency.
type RangeReader interface {
	// ReadRange returns up to length bytes starting at offset off.
	// A length of 0 means "read to EOF from off". Returns (nil, io.EOF)
	// when off is past the end. Implementations should not perform a
	// full-object download to satisfy a small range.
	ReadRange(ctx context.Context, off, length int64) ([]byte, error)
}

// ReadRange returns a slice of f's content [off, off+length). If f
// implements RangeReader the call is delegated; otherwise the helper
// falls back to Seek+Read, which on cloud backends may perform a
// full-object download.
//
// length == 0 means "read to EOF from off".
//
// Returns ErrNotSupported if f does not implement RangeReader AND the
// fallback (Seek + Read) is unavailable because f does not support
// io.Seeker. Returns io.EOF when off is past the end.
func ReadRange(ctx context.Context, f VFile, off, length int64) ([]byte, error) {
	if f == nil {
		return nil, errors.New("vfs: nil VFile")
	}
	if rr, ok := f.(RangeReader); ok {
		return rr.ReadRange(ctx, off, length)
	}
	// Fallback: seek + read. This may pull the whole object on cloud
	// backends that don't implement Seek natively, so this path is
	// best-effort and documented as such.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return nil, err
	}
	if length <= 0 {
		return io.ReadAll(f)
	}
	buf := make([]byte, length)
	n, err := io.ReadFull(f, buf)
	if err == io.ErrUnexpectedEOF {
		return buf[:n], nil
	}
	if err != nil {
		return nil, err
	}
	return buf, nil
}
