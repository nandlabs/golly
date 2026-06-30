package vfs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/url"
)

// ReadAllString reads the entire content of f into a string.
//
// Use deliberately: for large files this loads everything into memory
// and is the streaming-safe equivalent of VFile.AsString() which is
// deprecated. Prefer io.Copy / io.CopyBuffer when the caller can write
// directly to a sink.
//
// The file's read position is advanced to EOF. The caller is responsible
// for Close.
func ReadAllString(f VFile) (string, error) {
	b, err := ReadAllBytes(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReadAllBytes reads the entire content of f into a byte slice.
//
// Use deliberately: for large files this loads everything into memory
// and is the streaming-safe equivalent of VFile.AsBytes() which is
// deprecated. Prefer io.Copy / io.CopyBuffer for streaming flows.
//
// The file's read position is advanced to EOF. The caller is responsible
// for Close.
func ReadAllBytes(f VFile) ([]byte, error) {
	if f == nil {
		return nil, errors.New("vfs: nil VFile")
	}
	return io.ReadAll(f)
}

// Exists reports whether the given URL resolves to a real file or
// directory on the supplied filesystem.
//
// (true, nil)  — the entity exists
// (false, nil) — the entity does not exist (errors.Is(err, fs.ErrNotExist))
// (false, err) — some other error was encountered (permission, network…);
//
//	the returned err is wrapped, callers may errors.Is it
//	against the vfs sentinels.
func Exists(fs VFileSystem, u *url.URL) (bool, error) {
	return ExistsCtx(context.Background(), fs, u)
}

// ExistsCtx is the context-aware variant of Exists. If the underlying
// filesystem implements VFileSystemCtx, OpenCtx is used; otherwise the
// helper falls back to the non-ctx Open via runCtx.
func ExistsCtx(ctx context.Context, sys VFileSystem, u *url.URL) (bool, error) {
	if sys == nil {
		return false, errors.New("vfs: nil VFileSystem")
	}
	f, err := OpenCtx(ctx, sys, u)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	// Many backends (S3, GCS) construct a handle eagerly without
	// checking existence — Open succeeds even for missing keys. A
	// follow-up Info() is what actually probes existence on most
	// implementations.
	_, infoErr := f.Info()
	_ = f.Close()
	if infoErr != nil {
		if errors.Is(infoErr, fs.ErrNotExist) {
			return false, nil
		}
		return false, infoErr
	}
	return true, nil
}
