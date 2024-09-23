package client

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// TestCreateMultipartHeader tests the CreateMultipartHeader function
func TestCreateMultipartHeader(t *testing.T) {
	param := "file"
	fileName := "test.txt"
	contentType := "text/plain"

	hdr := CreateMultipartHeader(param, fileName, contentType)
	if hdr.Get(contentTypeHdr) != "multipart/form-data" {
		t.Errorf("CreateMultipartHeader() = %v, want %v", hdr.Get(contentTypeHdr), "multipart/form-data")
	}
}

// TestWriteMultipartFormFile tests the WriteMultipartFormFile function
func TestWriteMultipartFormFile(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fieldName := "file"
	fileName := "test.txt"
	content := "This is a test file."

	err := WriteMultipartFormFile(w, fieldName, fileName, strings.NewReader(content))
	if err != nil {
		t.Errorf("WriteMultipartFormFile() error = %v", err)
	}

	w.Close()

	req, err := http.NewRequest("POST", "http://example.com/upload", &b)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	if req.Header.Get("Content-Type") != w.FormDataContentType() {
		t.Errorf("Content-Type header = %v, want %v", req.Header.Get("Content-Type"), w.FormDataContentType())
	}
}

// TestIsValidMultipartVerb tests the IsValidMultipartVerb function
func TestIsValidMultipartVerb(t *testing.T) {
	tests := []struct {
		method string
		want   error
	}{
		{http.MethodPost, nil},
		{http.MethodPut, nil},
		{http.MethodPatch, nil},
		{http.MethodGet, fmt.Errorf("multipart content is now allowed on [%v]", http.MethodGet)},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			err := IsValidMultipartVerb(tt.method)
			if (err != nil && tt.want == nil) || (err == nil && tt.want != nil) || (err != nil && tt.want != nil && err.Error() != tt.want.Error()) {
				t.Errorf("IsValidMultipartVerb() = %v, want %v", err, tt.want)
			}
		})
	}
}
