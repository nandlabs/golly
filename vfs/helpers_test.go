package vfs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/url"
	"testing"
	"time"
)

func TestReadAllBytes_StreamsViaReader(t *testing.T) {
	f := &readableFakeFile{content: []byte("hello world")}
	got, err := ReadAllBytes(f)
	if err != nil {
		t.Fatalf("ReadAllBytes: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestReadAllString_StreamsViaReader(t *testing.T) {
	f := &readableFakeFile{content: []byte("text content")}
	got, err := ReadAllString(f)
	if err != nil {
		t.Fatalf("ReadAllString: %v", err)
	}
	if got != "text content" {
		t.Errorf("got %q, want %q", got, "text content")
	}
}

func TestReadAllBytes_NilFile(t *testing.T) {
	if _, err := ReadAllBytes(nil); err == nil {
		t.Fatal("expected error for nil VFile")
	}
}

func TestSentinelErrors_WrapStdlib(t *testing.T) {
	if !errors.Is(ErrNotExist, fs.ErrNotExist) {
		t.Error("ErrNotExist should wrap fs.ErrNotExist")
	}
	if !errors.Is(ErrPermission, fs.ErrPermission) {
		t.Error("ErrPermission should wrap fs.ErrPermission")
	}
}

func TestExistsCtx_DispatchesToInfo(t *testing.T) {
	fs := &existsFakeFS{info: &fakeInfo{name: "x"}}
	ok, err := ExistsCtx(context.Background(), fs, mustURLForExists(t, "file:///x"))
	if err != nil {
		t.Fatalf("ExistsCtx: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true, fakeFS returned a valid Info")
	}
}

func TestExistsCtx_NotExistReturnsFalseNoError(t *testing.T) {
	missing := errors.Join(ErrNotExist) // wraps fs.ErrNotExist via sentinel
	fs := &existsFakeFS{infoErr: missing}
	ok, err := ExistsCtx(context.Background(), fs, mustURLForExists(t, "file:///missing"))
	if err != nil {
		t.Fatalf("ExistsCtx should swallow not-exist; got err=%v", err)
	}
	if ok {
		t.Fatalf("expected ok=false for missing file")
	}
}

func TestExistsCtx_OpenError(t *testing.T) {
	fs := &existsFakeFS{openErr: errors.New("network down")}
	ok, err := ExistsCtx(context.Background(), fs, mustURLForExists(t, "file:///x"))
	if err == nil {
		t.Fatalf("expected error to bubble up")
	}
	if ok {
		t.Fatalf("ok must be false on error")
	}
}

func TestExistsCtx_NilFS(t *testing.T) {
	if _, err := ExistsCtx(context.Background(), nil, &url.URL{}); err == nil {
		t.Fatal("expected error for nil fs")
	}
}

// --- helpers ---

func mustURLForExists(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u
}

// readableFakeFile satisfies VFile minimally for ReadAllBytes /
// ReadAllString tests — only Read is exercised, the rest are nil-ish.
type readableFakeFile struct {
	content []byte
	pos     int
}

func (f *readableFakeFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.content) {
		return 0, io.EOF
	}
	n := copy(p, f.content[f.pos:])
	f.pos += n
	return n, nil
}
func (f *readableFakeFile) Write([]byte) (int, error)          { return 0, nil }
func (f *readableFakeFile) Seek(int64, int) (int64, error)     { return 0, nil }
func (f *readableFakeFile) Close() error                       { return nil }
func (f *readableFakeFile) AsString() (string, error)          { return string(f.content), nil }
func (f *readableFakeFile) AsBytes() ([]byte, error)           { return f.content, nil }
func (f *readableFakeFile) WriteString(string) (int, error)    { return 0, nil }
func (f *readableFakeFile) ContentType() string                { return "" }
func (f *readableFakeFile) ListAll() ([]VFile, error)          { return nil, nil }
func (f *readableFakeFile) Delete() error                      { return nil }
func (f *readableFakeFile) DeleteAll() error                   { return nil }
func (f *readableFakeFile) Info() (VFileInfo, error)           { return nil, nil }
func (f *readableFakeFile) Parent() (VFile, error)             { return nil, nil }
func (f *readableFakeFile) Url() *url.URL                      { return &url.URL{} }
func (f *readableFakeFile) AddProperty(string, string) error   { return nil }
func (f *readableFakeFile) GetProperty(string) (string, error) { return "", nil }

// existsFakeFS exercises ExistsCtx without needing a real backend.
type existsFakeFS struct {
	openErr error
	info    VFileInfo
	infoErr error
}

func (e *existsFakeFS) Open(*url.URL) (VFile, error) {
	if e.openErr != nil {
		return nil, e.openErr
	}
	return &infoFakeFile{info: e.info, err: e.infoErr}, nil
}
func (e *existsFakeFS) Copy(src, dst *url.URL) error               { return nil }
func (e *existsFakeFS) CopyRaw(src, dst string) error              { return nil }
func (e *existsFakeFS) Create(*url.URL) (VFile, error)             { return nil, nil }
func (e *existsFakeFS) CreateRaw(string) (VFile, error)            { return nil, nil }
func (e *existsFakeFS) Delete(*url.URL) error                      { return nil }
func (e *existsFakeFS) DeleteRaw(string) error                     { return nil }
func (e *existsFakeFS) List(*url.URL) ([]VFile, error)             { return nil, nil }
func (e *existsFakeFS) ListRaw(string) ([]VFile, error)            { return nil, nil }
func (e *existsFakeFS) Mkdir(*url.URL) (VFile, error)              { return nil, nil }
func (e *existsFakeFS) MkdirRaw(string) (VFile, error)             { return nil, nil }
func (e *existsFakeFS) MkdirAll(*url.URL) (VFile, error)           { return nil, nil }
func (e *existsFakeFS) MkdirAllRaw(string) (VFile, error)          { return nil, nil }
func (e *existsFakeFS) Move(src, dst *url.URL) error               { return nil }
func (e *existsFakeFS) MoveRaw(src, dst string) error              { return nil }
func (e *existsFakeFS) OpenRaw(string) (VFile, error)              { return nil, nil }
func (e *existsFakeFS) Schemes() []string                          { return []string{"file"} }
func (e *existsFakeFS) Walk(*url.URL, WalkFn) error                { return nil }
func (e *existsFakeFS) Find(*url.URL, FileFilter) ([]VFile, error) { return nil, nil }
func (e *existsFakeFS) WalkRaw(string, WalkFn) error               { return nil }
func (e *existsFakeFS) DeleteMatching(*url.URL, FileFilter) error  { return nil }

// infoFakeFile only meaningfully implements Info; reads/writes/etc. are nil.
type infoFakeFile struct {
	info VFileInfo
	err  error
}

func (i *infoFakeFile) Read([]byte) (int, error)           { return 0, io.EOF }
func (i *infoFakeFile) Write([]byte) (int, error)          { return 0, nil }
func (i *infoFakeFile) Seek(int64, int) (int64, error)     { return 0, nil }
func (i *infoFakeFile) Close() error                       { return nil }
func (i *infoFakeFile) AsString() (string, error)          { return "", nil }
func (i *infoFakeFile) AsBytes() ([]byte, error)           { return nil, nil }
func (i *infoFakeFile) WriteString(string) (int, error)    { return 0, nil }
func (i *infoFakeFile) ContentType() string                { return "" }
func (i *infoFakeFile) ListAll() ([]VFile, error)          { return nil, nil }
func (i *infoFakeFile) Delete() error                      { return nil }
func (i *infoFakeFile) DeleteAll() error                   { return nil }
func (i *infoFakeFile) Info() (VFileInfo, error)           { return i.info, i.err }
func (i *infoFakeFile) Parent() (VFile, error)             { return nil, nil }
func (i *infoFakeFile) Url() *url.URL                      { return &url.URL{} }
func (i *infoFakeFile) AddProperty(string, string) error   { return nil }
func (i *infoFakeFile) GetProperty(string) (string, error) { return "", nil }

type fakeInfo struct{ name string }

func (f *fakeInfo) Name() string       { return f.name }
func (f *fakeInfo) Size() int64        { return 0 }
func (f *fakeInfo) Mode() fs.FileMode  { return 0 }
func (f *fakeInfo) ModTime() time.Time { return time.Time{} }
func (f *fakeInfo) IsDir() bool        { return false }
func (f *fakeInfo) Sys() any           { return nil }
