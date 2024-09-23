package client

import (
	"fmt"
	"net/http"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
)

type Response struct {
	raw    *http.Response
	client *Client
}

// IsSuccess determines if the response is a success response
func (r *Response) IsSuccess() bool {
	return r.raw.StatusCode >= 200 && r.raw.StatusCode <= 204
}

// GetError gets the error with status code and value
func (r *Response) GetError() (err error) {
	if !r.IsSuccess() {
		err = fmt.Errorf("server responded with status code %d and status text %s",
			r.raw.StatusCode, r.raw.Status)
	}
	return
}

// Decode Function decodes the response body to a suitable object. The format of the body is determined by
// Content-Type header in the response
func (r *Response) Decode(v interface{}) (err error) {
	var c codec.Codec
	if r.IsSuccess() {
		defer ioutils.CloserFunc(r.raw.Body)
		contentType := r.raw.Header.Get(contentTypeHdr)
		c, err = codec.Get(contentType, r.client.codecOptions)
		if err == nil {
			err = c.Read(r.raw.Body, v)
		}
	} else {
		err = r.GetError()
	}
	return
}

// Status Provides status text of the http response
func (r *Response) Status() string {
	return r.Raw().Status
}

// StatusCode provides the status code of the response
func (r *Response) StatusCode() int {
	return r.Raw().StatusCode
}

// Raw Provides the backend raw response
func (r *Response) Raw() *http.Response {
	return r.raw
}
