package vfs

import (
	"context"
	"errors"
	"io"
	"net/url"
	"testing"
)

func TestReadRange_DelegatesToRangeReader(t *testing.T) {
	rr := &rangeReaderFake{want: []byte("partial")}
	got, err := ReadRange(context.Background(), rr, 10, 7)
	if err != nil {
		t.Fatalf("ReadRange: %v", err)
	}
	if string(got) != "partial" {
		t.Errorf("got %q, want %q", got, "partial")
	}
	if rr.lastOff != 10 || rr.lastLen != 7 {
		t.Errorf("off/len passthrough wrong: off=%d len=%d", rr.lastOff, rr.lastLen)
	}
}

func TestReadRange_FallbackSeekRead(t *testing.T) {
	// readableFakeFile implements Seek (returns 0,0,nil — accepts any offset)
	// and Read. ReadRange should drive both.
	f := &seekableFake{content: []byte("abcdefghij")}
	got, err := ReadRange(context.Background(), f, 3, 4)
	if err != nil {
		t.Fatalf("ReadRange fallback: %v", err)
	}
	if string(got) != "defg" {
		t.Errorf("got %q, want %q", got, "defg")
	}
}

func TestReadRange_FallbackZeroLengthReadsToEOF(t *testing.T) {
	f := &seekableFake{content: []byte("xyz")}
	got, err := ReadRange(context.Background(), f, 1, 0)
	if err != nil {
		t.Fatalf("ReadRange: %v", err)
	}
	if string(got) != "yz" {
		t.Errorf("got %q, want %q", got, "yz")
	}
}

func TestReadRange_NilFile(t *testing.T) {
	if _, err := ReadRange(context.Background(), nil, 0, 1); err == nil {
		t.Fatal("expected error for nil VFile")
	}
}

func TestReadRange_CtxCancelledBeforeFallback(t *testing.T) {
	f := &seekableFake{content: []byte("abc")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReadRange(ctx, f, 0, 1); !errors.Is(err, context.Canceled) {
		t.Errorf("expected ctx.Canceled, got %v", err)
	}
}

// --- helpers ---

// rangeReaderFake implements both VFile and RangeReader; only ReadRange is
// exercised by the range tests.
type rangeReaderFake struct {
	readableFakeFile
	want    []byte
	lastOff int64
	lastLen int64
}

func (r *rangeReaderFake) ReadRange(_ context.Context, off, length int64) ([]byte, error) {
	r.lastOff, r.lastLen = off, length
	return r.want, nil
}

// seekableFake supports a proper Seek + Read pair for the fallback path.
type seekableFake struct {
	content []byte
	pos     int64
}

func (s *seekableFake) Read(p []byte) (int, error) {
	if s.pos >= int64(len(s.content)) {
		return 0, io.EOF
	}
	n := copy(p, s.content[s.pos:])
	s.pos += int64(n)
	return n, nil
}
func (s *seekableFake) Write([]byte) (int, error) { return 0, nil }
func (s *seekableFake) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		s.pos = off
	case io.SeekCurrent:
		s.pos += off
	case io.SeekEnd:
		s.pos = int64(len(s.content)) + off
	}
	return s.pos, nil
}
func (s *seekableFake) Close() error                       { return nil }
func (s *seekableFake) AsString() (string, error)          { return string(s.content), nil }
func (s *seekableFake) AsBytes() ([]byte, error)           { return s.content, nil }
func (s *seekableFake) WriteString(string) (int, error)    { return 0, nil }
func (s *seekableFake) ContentType() string                { return "" }
func (s *seekableFake) ListAll() ([]VFile, error)          { return nil, nil }
func (s *seekableFake) Delete() error                      { return nil }
func (s *seekableFake) DeleteAll() error                   { return nil }
func (s *seekableFake) Info() (VFileInfo, error)           { return nil, nil }
func (s *seekableFake) Parent() (VFile, error)             { return nil, nil }
func (s *seekableFake) Url() *url.URL                      { return &url.URL{} }
func (s *seekableFake) AddProperty(string, string) error   { return nil }
func (s *seekableFake) GetProperty(string) (string, error) { return "", nil }
