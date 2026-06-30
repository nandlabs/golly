package vfs

import (
	"errors"
	"fmt"
	"io/fs"
)

// Sentinel errors that VFileSystem implementations should return (or wrap
// via fmt.Errorf with %w) so callers can switch on outcomes with
// errors.Is across backends.
//
// Where the meaning of the underlying error overlaps with io/fs, the
// sentinel is wrapped around io/fs.ErrNotExist / fs.ErrPermission so
// errors.Is(err, fs.ErrNotExist) also returns true. Code that checks
// stdlib sentinels keeps working.
var (
	// ErrNotExist is returned when a file or directory does not exist
	// at the given URL.
	ErrNotExist = fmt.Errorf("vfs: file does not exist: %w", fs.ErrNotExist)

	// ErrPermission is returned when the caller lacks the permission
	// required for the operation.
	ErrPermission = fmt.Errorf("vfs: permission denied: %w", fs.ErrPermission)

	// ErrNotSupported is returned by backends that do not implement an
	// optional capability (e.g. ranged reads against a backend that
	// only supports streaming) or by package-level helpers when a
	// backend does not satisfy the required optional interface.
	ErrNotSupported = errors.New("vfs: operation not supported")

	// ErrIsDir is returned when an operation that requires a regular
	// file is invoked on a directory.
	ErrIsDir = errors.New("vfs: is a directory")

	// ErrNotDir is returned when an operation that requires a directory
	// is invoked on a regular file.
	ErrNotDir = errors.New("vfs: not a directory")

	// ErrClosed is returned by VFile methods after Close() has been
	// called.
	ErrClosed = errors.New("vfs: closed")
)
