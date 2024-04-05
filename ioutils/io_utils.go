package ioutils

import "io"

var CloserFunc = func(closer io.Closer) {
	_ = closer.Close()
}
