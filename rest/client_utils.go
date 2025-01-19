package rest

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

// CreateMultipartHeader creates a multipart header with the given parameters
func CreateMultipartHeader(param, fileName, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)
	hdr.Set(ContentTypeHeader, "multipart/form-data")
	return hdr
}

// WriteMultipartFormFile writes a multipart form file to the writer
func WriteMultipartFormFile(w *multipart.Writer, fieldName, fileName string, r io.Reader) error {
	// Auto detect actual multipart content type
	cbuf := make([]byte, 512)
	size, err := r.Read(cbuf)
	if err != nil && err != io.EOF {
		return err
	}

	partWriter, err := w.CreatePart(CreateMultipartHeader(fieldName, fileName, http.DetectContentType(cbuf)))
	if err != nil {
		return err
	}

	if _, err = partWriter.Write(cbuf[:size]); err != nil {
		return err
	}

	_, err = io.Copy(partWriter, r)
	return err
}

// IsValidMultipartVerb checks if the method is valid for multipart content
func IsValidMultipartVerb(method string) (err error) {
	if !(method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) {
		err = fmt.Errorf("multipart content is now allowed on [%v]", method)
	}
	return
}
