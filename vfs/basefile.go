package vfs

import (
	"io"
)

type BaseFile struct {
	VFile
}

func (b *BaseFile) AsString() (s string, err error) {
	var bytes []byte
	bytes, err = io.ReadAll(b)
	if err == nil {
		s = string(bytes)
	}
	return
}

func (b *BaseFile) AsBytes() ([]byte, error) {
	return io.ReadAll(b)
}

func (b *BaseFile) WriteString(s string) (int, error) {
	return io.WriteString(b, s)
}
