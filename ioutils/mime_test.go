package ioutils

import (
	"testing"

	"go.nandlabs.io/golly/testing/assert"
)

func TestMimeToExt(t *testing.T) {
	tests := []struct {
		mime     string
		expected []string
	}{
		{MimeTextPlain, []string{".txt", ".text"}},
		{MimeTextHTML, []string{".html", ".htm"}},
		{MimeTextCSS, []string{".css"}},
		// Add more test cases here
	}

	for _, test := range tests {
		actual := mimeToExt[test.mime]
		assert.Equal(t, test.expected, actual)
	}
}

func TestExtToMime(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".txt", MimeTextPlain},
		{".text", MimeTextPlain},
		{".html", MimeTextHTML},
		// Add more test cases here
	}

	for _, test := range tests {
		actual := mapExtToMime[test.ext]
		assert.Equal(t, test.expected, actual)
	}
}
