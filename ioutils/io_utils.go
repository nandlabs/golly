package ioutils

import "io"

var CloserFunc = func(closer io.Closer) {
	_ = closer.Close()
}

func IsChanClosed[T any](ch chan T) (closed bool) {
	select {
	case <-ch:
		closed = true
	default:
		closed = false
	}
	return
}

func CloseChannel[T any](ch chan T) {
	if !IsChanClosed(ch) {
		close(ch)
	}
}
