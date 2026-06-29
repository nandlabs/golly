// Package vfs provides a URL-addressed, scheme-pluggable virtual filesystem
// abstraction with first-class support for cloud object stores.
//
// # Overview
//
// The package centers on three small interfaces:
//
//   - VFileSystem — operations: Open / Create / Delete / List / Walk / ...
//   - VFile       — a file handle: io.ReadWriteSeeker + Info + Delete + ...
//   - VFileInfo   — metadata: extends io/fs.FileInfo
//
// Backends register themselves with the package-level Manager by URL
// scheme; callers use URLs (or strings via the *Raw variants) and the
// Manager dispatches. Local FS ships in-tree (oss.nandlabs.io/golly/vfs);
// S3 lives in oss.nandlabs.io/golly-aws/s3; GCS in
// oss.nandlabs.io/golly-gcp/gs.
//
// # Optional capability interfaces
//
// Several capabilities are exposed as optional interfaces that backends
// opt in to. Callers use the package-level helpers, which delegate when
// the backend implements the capability and fall back to best-effort
// emulation otherwise:
//
//   - VFileSystemCtx — context.Context-aware variants of the heavy I/O
//     operations. Helpers: vfs.OpenCtx, CreateCtx, DeleteCtx, ListCtx,
//     MkdirAllCtx, CopyCtx, MoveCtx, WalkCtx.
//   - Lister         — paginated, cancellable listing. Helper:
//     vfs.ListIter.
//   - RangeReader    — HTTP-Range-style partial reads on a VFile.
//     Helper: vfs.ReadRange.
//
// Sentinel errors (vfs.ErrNotExist, ErrPermission, ErrNotSupported,
// ErrIsDir, ErrNotDir, ErrClosed) wrap io/fs equivalents where possible
// so errors.Is(err, fs.ErrNotExist) keeps working across backends.
//
// # Convenience helpers
//
//   - vfs.ReadAllString(f), vfs.ReadAllBytes(f) — replace the deprecated
//     VFile.AsString / AsBytes methods.
//   - vfs.Exists(fs, u) / vfs.ExistsCtx(ctx, fs, u) — existence check
//     that handles the open-then-Info pattern most backends need.
//
// # Concurrency
//
// A VFile is NOT safe for concurrent use; the read/write cursor is
// single-state. Open separate handles per goroutine. The VFileSystem
// itself IS safe for concurrent use across all in-tree implementations
// (caveats noted on individual backends).
//
// # Roadmap (v2)
//
// The current interface duplicates every method as Foo(*url.URL) and
// FooRaw(string). A future major version (v2) will collapse this — the
// URL form is the canonical one; the Raw form is convenience that can
// move to free helpers. Prefer the URL-typed methods in new code.
package vfs
