package ws

import (
	"bytes"
	"testing"
)

// FuzzReadFrame asserts the RFC 6455 frame parser never panics on arbitrary
// byte input. Untrusted bytes must either parse to a valid frame or return
// an error — never crash.
func FuzzReadFrame(f *testing.F) {
	seeds := [][]byte{
		// Minimal text frame, FIN=1, payload="hi"
		{0x81, 0x02, 'h', 'i'},
		// Empty payload
		{0x81, 0x00},
		// Masked frame from a client
		{0x81, 0x82, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f},
		// 126-byte extended length
		{0x82, 0x7e, 0x00, 0x7e},
		// 65536-byte extended length (header only)
		{0x82, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
		// Close frame
		{0x88, 0x02, 0x03, 0xe8},
		// Ping
		{0x89, 0x00},
		// Pong
		{0x8a, 0x00},
		// Continuation
		{0x80, 0x01, 'x'},
		// Truncated (1 byte)
		{0x81},
		// Empty
		{},
		// Garbage
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		// 64KiB cap mirrors a reasonable server-side default; the value
		// limits memory the fuzzer can allocate per input.
		_, _ = readFrame(bytes.NewReader(data), 64*1024)
	})
}
