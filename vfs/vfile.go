package vfs

import (
	"io"
	"io/fs"
	"net/url"
)

type FileFilter func(file VFile) (bool, error)

// VFile is a handle to a single file (or directory) in a VFileSystem.
//
// Concurrency: a VFile is NOT safe for concurrent use. The
// embedded io.Reader / io.Writer / io.Seeker share a single cursor,
// and the property bag (see AddProperty) is not synchronized. Open one
// handle per goroutine or wrap with your own mutex.
//
// Lifecycle: callers MUST Close the handle. Behavior of subsequent
// method calls after Close is implementation-defined; well-behaved
// backends return vfs.ErrClosed.
//
// Property bag (AddProperty / GetProperty): semantics are backend-
// defined and intentionally narrow. The local-filesystem backend
// stores properties in process memory only and DOES NOT persist them
// across process restarts. Cloud backends MAY map properties to
// native object metadata (S3 x-amz-meta-*, GCS metadata, Azure
// metadata) — see each backend's package docs for constraints (key
// charsets, total size limits) and persistence guarantees. Do not
// rely on properties as a primary metadata store across backends.
type VFile interface {
	// Closer interface included from io package
	io.Closer
	// VFileContent provider interface included
	VFileContent
	// ListAll children of this file instance. can be nil in case of file object instead of directory
	ListAll() ([]VFile, error)
	// Delete the file object. If the file type is directory all files and subdirectories will be deleted
	Delete() error
	// DeleteAll deletes all the files and it's subdirectories
	DeleteAll() error
	// Info  Get the file ifo
	Info() (VFileInfo, error)
	// Parent of the file system
	Parent() (VFile, error)
	// Url of the file
	Url() *url.URL
	// AddProperty associates a name/value pair with the file. See the
	// VFile godoc for the per-backend persistence contract — local FS
	// keeps properties only in-process; cloud backends may map them to
	// native object metadata with their own constraints.
	AddProperty(name string, value string) error
	// GetProperty retrieves a previously-set property value. Returns
	// ("", nil) when the property is unset; see VFile godoc for
	// per-backend persistence semantics.
	GetProperty(name string) (string, error)
}

// VFileContent provides streaming access to a file's content via the
// io.Reader / io.Writer / io.Seeker primitives.
//
// Prefer the streaming methods for any work that scales with file
// size. AsString / AsBytes load the entire file into memory and are
// kept only as a backwards-compatible convenience; see their
// deprecation notes below for replacements.
type VFileContent interface {
	io.ReadWriteSeeker

	// AsString reads the file's entire content into a string.
	//
	// Deprecated: use vfs.ReadAllString(f) instead. This method
	// will be removed in a future major version. Loading large
	// cloud objects (S3, GCS, Azure) via this path can OOM the
	// process — prefer io.Copy / io.ReadAll directly when the
	// size is unknown.
	AsString() (string, error)

	// AsBytes reads the file's entire content into a byte slice.
	//
	// Deprecated: use vfs.ReadAllBytes(f) instead. Same OOM
	// caveats as AsString.
	AsBytes() ([]byte, error)

	// WriteString writes s to the underlying writer.
	WriteString(s string) (int, error)
	// ContentType of the underlying content. If not set defaults to UTF-8 for text files
	ContentType() string
}

type VFileInfo interface {
	fs.FileInfo
}
