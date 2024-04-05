package vfs

import (
	"io"
	"io/ioutil"
)

type BaseFile struct {
	VFile
}

func (b *BaseFile) AsString() (s string, err error) {
	var bytes []byte
	bytes, err = ioutil.ReadAll(b)
	if err == nil {
		s = string(bytes)
	}
	return
}

func (b *BaseFile) AsBytes() ([]byte, error) {
	return ioutil.ReadAll(b)
}

func (b *BaseFile) WriteString(s string) (int, error) {
	return io.WriteString(b, s)
}
